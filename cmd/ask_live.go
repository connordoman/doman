package cmd

import (
	"fmt"
	"log"

	tea "charm.land/bubbletea/v2"
	"doman.sh/doman/internal/data"
	"doman.sh/doman/internal/pkg/ask"
	"doman.sh/doman/internal/txt"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var AskLiveCommand = &cobra.Command{
	Use:   "live",
	Short: "An ongoing back and forth chat",
	Long: `Open a full screen chat with the configured AI service.

Messages are stored in the same database as 'doman ask', so a live session can
be resumed with --continue. Type /quick anywhere in a message to have that one
message answered quickly by the smaller model.`,
	RunE: runAskLiveCommand,
}

func init() {
	AskLiveCommand.Flags().StringP("model", "m", "", "Model to use for the AI service (default: gpt-4o-mini)")
	AskLiveCommand.Flags().String("quick-model", "", "Model used for /quick messages (default: gpt-5-mini)")
	AskLiveCommand.Flags().StringP("api-key", "A", "", "API Key for the AI service (default: read from environment variable OPENAI_API_KEY)")
	AskLiveCommand.Flags().BoolP("raw", "R", false, "Enable raw output (disable Markdown formatting)")
	AskLiveCommand.Flags().String("style", "", "Markdown render style: dark|light|auto (default: dark)")
	AskLiveCommand.Flags().BoolP("continue", "c", false, "Continue previous conversation")
	AskLiveCommand.Flags().String("id", "", "Specific conversation ID to continue (requires --continue)")
	AskLiveCommand.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	AskCommand.AddCommand(AskLiveCommand)
}

func runAskLiveCommand(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	raw, _ := cmd.Flags().GetBool("raw")

	if style, _ := cmd.Flags().GetString("style"); style != "" {
		viper.Set("ask.render_style", style)
	}

	apiKey, _ := cmd.Flags().GetString("api-key")
	if apiKey == "" {
		apiKey = viper.GetString("ask.openai.api_key")
		if apiKey == "" {
			return fmt.Errorf("API Key is required, please set it using --api-key or environment variable OPENAI_API_KEY")
		}
	}

	shouldContinue, _ := cmd.Flags().GetBool("continue")
	conversationID, _ := cmd.Flags().GetString("id")

	var conversation *data.Conversation
	var history []ask.LiveMessage

	if shouldContinue {
		conv, err := resolveLiveConversation(conversationID, verbose)
		if err != nil {
			return err
		}
		conversation = conv
		conversationID = conv.ID

		if history, err = loadLiveHistory(conversationID); err != nil {
			return err
		}

		if verbose {
			log.Printf("Continuing conversation %s with %d previous messages", conversationID, len(history))
		}
	} else {
		conversationID = uuid.New().String()
	}

	model, _ := cmd.Flags().GetString("model")
	if model == "" {
		if conversation != nil && conversation.Model != "" {
			model = conversation.Model
		} else {
			model = viper.GetString("ask.openai.default_model")
		}
		if model == "" {
			return fmt.Errorf("model is required, please set it using --model or configure it in the setup")
		}
	}

	quickModel, _ := cmd.Flags().GetString("quick-model")
	if quickModel == "" {
		quickModel = viper.GetString("ask.openai.quick_model")
		if quickModel == "" {
			quickModel = "gpt-5-mini"
		}
	}

	titleModel := viper.GetString("ask.title_model")
	if titleModel == "" {
		titleModel = "gpt-5-nano"
	}

	store := &liveChatStore{
		model:   model,
		service: "openai",
		exists:  conversation != nil,
	}

	cfg := ask.LiveChatConfig{
		ConversationID: conversationID,
		Model:          model,
		QuickModel:     quickModel,
		TitleModel:     titleModel,
		APIKey:         apiKey,
		Raw:            raw,
		NeedsTitle:     conversation == nil,
		History:        history,
		Store:          store,
	}

	if conversation != nil {
		cfg.Title = conversation.Title
	}

	final, err := tea.NewProgram(ask.NewLiveChatModel(cfg)).Run()
	if err != nil {
		return err
	}

	if !store.exists {
		return nil
	}

	printLiveSummary(final, conversationID)

	return nil
}

func resolveLiveConversation(conversationID string, verbose bool) (*data.Conversation, error) {
	if conversationID == "" {
		conversations, err := data.ListConversations(1, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to list conversations: %w", err)
		}
		if len(conversations) == 0 {
			return nil, fmt.Errorf("no previous conversation found, start a new conversation first")
		}
		return conversations[0], nil
	}

	conv, err := data.GetConversation(conversationID)
	if err == nil {
		return conv, nil
	}

	if verbose {
		log.Printf("conversation %s not found by exact id: %v", conversationID, err)
		log.Printf("attempting prefix lookup for %s", conversationID)
	}

	conv, err = data.FindConversationByPrefix(conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	return conv, nil
}

func loadLiveHistory(conversationID string) ([]ask.LiveMessage, error) {
	messages, err := data.GetMessagesByConversationID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to load conversation history: %w", err)
	}

	history := make([]ask.LiveMessage, 0, len(messages))
	for _, msg := range messages {
		// Older versions of 'ask' stored the rendered response, escape codes
		// and all, so strip them before the transcript re-renders the text.
		history = append(history, ask.LiveMessage{
			Role:    msg.Role,
			Content: ask.SanitizeStoredContent(msg.Content),
		})
	}

	return history, nil
}

func printLiveSummary(final tea.Model, conversationID string) {
	title := ""
	cost := 0.0
	if chat, ok := final.(ask.LiveChatModel); ok {
		title = chat.Title()
		cost = chat.TotalCost()
	}

	if !ask.IsMeaningfulTitle(title) {
		title = conversationID[:8]
	}

	summary := fmt.Sprintf("• %s", title)
	if cost > 0 {
		summary += fmt.Sprintf(" • $%.5f", cost)
	}
	summary += fmt.Sprintf(" • resume with: doman ask live -c --id %s", conversationID[:8])

	fmt.Printf("%s %s\n", txt.Bluef("ChatGPT"), txt.Greyf("%s", summary))
}

// liveChatStore adapts the data package to ask.LiveStore. The conversation row
// is created on the first saved message so abandoned sessions leave nothing
// behind.
type liveChatStore struct {
	model   string
	service string
	exists  bool
}

func (s *liveChatStore) SaveMessage(conversationID, role, content string) error {
	if err := s.ensureConversation(conversationID, role, content); err != nil {
		return err
	}

	if _, err := data.CreateMessage(conversationID, role, content); err != nil {
		return err
	}

	return nil
}

func (s *liveChatStore) TouchConversation(conversationID string) error {
	if !s.exists {
		return nil
	}

	return data.UpdateConversationTimestamp(conversationID)
}

func (s *liveChatStore) SaveTitle(conversationID, title string) error {
	if !s.exists {
		return nil
	}

	return data.UpdateConversationTitle(conversationID, title)
}

func (s *liveChatStore) ensureConversation(conversationID, role, content string) error {
	if s.exists {
		return nil
	}

	title := ""
	if role == "user" {
		title = ask.FallbackTitleFromPrompt(content)
	}

	if _, err := data.CreateConversation(conversationID, title, s.model, s.service); err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	s.exists = true

	return nil
}

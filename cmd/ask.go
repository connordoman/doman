package cmd

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/connordoman/doman/internal/config"
	"github.com/connordoman/doman/internal/data"
	"github.com/connordoman/doman/internal/pkg"
	"github.com/connordoman/doman/internal/pkg/timer"
	"github.com/connordoman/doman/internal/txt"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type AskSetup struct {
	Service string `yaml:"service"`
	Model   string `yaml:"model"`
	ApiKey  string `yaml:"api_key"`
}

var askSetup = &AskSetup{
	Service: "openai",
	Model:   "gpt-4o-mini",
	ApiKey:  "",
}

var base16Theme *huh.Theme = huh.ThemeBase16()

var setupForm = huh.NewForm(
	huh.NewGroup(
		huh.NewSelect[string]().
			Title("Select AI Service").
			Options(
				huh.NewOption("OpenAI", "openai"),
			).
			Value(&askSetup.Service)),
	huh.NewGroup(
		huh.NewInput().
			Title("Model for "+askSetup.Service).
			Value(&askSetup.Model)),
	huh.NewGroup(
		huh.NewInput().
			Title("API Key for "+askSetup.Service).
			Value(&askSetup.ApiKey),
	),
).WithTheme(base16Theme)

var AskCommand = &cobra.Command{
	Use:               "ask [prompt]",
	Short:             "Ask a question to the configured AI service",
	RunE:              runAsk,
	PersistentPreRunE: initAskDB,
	PersistentPostRun: closeAskDB,
}

func init() {
	AskCommand.Flags().BoolP("setup", "s", false, "Setup AI service configuration")
	AskCommand.Flags().StringP("model", "m", "", "Model to use for the AI service (default: gpt-4o-mini)")
	AskCommand.Flags().StringP("api-key", "A", "", "API Key for the AI service (default: read from environment variable OPENAI_API_KEY)")
	AskCommand.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	AskCommand.Flags().BoolP("raw", "R", false, "Enable raw output (disable Markdown formatting)")
	AskCommand.Flags().String("style", "", "Markdown render style: dark|light|auto (default: dark)")
	AskCommand.Flags().BoolP("continue", "c", false, "Continue previous conversation")
	AskCommand.Flags().String("id", "", "Specific conversation ID to continue (requires --continue)")

	AskCommand.AddCommand(AskConvosCommand)
}

func initAskDB(cmd *cobra.Command, args []string) error {
	dbPath := config.ConfigPath("ask.db")
	if err := data.InitDB(dbPath); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	return nil
}

func closeAskDB(cmd *cobra.Command, args []string) {
	data.CloseDB()
}

func runAsk(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	raw, _ := cmd.Flags().GetBool("raw")
	// Optional render style override for this invocation
	if style, _ := cmd.Flags().GetString("style"); style != "" {
		viper.Set("ask.render_style", style)
	}

	setup, err := cmd.Flags().GetBool("setup")
	if err != nil {
		return fmt.Errorf("failed to get setup flag: %w", err)
	}

	if setup {
		if err := runSetup(); err != nil {
			return fmt.Errorf("failed to run setup: %w", err)
		} else {
			return nil
		}
	}

	// Handle --continue flag
	shouldContinue, _ := cmd.Flags().GetBool("continue")
	conversationID, _ := cmd.Flags().GetString("id")

	var conversation *data.Conversation
	var history []pkg.MessageHistory

	if shouldContinue {
		if conversationID == "" {
			// Get the most recent conversation
			conversations, err := data.ListConversations(1, 0)
			if err != nil {
				return fmt.Errorf("failed to list conversations: %w", err)
			}
			if len(conversations) == 0 {
				return fmt.Errorf("no previous conversation found, start a new conversation first")
			}
			conversation = conversations[0]
			conversationID = conversation.ID
		} else {
			// Get specific conversation
			conv, err := data.GetConversation(conversationID)
			if err != nil {
				return fmt.Errorf("conversation not found: %w", err)
			}
			conversation = conv
		}

		// Load conversation history
		messages, err := data.GetMessagesByConversationID(conversationID)
		if err != nil {
			return fmt.Errorf("failed to load conversation history: %w", err)
		}

		for _, msg := range messages {
			history = append(history, pkg.MessageHistory{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		if verbose {
			log.Printf("Continuing conversation %s with %d previous messages", conversationID, len(history))
		}
	}

	// run normal ask command
	prompt := ""
	if len(args) > 0 {
		prompt = strings.TrimSpace(strings.Join(args, " "))
	} else {
		err := huh.NewInput().
			Title("Enter your question").
			Value(&prompt).
			Run()
		if err != nil {
			return fmt.Errorf("failed to get user input: %w", err)
		}

		prompt = strings.TrimSpace(prompt)

	}
	if prompt == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	promptStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)

	fmt.Println(promptStyle.Render(txt.Boldf("%s", txt.Bluef("You:")), prompt))

	apiKey, err := cmd.Flags().GetString("api-key")
	if err != nil {
		return fmt.Errorf("failed to get API Key flag: %w", err)
	}
	if apiKey == "" {
		apiKey = viper.GetString("ask.openai.api_key")
		if apiKey == "" {
			return fmt.Errorf("API Key is required, please set it using --api-key or environment variable OPENAI_API_KEY")
		}
	}

	model, err := cmd.Flags().GetString("model")
	if err != nil {
		return fmt.Errorf("failed to get model flag: %w", err)
	}
	if model == "" {
		if conversation != nil {
			model = conversation.Model
		} else {
			model = viper.GetString("ask.openai.default_model")
			if model == "" {
				return fmt.Errorf("model is required, please set it using --model or configure it in the setup")
			}
		}
	}

	// Create new conversation if not continuing
	if !shouldContinue {
		conversationID = uuid.New().String()
		service := "openai" // Default service
		_, err = data.CreateConversation(conversationID, model, service)
		if err != nil {
			return fmt.Errorf("failed to create conversation: %w", err)
		}
	} else {
		// Update conversation timestamp
		if err := data.UpdateConversationTimestamp(conversationID); err != nil {
			return fmt.Errorf("failed to update conversation timestamp: %w", err)
		}
	}

	askingMessage := pkg.AskSplashText[rand.Intn(len(pkg.AskSplashText))]

	spinnerPrompt := askingMessage + "..."
	if verbose {
		spinnerPrompt = fmt.Sprintf("%s %s...", askingMessage, txt.Boldf("%s", model))
	}

	timer := timer.NewStopwatch(true)

	var response string
	var pricing string
	if err := pkg.AskingSpinner(spinnerPrompt, func(ctx context.Context) error {
		completion, err := pkg.PromptAi(model, apiKey, prompt, history)
		if err != nil {
			return err
		}

		if response, err = pkg.CollectResponse(completion.Choices, raw); err != nil {
			return err
		}

		if cost, exists := pkg.CalculateCost(model, completion); exists {
			pricing = fmt.Sprintf(" \u2022 $%.5f", cost)
		}

		return nil
	}).Run(); err != nil {
		return err
	}

	timer.Stop()

	// Store user message
	if _, err := data.CreateMessage(conversationID, "user", prompt); err != nil {
		return fmt.Errorf("failed to store user message: %w", err)
	}

	// Store assistant response
	if _, err := data.CreateMessage(conversationID, "assistant", response); err != nil {
		return fmt.Errorf("failed to store assistant message: %w", err)
	}

	if response != "" {
		responseStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)

		fmt.Println()
		fmt.Println(responseStyle.Render(response))
		conversationInfo := txt.Greyf("\u2022 Check important info for mistakes.")
		if shouldContinue {
			idStyle := lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("#6b7280"))
			conversationInfo = fmt.Sprintf(" \u2022 %s \u2022 %s", idStyle.Render(conversationID[:8]), txt.Greyf("Check important info for mistakes."))
		}
		fmt.Printf("%s %s %s\n", txt.Bluef("ChatGPT"), txt.Greyf("\u2022 %s%s \u2022 %s", model, pricing, timer), conversationInfo)
	} else {
		fmt.Println(txt.Italicf("No response received"))
	}

	return nil
}

func runSetup() error {
	fmt.Printf("Configuring %s:\n", txt.Boldf("doman ask"))

	if err := setupForm.Run(); err != nil {
		return fmt.Errorf("failed to run setup form: %w", err)
	}

	if askSetup.ApiKey == "" {
		return fmt.Errorf("API Key is required")
	}

	viper.Set("ask.default_service", askSetup.Service)

	switch askSetup.Service {
	case "openai":
		if askSetup.Model == "" {
			return fmt.Errorf("model is required for OpenAI service")
		}
		viper.Set("ask.openai.default_model", askSetup.Model)
		viper.Set("ask.openai.api_key", askSetup.ApiKey)
	default:
		return fmt.Errorf("unsupported service: %s", askSetup.Service)
	}

	if err := config.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	pkg.PrintSuccess("Configuration saved successfully!")
	fmt.Printf("%s %s %s\n", txt.Greyf("You can now run"), txt.Boldf("doman ask"), txt.Greyf("to use your configuration."))

	return nil
}

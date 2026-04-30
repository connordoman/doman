package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/connordoman/doman/internal/config"
	"github.com/connordoman/doman/internal/data"
	"github.com/connordoman/doman/internal/pkg"
	"github.com/connordoman/doman/internal/pkg/ask"
	"github.com/connordoman/doman/internal/pkg/timer"
	"github.com/connordoman/doman/internal/txt"
	"github.com/connordoman/windy"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var askSetup = &ask.Setup{
	Service:    "openai",
	Model:      "gpt-4o-mini",
	QuickModel: "gpt-5-mini",
	ApiKey:     "",
}

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
	AskCommand.Flags().BoolP("dry-run", "d", false, "Do not generate a response (titles will still be generated)")
	AskCommand.Flags().BoolP("quick", "q", false, "Respond in a short format with a small model (gpt-5-mini)")

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
	quick, _ := cmd.Flags().GetBool("quick")
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
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	var conversation *data.Conversation
	var history []ask.MessageHistory

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
			// Get specific conversation (supports prefix)
			conv, err := data.GetConversation(conversationID)
			if err != nil {
				if verbose {
					log.Printf("conversation %s not found by exact id: %v", conversationID, err)
					log.Printf("attempting prefix lookup for %s", conversationID)
				}

				conv, err = data.FindConversationByPrefix(conversationID)
				if err != nil {
					return fmt.Errorf("conversation not found: %w", err)
				}
				conversationID = conv.ID
			}
			conversation = conv
		}

		// Load conversation history
		messages, err := data.GetMessagesByConversationID(conversationID)
		if err != nil {
			return fmt.Errorf("failed to load conversation history: %w", err)
		}

		for _, msg := range messages {
			history = append(history, ask.MessageHistory{
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
		inputTitle := "Enter your question:"
		if quick {
			inputTitle = "Enter your quick question:"
		}
		if shouldContinue {
			title := conversation.Title
			if !ask.IsMeaningfulTitle(title) {
				title = conversation.ID
			}
			inputTitle = fmt.Sprintf(`Follow up on "%s"`, title)
		}

		err := huh.NewText().
			Title(inputTitle).
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

	terminalWidth := pkg.DetectTerminalWidth()
	responseWrapWidth := max(terminalWidth-4, 20)

	fmt.Println(ask.FormatPrompt(prompt, quick))

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
		if quick {
			model = viper.GetString("ask.openai.quick_model")
			if model == "" {
				model = "gpt-5-mini"
			}
		} else if conversation != nil {
			model = conversation.Model
		} else {
			model = viper.GetString("ask.openai.default_model")
			if model == "" {
				return fmt.Errorf("model is required, please set it using --model or configure it in the setup")
			}
		}
	}

	// Create new conversation if not continuing
	var titleWaitGroup *sync.WaitGroup
	if !shouldContinue {
		conversationID = uuid.New().String()
		service := "openai" // Default service
		titleModel := viper.GetString("ask.title_model")
		if titleModel == "" {
			titleModel = "gpt-5-nano"
		}

		fallbackTitle := ask.FallbackTitleFromPrompt(prompt)
		conversation, err = data.CreateConversation(conversationID, fallbackTitle, model, service)
		if err != nil {
			return fmt.Errorf("failed to create conversation: %w", err)
		}

		titleWaitGroup = &sync.WaitGroup{}
		titleWaitGroup.Add(1)
		go func() {
			defer titleWaitGroup.Done()
			ctx := context.Background()
			var generatedTitle string

			if title, err := ask.GenerateShortTitle(ctx, apiKey, titleModel, prompt); err != nil {
				if verbose {
					log.Printf("failed to generate short title with %s: %v", titleModel, err)
				}

				if titleModel != model {
					if verbose {
						log.Printf("retrying title generation with conversation model %s", model)
					}
					if altTitle, altErr := ask.GenerateShortTitle(ctx, apiKey, model, prompt); altErr != nil {
						if verbose {
							log.Printf("fallback title generation failed: %v", altErr)
						}
					} else if ask.IsMeaningfulTitle(altTitle) {
						generatedTitle = altTitle
					}
				}
			} else if ask.IsMeaningfulTitle(title) {
				generatedTitle = title
			} else if verbose {
				log.Printf("generated title was empty or invalid; keeping fallback")
			}

			if generatedTitle != "" {
				if err := data.UpdateConversationTitle(conversationID, generatedTitle); err != nil {
					if verbose {
						log.Printf("failed to update conversation title: %v", err)
					}
				} else {
					// Update the conversation object with the new title
					conversation.Title = generatedTitle
					if verbose {
						log.Printf("updated conversation title to: %q (model=%s)", generatedTitle, titleModel)
					}
				}
			}
		}()
	} else {
		// Update conversation timestamp
		if err := data.UpdateConversationTimestamp(conversationID); err != nil {
			return fmt.Errorf("failed to update conversation timestamp: %w", err)
		}
	}

	askingMessage := ask.RandomSplashText()

	var spinnerPrompt strings.Builder
	if verbose {
		spinnerPrompt.WriteString(txt.Boldf("%s ", model))
	}
	spinnerPrompt.WriteString(askingMessage)
	spinnerPrompt.WriteString("...")

	timer := timer.NewStopwatch(true)

	var response string
	var pricing string

	if !dryRun {
		systemMessage := ask.DeveloperDefinedSystemMessage
		if quick {
			systemMessage = ask.QuickSystemMessage
		}

		if err := ask.AskingSpinner(spinnerPrompt.String(), quick, func(ctx context.Context) error {
			completion, err := ask.PromptAIWithSystemMessage(model, apiKey, prompt, history, systemMessage)
			if err != nil {
				return err
			}

			if response, err = ask.CollectResponse(completion.Choices, raw, responseWrapWidth); err != nil {
				return err
			}

			if cost, exists := ask.CalculateCost(model, completion); exists {
				pricing = fmt.Sprintf(" \u2022 $%.5f", cost)
			}

			return nil
		}).Run(); err != nil {
			return err
		}

	} else {
		response = "\n<dry-run> No response was generated.\n"
	}

	timer.Stop()

	if !dryRun {
		// Store user message
		if _, err := data.CreateMessage(conversationID, "user", prompt); err != nil {
			return fmt.Errorf("failed to store user message: %w", err)
		}

		// Store assistant response
		if _, err := data.CreateMessage(conversationID, "assistant", response); err != nil {
			return fmt.Errorf("failed to store assistant message: %w", err)
		}
	}

	if response != "" || dryRun {
		if dryRun {
			fmt.Println(lipgloss.NewStyle().Italic(dryRun).Render(response))
		} else {
			fmt.Println(response)
		}

		conversationInfo := txt.Greyf("\u2022 Check important info for mistakes.")

		// Wait for title generation to complete if it's a new conversation
		if titleWaitGroup != nil {
			ask.AskingSpinner("", quick, func(ctx context.Context) error {
				titleWaitGroup.Wait()
				return nil
			}).Run()
		}

		conversationTitle := ""
		if conversation != nil {
			conversationTitle = conversation.Title
		}
		if conversationTitle == "" || !ask.IsMeaningfulTitle(conversationTitle) {
			conversationTitle = conversationID[:8]
		}

		idStyle := lipgloss.NewStyle().Underline(true).Foreground(windy.Neutral500.Glossy())

		fmt.Printf("%s %s %s\n", txt.Bluef("ChatGPT"), txt.Greyf("\u2022 %s%s \u2022 %s \u2022 %s", model, pricing, timer, idStyle.Render(conversationTitle)), conversationInfo)
	} else {
		fmt.Println(txt.Italicf("No response received"))
	}

	return nil
}

func runSetup() error {
	return ask.RunSetup(askSetup)
}

package ask

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/connordoman/doman/internal/config"
	"github.com/connordoman/doman/internal/pkg"
	"github.com/connordoman/doman/internal/txt"
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
)

var AskCommand = &cobra.Command{
	Use:   "ask [prompt]",
	Short: "Ask a question to the configured AI service",
	RunE:  runAsk,
}

func init() {
	AskCommand.Flags().BoolP("setup", "s", false, "Setup AI service configuration")
	AskCommand.Flags().StringP("model", "m", "", "Model to use for the AI service (default: gpt-4o-mini)")
	AskCommand.Flags().StringP("api-key", "A", "", "API Key for the AI service (default: read from environment variable OPENAI_API_KEY)")
	AskCommand.Flags().BoolP("verbose", "v", false, "Enable verbose output")
}

func runAsk(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

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
		if prompt == "" {
			return fmt.Errorf("prompt cannot be empty")
		}
	}

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
		model = viper.GetString("ask.openai.default_model")
		if model == "" {
			return fmt.Errorf("model is required, please set it using --model or configure it in the setup")
		}
	}

	spinnerPrompt := "Asking AI..."
	if verbose {
		spinnerPrompt = fmt.Sprintf("Asking AI with model %s...", txt.Boldf("%s", model))
	}

	var response string
	if err := spinner.New().Title(spinnerPrompt).ActionWithErr(func(ctx context.Context) error {
		response, err = pkg.PromptAi(model, apiKey, prompt)
		if err != nil {
			pkg.PrintError("Failed to get AI response: %v", err)
			return err
		}
		return nil
	}).Run(); err != nil {
		return err
	}

	if response != "" {
		fmt.Println(response)
		fmt.Printf("%s %s %s\n", txt.Bluef("ChatGPT"), txt.Greyf("\u2022 %s", model), txt.Greyf("\u2022 Check important info for mistakes."))
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
	fmt.Printf("%s %s %s\n", txt.Greyf("You can now run"), txt.Boldf("%s", "doman ask"), txt.Greyf("to use your configuration."))

	return nil
}

package ask

import (
	"fmt"

	"doman.sh/internal/config"
	"doman.sh/internal/pkg"
	"doman.sh/internal/txt"
	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

type Setup struct {
	Service    string `yaml:"service"`
	Model      string `yaml:"model"`
	QuickModel string `yaml:"quick_model"`
	ApiKey     string `yaml:"api_key"`
}

var base16Theme *huh.Theme = huh.ThemeBase16()

func NewSetupForm(setup *Setup) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select AI Service").
				Options(
					huh.NewOption("OpenAI", "openai"),
				).
				Value(&setup.Service)),
		huh.NewGroup(
			huh.NewInput().
				Title("Model for "+setup.Service).
				Value(&setup.Model)),
		huh.NewGroup(
			huh.NewInput().
				Title("Quick model for "+setup.Service).
				Value(&setup.QuickModel)),
		huh.NewGroup(
			huh.NewInput().
				Title("API Key for "+setup.Service).
				Value(&setup.ApiKey),
		),
	).WithTheme(base16Theme)
}

func RunSetup(setup *Setup) error {
	fmt.Printf("Configuring %s:\n", txt.Boldf("doman ask"))

	form := NewSetupForm(setup)
	if err := form.Run(); err != nil {
		return fmt.Errorf("failed to run setup form: %w", err)
	}

	if setup.ApiKey == "" {
		return fmt.Errorf("API Key is required")
	}

	viper.Set("ask.default_service", setup.Service)

	switch setup.Service {
	case "openai":
		if setup.Model == "" {
			return fmt.Errorf("model is required for OpenAI service")
		}
		if setup.QuickModel == "" {
			return fmt.Errorf("quick model is required for OpenAI service")
		}
		viper.Set("ask.openai.default_model", setup.Model)
		viper.Set("ask.openai.quick_model", setup.QuickModel)
		viper.Set("ask.openai.api_key", setup.ApiKey)
	default:
		return fmt.Errorf("unsupported service: %s", setup.Service)
	}

	if err := config.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	pkg.PrintSuccess("Configuration saved successfully!")
	fmt.Printf("%s %s %s\n", txt.Greyf("You can now run"), txt.Boldf("doman ask"), txt.Greyf("to use your configuration."))

	return nil
}

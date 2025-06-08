package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/padam-meesho/NotificationService/internal/models"
	"github.com/spf13/viper"
)

func LoadAppConfig(cfg *models.AppConfig) (*models.AppConfig, error) {
	v := viper.New()

	// Set config file name and path
	v.SetConfigName("app_config")
	v.SetConfigType("yaml")
	v.AddConfigPath("configs")

	// Read YAML config
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Read .env file (if present)
	envPath := filepath.Join("cmd", ".env") // place the .env in the cmd folder
	if _, err := os.Stat(envPath); err == nil {
		v.SetConfigFile(envPath)
		_ = v.MergeInConfig() // Merge .env values (ignore error if not present)
	}

	// Optionally bind ENV variables (for runtime overrides)
	v.AutomaticEnv()

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return cfg, nil
}

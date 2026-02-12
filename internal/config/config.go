package config

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/stackgen-cli/dataclean/internal/models"
)

// Load loads configuration from file or returns defaults with auto-detection
func Load(cfgFile string) (*models.Config, error) {
	cfg := models.DefaultConfig()

	// Try to load from specified file
	if cfgFile != "" {
		data, err := os.ReadFile(cfgFile)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	// Try default config file
	defaultFiles := []string{".dataclean.yaml", ".dataclean.yml"}
	for _, file := range defaultFiles {
		if data, err := os.ReadFile(file); err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
			return cfg, nil
		}
	}

	// Return default config (auto-detection mode)
	return cfg, nil
}

// Save writes configuration to a file
func Save(cfg *models.Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

package models

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadAppConfig loads the application configuration from a YAML file.
func LoadAppConfig(path string) (AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, fmt.Errorf("failed to read config: %w", err)
	}

	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return AppConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

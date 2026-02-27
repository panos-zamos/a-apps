package models

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ChangelogEntry describes a single changelog entry.
type ChangelogEntry struct {
	Version string   `yaml:"version"`
	Date    string   `yaml:"date"`
	Changes []string `yaml:"changes"`
}

// LoadChangelog loads changelog entries from a YAML file.
func LoadChangelog(path string) ([]ChangelogEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read changelog: %w", err)
	}

	var entries []ChangelogEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse changelog: %w", err)
	}

	return entries, nil
}

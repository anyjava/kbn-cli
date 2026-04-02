package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Fields struct {
	ID       string `yaml:"id"`
	Title    string `yaml:"title"`
	Status   string `yaml:"status"`
	Priority string `yaml:"priority"`
	Type     string `yaml:"type"`
}

type Config struct {
	Vault          string   `yaml:"vault"`
	Path           string   `yaml:"path"`
	Glob           string   `yaml:"glob"`
	Fields         Fields   `yaml:"fields"`
	HiddenStatuses []string  `yaml:"hidden_statuses"`
	ColumnOrder    []string  `yaml:"column_order"`
}

func (c *Config) FullPath() string {
	return filepath.Join(c.Vault, c.Path)
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Glob == "" {
		cfg.Glob = "*.md"
	}

	if cfg.Fields.Status == "" {
		return nil, fmt.Errorf("fields.status is required in config")
	}

	return cfg, nil
}

func Load(overridePath string) (*Config, error) {
	if overridePath != "" {
		return LoadFromFile(overridePath)
	}

	if _, err := os.Stat(".kbn.yml"); err == nil {
		return LoadFromFile(".kbn.yml")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home dir: %w", err)
	}
	globalPath := filepath.Join(home, ".config", "kbn", "config.yml")
	if _, err := os.Stat(globalPath); err == nil {
		return LoadFromFile(globalPath)
	}

	return nil, fmt.Errorf("no config file found: create .kbn.yml or ~/.config/kbn/config.yml")
}

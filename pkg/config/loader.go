package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App            AppConfig           `yaml:"app"`
	ActiveProvider string              `yaml:"active_provider"`
	Providers      map[string]Provider `yaml:"providers"`
}

type AppConfig struct {
	Timeout int `yaml:"timeout"`
}

type Provider struct {
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
	BaseURL string `yaml:"base_url"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.ActiveProvider == "" {
		return errors.New("active_provider is required in config.yaml")
	}

	provider, ok := c.Providers[c.ActiveProvider]
	if !ok {
		return fmt.Errorf("provider '%s' not found in providers list", c.ActiveProvider)
	}

	if provider.APIKey == "" && c.ActiveProvider != "ollama" {
		return fmt.Errorf("api_key is required for provider '%s'", c.ActiveProvider)
	}

	return nil
}

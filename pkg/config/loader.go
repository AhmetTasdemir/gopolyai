package config

import (
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

	return &cfg, nil
}

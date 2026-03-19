package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type LocalConfig struct {
	Server    ServerConfig              `yaml:"server"`
	Providers map[string]ProviderConfig `yaml:"providers"`
	Routing   map[string]RouteConfig    `yaml:"routing"`
	Limits    LimitsConfig              `yaml:"limits"`
	Tracking  TrackingConfig            `yaml:"tracking"`
	Keys      map[string]KeyConfig      `yaml:"keys"`
}

type ServerConfig struct {
	Port    string        `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type ProviderConfig struct {
	Enabled bool          `yaml:"enabled"`
	APIKey  string        `yaml:"api_key"`
	BaseURL string        `yaml:"base_url"`
	Models  []ModelConfig `yaml:"models"`
}

type ModelConfig struct {
	Name       string  `yaml:"name"`
	InputCost  float64 `yaml:"input_cost"`
	OutputCost float64 `yaml:"output_cost"`
	RPM        int     `yaml:"rpm"`
	TPM        int     `yaml:"tpm"`
}

type RouteConfig struct {
	Strategy string        `yaml:"strategy"`
	Order    []RouteTarget `yaml:"order,omitempty"`
	Pool     []RouteTarget `yaml:"pool,omitempty"`
}

type RouteTarget struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
}

type LimitsConfig struct {
	Global      RateLimitConfig            `yaml:"global"`
	PerProvider map[string]RateLimitConfig `yaml:"per_provider"`
}

type RateLimitConfig struct {
	RPM int `yaml:"rpm"`
	TPM int `yaml:"tpm"`
}

type KeyConfig struct {
	RPM int `yaml:"rpm"`
}

type TrackingConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Currency string `yaml:"currency"`
	LogUsage bool   `yaml:"log_usage"`
}

func Load(path string) (*LocalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config.yaml: %w", err)
	}

	var cfg LocalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	for name, provider := range cfg.Providers {
		if provider.APIKey == "" {
			continue
		}

		key := provider.APIKey
		if strings.HasPrefix(key, "${") && strings.HasSuffix(key, "}") {
			envName := strings.TrimSuffix(strings.TrimPrefix(key, "${"), "}")
			envValue := os.Getenv(envName)

			if envValue == "" {
				return nil, fmt.Errorf("env var %s not set for provider %s", envName, name)
			}

			provider.APIKey = envValue
			cfg.Providers[name] = provider
		}
	}

	return &cfg, nil
}

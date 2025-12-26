package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	LLM LLMConfig `mapstructure:"llm"`
}

type LLMConfig struct {
	Provider string `mapstructure:"provider"`
	BaseURL  string `mapstructure:"base_url"`
	APIKey   string `mapstructure:"api_key"`
	Model    string `mapstructure:"model"`
}

func Load() (Config, error) {
	v := viper.New()
	v.SetConfigName(".cgpd")
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("get working directory: %w", err)
	}

	dir := cwd
	for {
		v.AddConfigPath(dir)
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	_ = v.BindEnv("llm.provider", "CGPD_LLM_PROVIDER", "LLM_PROVIDER")
	_ = v.BindEnv("llm.base_url", "CGPD_LLM_BASE_URL", "LLM_BASE_URL", "OPENAI_BASE_URL")
	_ = v.BindEnv("llm.model", "CGPD_LLM_MODEL", "LLM_MODEL", "OPENAI_MODEL")
	_ = v.BindEnv("llm.api_key", "CGPD_LLM_API_KEY", "OPENAI_API_KEY", "LLM_API_KEY")

	readErr := v.ReadInConfig()
	var notFound viper.ConfigFileNotFoundError
	if readErr != nil && !errors.As(readErr, &notFound) {
		return Config{}, fmt.Errorf("read config: %w", readErr)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if readErr != nil {
		if strings.TrimSpace(cfg.LLM.Provider) == "" &&
			strings.TrimSpace(cfg.LLM.APIKey) == "" &&
			strings.TrimSpace(cfg.LLM.Model) == "" {
			return Config{}, errors.New("config file .cgpd.yaml not found and no LLM env vars set (try CGPD_LLM_PROVIDER, CGPD_LLM_MODEL, OPENAI_API_KEY)")
		}
	}

	return cfg, nil
}

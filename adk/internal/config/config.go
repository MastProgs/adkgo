package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

// AI-NOTE: 전역 환경설정을 파싱하는 구조체입니다.
type Config struct {
	Models   map[string]ModelConfig `toml:"models"` // AI-NOTE: 애플리케이션에 띄어둘 LLM Provider만 관리
}

type ModelConfig struct {
	Provider  string   `toml:"provider"`
	ModelName string   `toml:"model_name"`
	APIKeys   []string `toml:"api_keys"`
}

func Load(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("config.toml 파싱 실패: %w", err)
	}
	return &cfg, nil
}

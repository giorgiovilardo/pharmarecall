package config

import (
	"fmt"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Server    ServerConfig    `koanf:"server"`
	DB        DBConfig        `koanf:"db"`
	Session   SessionConfig   `koanf:"session"`
	Lookahead LookaheadConfig `koanf:"lookahead"`
}

type ServerConfig struct {
	Port int `koanf:"port"`
}

type DBConfig struct {
	URL string `koanf:"url"`
}

type SessionConfig struct {
	Secret string `koanf:"secret"`
}

type LookaheadConfig struct {
	Days int `koanf:"days"`
}

func Load(path string) (Config, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(path), toml.Parser()); err != nil {
		return Config{}, fmt.Errorf("loading config from %s: %w", path, err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshalling config: %w", err)
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Lookahead.Days == 0 {
		cfg.Lookahead.Days = 7
	}

	return cfg, nil
}

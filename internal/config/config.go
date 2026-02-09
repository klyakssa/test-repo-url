package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

type FileStorage struct {
	Path string `env:"FILE_STORAGE_PATH" envDefault:""`
}

type WebConfig struct {
	HostPort string `env:"SERVER_ADDRESS" envDefault:""`
	BaseUrl  string `env:"BASE_URL" envDefault:""`
}

type Config struct {
	WebConfig WebConfig
	File      FileStorage
}

func InitFlagConfig() *Config {
	cfg := Config{}

	env.Parse(&cfg.WebConfig)
	env.Parse(&cfg.File)

	flagHostPort := pflag.StringP("server", "a", "localhost:8080", "server host")
	flagBaseUrl := pflag.StringP("base", "b", "http://localhost:8080", "base url")
	flagFilePath := pflag.StringP("file", "f", "./storage.json", "file storage path")

	pflag.Parse()

	if cfg.WebConfig.HostPort == "" {
		cfg.WebConfig.HostPort = *flagHostPort
	}

	if cfg.WebConfig.BaseUrl == "" {
		cfg.WebConfig.BaseUrl = *flagBaseUrl
	}

	if cfg.File.Path == "" {
		cfg.File.Path = *flagFilePath
	}

	return &cfg
}

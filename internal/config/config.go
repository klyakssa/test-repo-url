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

type DBConfig struct {
	ConnString string `env:"DATABASE_DSN" envDefault:""`
}

type Config struct {
	WebConfig WebConfig
	File      FileStorage
	PostDB    DBConfig
}

func InitFlagConfig() *Config {
	cfg := Config{}

	env.Parse(&cfg.WebConfig)
	env.Parse(&cfg.File)
	env.Parse(&cfg.PostDB)

	flagHostPort := pflag.StringP("server", "a", "localhost:8080", "server host")
	flagBaseUrl := pflag.StringP("base", "b", "http://localhost:8080", "base url")
	flagFilePath := pflag.StringP("file", "f", "./storage.json", "file storage path")
	flagConnString := pflag.StringP("postgresdb", "d", "postgres://postgres:postgres@localhost:9999/postgres?sslmode=disable", "database connection string")

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

	if cfg.PostDB.ConnString == "" {
		cfg.PostDB.ConnString = *flagConnString
	}

	return &cfg
}

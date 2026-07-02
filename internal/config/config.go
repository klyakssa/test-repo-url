package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

type FileStorage struct {
	Path string `env:"FILE_STORAGE_PATH" envDefault:""`
}

type WebConfig struct {
	HostPort      string `env:"SERVER_ADDRESS" envDefault:""`
	BaseURL       string `env:"BASE_URL" envDefault:""`
	Secret        string `env:"SECRET" envDefault:""`
	EnableHTTPS   bool   `env:"ENABLE_HTTPS"`
	CertFile      string `env:"CERT_FILE" envDefault:"server.crt"`
	KeyFile       string `env:"KEY_FILE" envDefault:"server.key"`
	TrustedSubnet string `env:"TRUSTED_SUBNET" envDefault:""`
}

type DBConfig struct {
	ConnString string `env:"DATABASE_DSN" envDefault:""`
}

type AuditConfig struct {
	AuditFile string `env:"AUDIT_FILE" envDefault:""`
	AuditURL  string `env:"AUDIT_URL" envDefault:""`
}

type Config struct {
	WebConfig WebConfig
	File      FileStorage
	Audit     AuditConfig
	PostDB    DBConfig
	PathFile  string
}

type JSONConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
	Secret          string `json:"secret"`
	AuditFile       string `json:"audit_file"`
	AuditURL        string `json:"audit_url"`
	TrustedSubnet   string `json:"trusted_subnet"`
}

func getConfigFilePath() (configFlag string) {

	configFlag = os.Getenv("CONFIG")

	flagConfig := pflag.StringP("config", "c", "", "path to config file")
	pflag.Parse()

	if configFlag == "" {
		configFlag = *flagConfig
	}

	return configFlag
}

func loadJSONConfig(filePath string) (*JSONConfig, error) {
	if filePath == "" {
		return nil, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	var jsonCfg JSONConfig
	if err := json.Unmarshal(data, &jsonCfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filePath, err)
	}

	return &jsonCfg, nil
}

func applyJSONConfig(cfg *Config, jsonCfg *JSONConfig) {
	if jsonCfg == nil {
		return
	}

	if jsonCfg.ServerAddress != "" && cfg.WebConfig.HostPort == "" {
		cfg.WebConfig.HostPort = jsonCfg.ServerAddress
	}

	if jsonCfg.BaseURL != "" && cfg.WebConfig.BaseURL == "" {
		cfg.WebConfig.BaseURL = jsonCfg.BaseURL
	}

	if jsonCfg.FileStoragePath != "" && cfg.File.Path == "" {
		cfg.File.Path = jsonCfg.FileStoragePath
	}

	if jsonCfg.DatabaseDSN != "" && cfg.PostDB.ConnString == "" {
		cfg.PostDB.ConnString = jsonCfg.DatabaseDSN
	}

	if jsonCfg.Secret != "" && cfg.WebConfig.Secret == "" {
		cfg.WebConfig.Secret = jsonCfg.Secret
	}

	if jsonCfg.AuditFile != "" && cfg.Audit.AuditFile == "" {
		cfg.Audit.AuditFile = jsonCfg.AuditFile
	}

	if jsonCfg.AuditURL != "" && cfg.Audit.AuditURL == "" {
		cfg.Audit.AuditURL = jsonCfg.AuditURL
	}

	if !cfg.WebConfig.EnableHTTPS && jsonCfg.EnableHTTPS {
		cfg.WebConfig.EnableHTTPS = jsonCfg.EnableHTTPS
	}

	if jsonCfg.TrustedSubnet != "" && cfg.WebConfig.TrustedSubnet == "" {
		cfg.WebConfig.TrustedSubnet = jsonCfg.TrustedSubnet
	}
}

func InitFlagConfig() *Config {
	cfg := Config{}

	configPath := getConfigFilePath()
	jsonCfg, err := loadJSONConfig(configPath)
	if err != nil {
		fmt.Printf("Warning: failed to load config file: %v\n", err)
	}
	applyJSONConfig(&cfg, jsonCfg)

	env.Parse(&cfg.WebConfig)
	env.Parse(&cfg.File)
	env.Parse(&cfg.PostDB)
	env.Parse(&cfg.Audit)

	flagHostPort := pflag.StringP("server", "a", "localhost:8080", "server host")
	flagBaseURL := pflag.StringP("base", "b", "http://localhost:8080", "base url")
	flagFilePath := pflag.StringP("file", "f", "./storage.json", "file storage path")
	flagConnString := pflag.StringP("postgresdb", "d", "", "database connection string") //-d=postgres://postgres:11@localhost:5432/test_prac?sslmode=disable -d=postgres://test:11@localhost:5432/prac?sslmode=disable
	flagSecret := pflag.StringP("secret", "k", "GASGIOPHFAISGFAHBKWAYFGS", "secret key")
	flagAuditFile := pflag.StringP("audit-file", "l", "", "audit log file path")
	flagAuditURL := pflag.StringP("audit-url", "u", "", "audit log server url")
	flagEnableHTTPS := pflag.BoolP("enable-https", "s", false, "enable https")
	flagTrustedSubnet := pflag.StringP("trusted-subnet", "t", "", "trusted subnet")

	pflag.Parse()

	if cfg.WebConfig.HostPort == "" {
		cfg.WebConfig.HostPort = *flagHostPort
	}

	if cfg.WebConfig.BaseURL == "" {
		cfg.WebConfig.BaseURL = *flagBaseURL
	}

	if cfg.File.Path == "" {
		cfg.File.Path = *flagFilePath
	}

	if cfg.PostDB.ConnString == "" {
		cfg.PostDB.ConnString = *flagConnString
	}

	if cfg.WebConfig.Secret == "" {
		cfg.WebConfig.Secret = *flagSecret
	}

	if cfg.Audit.AuditFile == "" {
		cfg.Audit.AuditFile = *flagAuditFile
	}

	if cfg.Audit.AuditURL == "" {
		cfg.Audit.AuditURL = *flagAuditURL
	}

	if !cfg.WebConfig.EnableHTTPS {
		cfg.WebConfig.EnableHTTPS = *flagEnableHTTPS
	}

	if cfg.WebConfig.TrustedSubnet == "" {
		cfg.WebConfig.TrustedSubnet = *flagTrustedSubnet
	}

	return &cfg
}

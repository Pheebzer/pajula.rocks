package config

import (
	"fmt"
	"os"

	logger "pajula.rocks/internal/log"

	"gopkg.in/yaml.v2"
	"pajula.rocks/internal/utils"
)

type Config struct {
	TokenEndpoint    string `yaml:"tokenEndpoint"`
	UserEndpoint     string `yaml:"userEndpoint"`
	PlaylistEndpoint string `yaml:"playlistEndpoint"`
	PlaylistId       string `yaml:"playlistId"`
	ApiKey           string `yaml:"apiKey"`
	MysqlDsn         string `yaml:"mysqlDsn"`
}

func GetConfigs() Config {
	var cfg Config
	fp := os.Getenv("CONFIG_FILE")
	if fp == "" {
		fp = fmt.Sprintf("%s/config.yaml", utils.ProjectRoot)
		logger.Warn("Environment variable 'CONFIG_FILE' not set")
		logger.Infof("using default file: %s", fp)
	}
	f, err := os.Open(fp)
	if err != nil {
		logger.Warn("Unable to load config file or not found, using env variables only")
	}
	defer f.Close()
	parseConfig(f, &cfg)
	return cfg
}

func parseConfig(f *os.File, cfg *Config) {
	decoder := yaml.NewDecoder(f)
	err := decoder.Decode(cfg)
	if err != nil {
		logger.Fatalf("Cannot parse configuration:\n%s", err)
	}
	// env variables override if they exist
	parseEnv(cfg)
}

func parseEnv(cfg *Config) {
	cfg.TokenEndpoint = getEnv("SPOTIFY_TOKEN_ENDPOINT", cfg.TokenEndpoint)
	cfg.UserEndpoint = getEnv("SPOTIFY_USER_ENDPOINT", cfg.UserEndpoint)
	cfg.PlaylistEndpoint = getEnv("SPOTIFY_PLAYLIST_ENDPOINT", cfg.PlaylistEndpoint)
	cfg.PlaylistId = getEnv("SPOTIFY_PLAYLIST_ID", cfg.PlaylistId)
	cfg.ApiKey = getEnv("SPOTIFY_API_KEY", cfg.ApiKey)
	cfg.MysqlDsn = getEnv("MYSQL_DSN", cfg.MysqlDsn)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	} else if fallback != "" {
		return fallback
	}
	logger.Fatalf("Missing config value '%s'", key)
	return "" // dummy return
}

package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	TokenEndpoint    string `yaml:"tokenEndpoint"`
	UserEndpoint     string `yaml:"userEndpoint"`
	PlaylistEndpoint string `yaml:"playlistEndpoint"`
	PlaylistId       string `yaml:"playlistId"`
	ApiKey           string `yaml:"apiKey"`
	MysqlDsn         string `yaml:"mysqlDsn"`
}

func parseConfig(f *os.File, cfg *Config) {
	decoder := yaml.NewDecoder(f)
	err := decoder.Decode(cfg)
	if err != nil {
		log.Fatalf("ERROR: Cannot parse configuration: \n%s", err)
	}
	// env variables override if they exist
	parseEnv(cfg)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	} else if fallback != "" {
		return fallback
	}
	log.Fatalf("ERROR: Missing config value '%s' - check config type for missing value", key)
	return "" // dummy return
}

func parseEnv(cfg *Config) {
	cfg.TokenEndpoint = getEnv("SPOTIFY_TOKEN_ENDPOINT", cfg.TokenEndpoint)
	cfg.UserEndpoint = getEnv("SPOTIFY_USER_ENDPOINT", cfg.UserEndpoint)
	cfg.PlaylistEndpoint = getEnv("SPOTIFY_PLAYLIST_ENDPOINT", cfg.PlaylistEndpoint)
	cfg.PlaylistId = getEnv("SPOTIFY_PLAYLIST_ID", cfg.PlaylistId)
	cfg.ApiKey = getEnv("SPOTIFY_API_KEY", cfg.ApiKey)
	cfg.MysqlDsn = getEnv("MYSQL_DSN", cfg.MysqlDsn)
}

func getConfigs() Config {
	var cfg Config
	fp := os.Getenv("CONFIG_FILE")
	if fp == "" {
		log.Println("INFO: Environment variable 'CONFIG_FILE' not set, using default file")
		fp = "config.yaml"
	}
	f, err := os.Open(fp)
	if err != nil {
		log.Println("WARN: Unable to load config file or not found, using env variables only")
	}
	defer f.Close()
	parseConfig(f, &cfg)
	return cfg
}

func initLogs() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime)
}

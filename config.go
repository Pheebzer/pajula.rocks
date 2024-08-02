package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Spotify struct {
		TokenEndpoint    string `yaml:"tokenEndpoint"`
		UserEndpoint     string `yaml:"userEndpoint"`
		PlaylistEndpoint string `yaml:"playlistEndpoint"`
		PlaylistId       string `yaml:"playlistId"`
		ApiKey           string `yaml:"apiKey"`
	} `yaml:"spotify"`
	Db struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"db"`
}

func parseConfig(filepath string) *Config {

	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		panic(err)
	}
	return &cfg
}

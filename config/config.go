package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Player struct {
	AccountName string `yaml:"accountName"`
	SteamID     string `yaml:"steamId"`
	Track       bool   `yaml:"track"`
}

type AppConfig struct {
	SteamAPIKey   string   `yaml:"steam_api_key"`
	SteamAPIURL   string   `yaml:"steam_api_url"`
	LeetifyAPIURL string   `yaml:"leetify_api_url"`
	DiscordHook   string   `yaml:"discord_hook"`
	Players       []Player `yaml:"players"`
}

func MustLoadConfig(filename string) *AppConfig {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Config: Error reading config file: %v", err)
	}

	var config AppConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Config: Error parsing config file: %v", err)
	}

	if len(config.Players) == 0 {
		log.Fatal("Config: No players configured")
	}

	return &config
}

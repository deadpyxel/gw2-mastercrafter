package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	ApiKey string `json:"api_key"`
}

func ReadConfig() Config {
	file, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Unable to read config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("Unable to parse config file: %v", err)
	}

	return config
}

package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	ApiKey          string   `json:"api_key"`
	ProfitThreshold float64  `json:"profit_threshold"`
	LogLevel        string   `json:"log_level"`
	RemovedTypes    []string `json:"removed_types"`
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

	if config.LogLevel == "" {
		config.LogLevel = "INFO"
	}

	return config
}

package main

import (
	"os"

	"github.com/deadpyxel/gw2-mastercrafter/internal"
)

func main() {
	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		config := config.ReadConfig()
		apiToken = config.ApiKey
	}
	gw2Client := NewAPIClient("https://api.guildwars2.com/v2", apiToken)
	crafter := NewCrafter(*gw2Client)
	crafter.FindProfitableOptions(19700)
}

package main

import (
	"os"

	"github.com/deadpyxel/gw2-mastercrafter/internal/config"
	"github.com/deadpyxel/gw2-mastercrafter/pkg/api"
	"github.com/deadpyxel/gw2-mastercrafter/pkg/crafting"
)

func main() {
	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		config := config.ReadConfig()
		apiToken = config.ApiKey
	}
	gw2Client := api.NewAPIClient("https://api.guildwars2.com/v2", apiToken)
	crafter := crafting.NewCrafter(*gw2Client)
	crafter.FindProfitableOptions(19700)
}

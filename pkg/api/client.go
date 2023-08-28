package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/deadpyxel/gw2-mastercrafter/pkg/types"
)

type APIClient struct {
	baseURL   string
	authToken string
}

func NewAPIClient(baseURL, authToken string) *APIClient {
	return &APIClient{baseURL: baseURL, authToken: authToken}
}

func (client *APIClient) makeRequest(endpoint string) (*http.Response, error) {
	url := client.baseURL + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+client.authToken)
	httpClient := &http.Client{}
	return httpClient.Do(req)
}

func (client *APIClient) FetchAvailableRecipesIds(itemID int) (types.RecipeIds, error) {
	endpoint := fmt.Sprintf("/recipes/search?input=%d", itemID)
	response, err := client.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var recipeIds types.RecipeIds
	err = json.Unmarshal(body, &recipeIds)
	if err != nil {
		return nil, err
	}

	return recipeIds, nil
}

func (client *APIClient) FetchKnownRecipesIds() (types.RecipeIds, error) {
	endpoint := "/account/recipes"
	response, err := client.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var recipeIds types.RecipeIds
	err = json.Unmarshal(body, &recipeIds)
	if err != nil {
		return nil, err
	}

	return recipeIds, nil
}

func (client *APIClient) FetchRecipe(recipeID int) (*types.Recipe, error) {
	endpoint := fmt.Sprintf("/recipes/%d", recipeID)
	response, err := client.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var recipe types.Recipe
	err = json.Unmarshal(body, &recipe)
	if err != nil {
		return nil, err
	}

	return &recipe, nil
}

func (client *APIClient) FetchItem(recipeID int) (*types.Item, error) {
	endpoint := fmt.Sprintf("/items/%d", recipeID)
	response, err := client.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var item types.Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

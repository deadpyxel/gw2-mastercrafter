package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIClient struct {
	baseURL   string
	authToken string
}

func (client APIClient) String() string {
	return fmt.Sprintf("API Client{baseURL: %s, authToken: ########}", client.baseURL)
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

func (client *APIClient) fetchAndDecode(endpoint string, targetType interface{}) error {
	response, err := client.makeRequest(endpoint)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("API request error: StatusCode=%s, Response: %+v", response.Status, response)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, targetType)
	if err != nil {
		return err
	}

	return nil
}

func (client *APIClient) FetchBuildNumber() (Metadata, error) {
	endpoint := "/build"
	var metadata Metadata
	err := client.fetchAndDecode(endpoint, &metadata)
	return metadata, err
}

func (client *APIClient) FetchAvailableRecipesIds(itemID int) (RecipeIds, error) {
	endpoint := fmt.Sprintf("/recipes/search?input=%d", itemID)
	var recipeIds RecipeIds
	err := client.fetchAndDecode(endpoint, &recipeIds)
	return recipeIds, err
}

func (client *APIClient) FetchKnownRecipesIds() (RecipeIds, error) {
	endpoint := "/account/recipes"
	var knownRecipeIds RecipeIds
	err := client.fetchAndDecode(endpoint, &knownRecipeIds)
	return knownRecipeIds, err
}

func (client *APIClient) FetchAllRecipesIds() (RecipeIds, error) {
	endpoint := "/recipes/"
	var recipeIds RecipeIds
	err := client.fetchAndDecode(endpoint, &recipeIds)
	return recipeIds, err
}

func (client *APIClient) FetchRecipe(recipeID int) (*Recipe, error) {
	endpoint := fmt.Sprintf("/recipes/%d", recipeID)
	var recipe Recipe
	err := client.fetchAndDecode(endpoint, &recipe)
	return &recipe, err
}

func (client *APIClient) FetchItem(itemID int) (*Item, error) {
	endpoint := fmt.Sprintf("/items/%d", itemID)
	var item Item
	err := client.fetchAndDecode(endpoint, &item)
	return &item, err
}

func (client *APIClient) FetchItemPrice(itemID int) (*ItemPrice, error) {
	endpoint := fmt.Sprintf("/commerce/prices/%d", itemID)
	var itemPrice ItemPrice
	err := client.fetchAndDecode(endpoint, &itemPrice)
	return &itemPrice, err
}

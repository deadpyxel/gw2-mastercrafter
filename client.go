package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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

func formatIntSliceAsStr(ids []int) string {
	idsAsStr := make([]string, len(ids))
	for i, id := range ids {
		idsAsStr[i] = strconv.Itoa(id)
	}
	return strings.Join(idsAsStr, ",")
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
		return fmt.Errorf("API request error querying [%s]: StatusCode=%s, Response: %+v", endpoint, response.Status, response)
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

func (client *APIClient) FetchAllIds(endpoint string) ([]int, error) {
	var ids []int
	err := client.fetchAndDecode(endpoint, &ids)
	return ids, err
}

func (client *APIClient) BatchFetch(ids []int, endpoint string, dataType string) ([]interface{}, error) {
	idsAsStr := formatIntSliceAsStr(ids)
	endpoint = fmt.Sprintf("%s?ids=%s", endpoint, idsAsStr)

	switch dataType {
	case "item":
		var items []Item
		err := client.fetchAndDecode(endpoint, &items)
		//Convert []Item to []interface{}
		data := make([]interface{}, len(items))
		for i, v := range items {
			data[i] = v
		}
		return data, err
	case "recipes":
		var recipes []Recipe
		err := client.fetchAndDecode(endpoint, &recipes)
		//Convert []Item to []interface{}
		data := make([]interface{}, len(recipes))
		for i, v := range recipes {
			data[i] = v
		}
		return data, err
	default:
		return nil, fmt.Errorf("Invalid data type %s", dataType)
	}
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

func (client *APIClient) BatchFetchRecipes(recipeIds RecipeIds) ([]Recipe, error) {
	recipeIdsAsStr := formatIntSliceAsStr(recipeIds)
	endpoint := fmt.Sprintf("/recipes?ids=%s", recipeIdsAsStr)
	var recipes []Recipe
	err := client.fetchAndDecode(endpoint, &recipes)
	return recipes, err
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

func (client *APIClient) FetchAllItemsIds() ([]int, error) {
	endpoint := "/items/"
	var itemIds []int
	err := client.fetchAndDecode(endpoint, &itemIds)
	return itemIds, err
}

func (client *APIClient) BatchFetchItems(itemIds []int) ([]Item, error) {
	itemIdsAsStr := formatIntSliceAsStr(itemIds)
	endpoint := fmt.Sprintf("/items?ids=%s", itemIdsAsStr)
	var items []Item
	err := client.fetchAndDecode(endpoint, &items)
	return items, err
}

func (client *APIClient) FetchItemPrice(itemID int) (*ItemPrice, error) {
	endpoint := fmt.Sprintf("/commerce/prices/%d", itemID)
	var itemPrice ItemPrice
	err := client.fetchAndDecode(endpoint, &itemPrice)
	return &itemPrice, err
}

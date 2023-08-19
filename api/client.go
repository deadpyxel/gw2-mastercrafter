package api

import "net/http"

type APIClient struct {
	baseURL   string
	authToken string
}

func NewAPIClient(baseURL, authToken string) *APIClient {
	return &APIClient{baseURL: baseURL, authToken: authToken}
}

func (client *APIClient) MakeRequest(endpoint string) (*http.Response, error) {
	url := client.baseURL + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+client.authToken)
	httpClient := &http.Client{}
	return httpClient.Do(req)
}

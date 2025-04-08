package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestCreateServiceAccount(t *testing.T) {
	// Test data
	projectID := "proj_123"
	serviceAccountName := "test-service-account"
	expectedServiceAccount := ServiceAccount{
		ID:        "sa_123",
		Object:    "service_account",
		Name:      serviceAccountName,
		Role:      "owner",
		CreatedAt: 1617123456,
		APIKey: struct {
			Object    string `json:"object"`
			Value     string `json:"value"`
			Name      string `json:"name"`
			CreatedAt int64  `json:"created_at"`
			ID        string `json:"id"`
		}{
			Object:    "api_key",
			Value:     "sk-test-key",
			Name:      "test-key",
			CreatedAt: 1617123456,
			ID:        "key_123",
		},
	}
	responseBody, _ := json.Marshal(expectedServiceAccount)

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify request
			if req.Method != "POST" {
				t.Errorf("Expected method to be POST, got %s", req.Method)
			}
			expectedURL := "https://api.openai.com/v1/organization/projects/" + projectID + "/service_accounts"
			if req.URL.String() != expectedURL {
				t.Errorf("Expected URL to be %s, got %s", expectedURL, req.URL.String())
			}

			// Verify request body
			body, _ := io.ReadAll(req.Body)
			var requestBody map[string]string
			json.Unmarshal(body, &requestBody)
			if requestBody["name"] != serviceAccountName {
				t.Errorf("Expected request body to contain name=%s, got %v", serviceAccountName, requestBody)
			}

			// Return mock response
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(responseBody))),
			}, nil
		},
	}

	// Create client
	client := &Client{
		APIKey:     "test-api-key",
		HTTPClient: mockClient,
		BaseURL:    "https://api.openai.com/v1/organization",
	}

	// Test CreateServiceAccount
	serviceAccount, err := client.CreateServiceAccount(context.Background(), projectID, serviceAccountName)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(*serviceAccount, expectedServiceAccount) {
		t.Errorf("Expected service account to be %+v, got %+v", expectedServiceAccount, *serviceAccount)
	}
}

func TestListServiceAccounts(t *testing.T) {
	// Test data
	projectID := "proj_123"
	serviceAccounts := []ServiceAccount{
		{
			ID:        "sa_123",
			Object:    "service_account",
			Name:      "service-account-1",
			Role:      "owner",
			CreatedAt: 1617123456,
		},
		{
			ID:        "sa_456",
			Object:    "service_account",
			Name:      "service-account-2",
			Role:      "owner",
			CreatedAt: 1617123456,
		},
	}
	listResponse := ListServiceAccountResponse{
		Object:  "list",
		Data:    serviceAccounts,
		FirstID: "sa_123",
		LastID:  "sa_456",
		HasMore: false,
	}
	responseBody, _ := json.Marshal(listResponse)

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify request
			if req.Method != "GET" {
				t.Errorf("Expected method to be GET, got %s", req.Method)
			}
			expectedURLPrefix := "https://api.openai.com/v1/organization/projects/" + projectID + "/service_accounts"
			if !strings.HasPrefix(req.URL.String(), expectedURLPrefix) {
				t.Errorf("Expected URL to start with %s, got %s", expectedURLPrefix, req.URL.String())
			}

			// Return mock response
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(responseBody))),
			}, nil
		},
	}

	// Create client
	client := &Client{
		APIKey:     "test-api-key",
		HTTPClient: mockClient,
		BaseURL:    "https://api.openai.com/v1/organization",
	}

	// Test ListServiceAccounts
	result, err := client.ListServiceAccounts(context.Background(), projectID)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(*result, serviceAccounts) {
		t.Errorf("Expected service accounts to be %+v, got %+v", serviceAccounts, *result)
	}
}

func TestListServiceAccounts_Pagination(t *testing.T) {
	// Test data
	projectID := "proj_123"

	// First page data
	firstPageAccounts := []ServiceAccount{
		{
			ID:        "sa_123",
			Object:    "service_account",
			Name:      "service-account-1",
			Role:      "owner",
			CreatedAt: 1617123456,
		},
	}
	firstPageResponse := ListServiceAccountResponse{
		Object:  "list",
		Data:    firstPageAccounts,
		FirstID: "sa_123",
		LastID:  "sa_123",
		HasMore: true,
	}
	firstPageBody, _ := json.Marshal(firstPageResponse)

	// Second page data
	secondPageAccounts := []ServiceAccount{
		{
			ID:        "sa_456",
			Object:    "service_account",
			Name:      "service-account-2",
			Role:      "owner",
			CreatedAt: 1617123456,
		},
	}
	secondPageResponse := ListServiceAccountResponse{
		Object:  "list",
		Data:    secondPageAccounts,
		FirstID: "sa_456",
		LastID:  "sa_456",
		HasMore: false,
	}
	secondPageBody, _ := json.Marshal(secondPageResponse)

	// Create mock HTTP client
	callCount := 0
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			callCount++

			// First call should have no 'after' parameter
			// Second call should have 'after' parameter set to 'sa_123'
			if callCount == 1 {
				if req.URL.Query().Get("after") != "" {
					t.Errorf("Expected no 'after' query parameter on first call, got '%s'", req.URL.Query().Get("after"))
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(string(firstPageBody))),
				}, nil
			} else {
				if req.URL.Query().Get("after") != "sa_123" {
					t.Errorf("Expected 'after' query parameter to be 'sa_123' on second call, got '%s'", req.URL.Query().Get("after"))
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(string(secondPageBody))),
				}, nil
			}
		},
	}

	// Create client
	client := &Client{
		APIKey:     "test-api-key",
		HTTPClient: mockClient,
		BaseURL:    "https://api.openai.com/v1/organization",
	}

	// Test ListServiceAccounts
	accounts, err := client.ListServiceAccounts(context.Background(), projectID)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have combined both pages
	expectedAccounts := append(firstPageAccounts, secondPageAccounts...)
	if !reflect.DeepEqual(*accounts, expectedAccounts) {
		t.Errorf("Expected accounts to be %+v, got %+v", expectedAccounts, *accounts)
	}

	// Should have made exactly 2 calls
	if callCount != 2 {
		t.Errorf("Expected 2 API calls, got %d", callCount)
	}
}

func TestListServiceAccounts_SinglePage(t *testing.T) {
	// Test data
	projectID := "proj_123"
	after := "sa_prev"
	limit := 50

	serviceAccounts := []ServiceAccount{
		{
			ID:        "sa_123",
			Object:    "service_account",
			Name:      "service-account-1",
			Role:      "owner",
			CreatedAt: 1617123456,
		},
	}
	listResponse := ListServiceAccountResponse{
		Object:  "list",
		Data:    serviceAccounts,
		FirstID: "sa_123",
		LastID:  "sa_123",
		HasMore: false,
	}
	responseBody, _ := json.Marshal(listResponse)

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify request
			if req.Method != "GET" {
				t.Errorf("Expected method to be GET, got %s", req.Method)
			}

			// Check query parameters
			query := req.URL.Query()
			if query.Get("after") != after {
				t.Errorf("Expected 'after' query parameter to be '%s', got '%s'", after, query.Get("after"))
			}
			if query.Get("limit") != "50" {
				t.Errorf("Expected 'limit' query parameter to be '50', got '%s'", query.Get("limit"))
			}

			// Return mock response
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(responseBody))),
			}, nil
		},
	}

	// Create client
	client := &Client{
		APIKey:     "test-api-key",
		HTTPClient: mockClient,
		BaseURL:    "https://api.openai.com/v1/organization",
	}

	// Test listServiceAccounts
	result, err := client.listServiceAccounts(context.Background(), projectID, after, limit)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(*result, listResponse) {
		t.Errorf("Expected result to be %+v, got %+v", listResponse, *result)
	}
}

func TestDeleteServiceAccount(t *testing.T) {
	// Test data
	projectID := "proj_123"
	serviceAccountID := "sa_456"
	expectedResponse := DeletedServiceAccountResponse{
		Object:  "service_account.deleted",
		ID:      serviceAccountID,
		Deleted: true,
	}
	responseBody, _ := json.Marshal(expectedResponse)

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify request
			if req.Method != "DELETE" {
				t.Errorf("Expected method to be DELETE, got %s", req.Method)
			}
			expectedURL := "https://api.openai.com/v1/organization/projects/" + projectID + "/service_accounts/" + serviceAccountID
			if req.URL.String() != expectedURL {
				t.Errorf("Expected URL to be %s, got %s", expectedURL, req.URL.String())
			}

			// Return mock response
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(responseBody))),
			}, nil
		},
	}

	// Create client
	client := &Client{
		APIKey:     "test-api-key",
		HTTPClient: mockClient,
		BaseURL:    "https://api.openai.com/v1/organization",
	}

	// Test DeleteServiceAccount
	result, err := client.DeleteServiceAccount(context.Background(), projectID, serviceAccountID)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(*result, expectedResponse) {
		t.Errorf("Expected result to be %+v, got %+v", expectedResponse, *result)
	}
}

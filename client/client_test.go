package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// MockHTTPClient is a mock implementation of the HTTPClient interface for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	httpClient := &MockHTTPClient{}

	client := NewClient(apiKey, httpClient)

	if client.APIKey != apiKey {
		t.Errorf("Expected APIKey to be %s, got %s", apiKey, client.APIKey)
	}

	if client.HTTPClient != httpClient {
		t.Errorf("Expected HTTPClient to be %v, got %v", httpClient, client.HTTPClient)
	}

	if client.BaseURL != "https://api.openai.com/v1/organization" {
		t.Errorf("Expected BaseURL to be %s, got %s", "https://api.openai.com/v1/organization", client.BaseURL)
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		StatusCode: 400,
		Message:    "Bad Request",
	}

	expected := "receive api response: Bad Request (status code: 400)"
	if err.Error() != expected {
		t.Errorf("Expected error message to be %s, got %s", expected, err.Error())
	}
}

func TestDoRequest_Success(t *testing.T) {
	// Test data
	expectedResponse := map[string]string{"key": "value"}
	responseBody, _ := json.Marshal(expectedResponse)

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify request
			if req.Method != "GET" {
				t.Errorf("Expected method to be GET, got %s", req.Method)
			}
			if req.URL.String() != "https://api.openai.com/v1/organization/test-path?param=value" {
				t.Errorf("Expected URL to be %s, got %s", "https://api.openai.com/v1/organization/test-path?param=value", req.URL.String())
			}
			if req.Header.Get("Authorization") != "Bearer test-api-key" {
				t.Errorf("Expected Authorization header to be %s, got %s", "Bearer test-api-key", req.Header.Get("Authorization"))
			}
			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type header to be %s, got %s", "application/json", req.Header.Get("Content-Type"))
			}

			// Return mock response
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
			}, nil
		},
	}

	// Create client
	client := &Client{
		APIKey:     "test-api-key",
		HTTPClient: mockClient,
		BaseURL:    "https://api.openai.com/v1/organization",
	}

	// Test doRequest
	query := url.Values{}
	query.Set("param", "value")
	result, err := client.doRequest(context.Background(), "GET", "/test-path", query, nil)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(result) != string(responseBody) {
		t.Errorf("Expected response body to be %s, got %s", string(responseBody), string(result))
	}
}

func TestDoRequest_WithRequestBody(t *testing.T) {
	// Test data
	requestBody := map[string]string{"request": "data"}
	expectedResponse := map[string]string{"key": "value"}
	responseBody, _ := json.Marshal(expectedResponse)

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify request body
			body, _ := io.ReadAll(req.Body)
			var receivedBody map[string]string
			json.Unmarshal(body, &receivedBody)
			if receivedBody["request"] != "data" {
				t.Errorf("Expected request body to contain %v, got %v", requestBody, receivedBody)
			}

			// Return mock response
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
			}, nil
		},
	}

	// Create client
	client := &Client{
		APIKey:     "test-api-key",
		HTTPClient: mockClient,
		BaseURL:    "https://api.openai.com/v1/organization",
	}

	// Test doRequest with body
	result, err := client.doRequest(context.Background(), "POST", "/test-path", nil, requestBody)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(result) != string(responseBody) {
		t.Errorf("Expected response body to be %s, got %s", string(responseBody), string(result))
	}
}

func TestDoRequest_HTTPClientError(t *testing.T) {
	// Create mock HTTP client that returns an error
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("http client error")
		},
	}

	// Create client
	client := &Client{
		APIKey:     "test-api-key",
		HTTPClient: mockClient,
		BaseURL:    "https://api.openai.com/v1/organization",
	}

	// Test doRequest
	_, err := client.doRequest(context.Background(), "GET", "/test-path", nil, nil)

	// Verify error
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "execute http request: http client error") {
		t.Errorf("Expected error to contain 'execute http request: http client error', got '%v'", err)
	}
}

func TestDoRequest_APIError(t *testing.T) {
	// Create mock HTTP client that returns an API error
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 400,
				Body:       io.NopCloser(strings.NewReader("Bad Request")),
			}, nil
		},
	}

	// Create client
	client := &Client{
		APIKey:     "test-api-key",
		HTTPClient: mockClient,
		BaseURL:    "https://api.openai.com/v1/organization",
	}

	// Test doRequest
	_, err := client.doRequest(context.Background(), "GET", "/test-path", nil, nil)

	// Verify error
	if err == nil {
		t.Error("Expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("Expected error to be of type *APIError, got %T", err)
	}

	if apiErr.StatusCode != 400 {
		t.Errorf("Expected status code to be 400, got %d", apiErr.StatusCode)
	}

	if apiErr.Message != "Bad Request" {
		t.Errorf("Expected message to be 'Bad Request', got '%s'", apiErr.Message)
	}
}

func TestDoRequest_InvalidJSON(t *testing.T) {
	// Test data
	requestBody := make(chan int) // Channels cannot be marshaled to JSON

	// Create client
	client := &Client{
		APIKey:     "test-api-key",
		HTTPClient: &MockHTTPClient{},
		BaseURL:    "https://api.openai.com/v1/organization",
	}

	// Test doRequest with invalid JSON body
	_, err := client.doRequest(context.Background(), "POST", "/test-path", nil, requestBody)

	// Verify error
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "marshal json") {
		t.Errorf("Expected error to contain 'marshal json', got '%v'", err)
	}
}

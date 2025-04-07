package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

// ---------------------------
// Structs
// ---------------------------

// HTTPClient interface for making HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client for interacting with the OpenAI API.
type Client struct {
	APIKey     string
	HTTPClient HTTPClient
	BaseURL    string
}

// APIError is a custom error type for API errors.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("receive api response: %s (status code: %d)", e.Message, e.StatusCode)
}

// NewClient creates a new Client with a custom HTTP client.
func NewClient(apiKey string, httpClient HTTPClient) *Client {
	return &Client{
		APIKey:     apiKey,
		HTTPClient: httpClient,
		BaseURL:    "https://api.openai.com/v1/organization",
	}
}

func (c *Client) doRequest(ctx context.Context, method string, path string, query url.Values, body interface{}) ([]byte, error) {
	fullURL := c.BaseURL + path
	if query != nil {
		fullURL += "?" + query.Encode()
	}

	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal json: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute http request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("close response body", "error", err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return respBody, nil
}

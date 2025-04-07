package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type ServiceAccount struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	CreatedAt int64  `json:"created_at"`
	APIKey    struct {
		Object    string `json:"object"`
		Value     string `json:"value"`
		Name      string `json:"name"`
		CreatedAt int64  `json:"created_at"`
		ID        string `json:"id"`
	} `json:"api_key"`
}

type ListServiceAccountResponse struct {
	Object  string           `json:"object"`
	Data    []ServiceAccount `json:"data"`
	FirstID string           `json:"first_id"`
	LastID  string           `json:"last_id"`
	HasMore bool             `json:"has_more"`
}

type DeletedServiceAccountResponse struct {
	Object  string `json:"object"`
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// CreateServiceAccount creates a new service account in the project and returns its API key.
func (c *Client) CreateServiceAccount(ctx context.Context, projectID string, name string) (*ServiceAccount, error) {
	body := map[string]string{"name": name}
	respBody, err := c.doRequest(ctx, "POST", fmt.Sprintf("/projects/%s/service_accounts", projectID), nil, body)
	if err != nil {
		return nil, fmt.Errorf("create service account: %w", err)
	}
	var sa ServiceAccount
	err = json.Unmarshal(respBody, &sa)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}
	return &sa, nil
}

// ListServiceAccounts lists all service accounts with optional pagination.
func (c *Client) ListServiceAccounts(ctx context.Context, projectID string) (*[]ServiceAccount, error) {
	var allAccounts []ServiceAccount
	var after string
	const pageSize = 100

	for {
		resp, err := c.listServiceAccounts(ctx, projectID, after, pageSize)
		if err != nil {
			return nil, fmt.Errorf("get service account list: %w", err)
		}
		allAccounts = append(allAccounts, resp.Data...)
		if !resp.HasMore {
			break
		}
		after = resp.LastID
	}

	return &allAccounts, nil
}

// listServiceAccounts retrieves all service accounts with optional pagination.
func (c *Client) listServiceAccounts(ctx context.Context, projectID string, after string, limit int) (*ListServiceAccountResponse, error) {
	query := url.Values{}
	if after != "" {
		query.Set("after", after)
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	path := fmt.Sprintf("/projects/%s/service_accounts", projectID)
	respBody, err := c.doRequest(ctx, "GET", path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	var result ListServiceAccountResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}
	return &result, nil
}

// DeleteServiceAccount deletes a specific service account from a project.
func (c *Client) DeleteServiceAccount(ctx context.Context, projectID string, serviceAccountID string) (*DeletedServiceAccountResponse, error) {
	respBody, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/projects/%s/service_accounts/%s", projectID, serviceAccountID), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("delete service account: %w", err)
	}
	var result DeletedServiceAccountResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}
	return &result, nil
}

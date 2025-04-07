package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type Project struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Name       string `json:"name"`
	CreatedAt  int64  `json:"created_at"`
	ArchivedAt *int64 `json:"archived_at"`
	Status     string `json:"status"`
}

type ListProjectResponse struct {
	Object  string    `json:"object"`
	Data    []Project `json:"data"`
	FirstID string    `json:"first_id"`
	LastID  string    `json:"last_id"`
	HasMore bool      `json:"has_more"`
}

// CreateProject creates a new project.
func (c *Client) CreateProject(ctx context.Context, name string) (*Project, error) {
	body := map[string]string{"name": name}
	respBody, err := c.doRequest(ctx, "POST", "/projects", nil, body)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	var project Project
	err = json.Unmarshal(respBody, &project)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}
	return &project, nil
}

// GetProject retrieves a project by its name.
func (c *Client) GetProject(ctx context.Context, projectName string) (*Project, bool, error) {
	projects, err := c.listProjects(ctx, false)
	if err != nil {
		return nil, false, fmt.Errorf("get project list: %w", err)
	}
	for _, project := range *projects {
		if project.Name == projectName {
			return &project, true, nil
		}
	}
	return nil, false, nil
}

// ListAllProjects retrieves all projects across all pages.
func (c *Client) listProjects(ctx context.Context, includeArchived bool) (*[]Project, error) {
	var allProjects []Project
	var after string
	const pageSize = 100

	for {
		resp, err := c.listProject(ctx, after, pageSize, includeArchived)
		if err != nil {
			return nil, fmt.Errorf("get project list: %w", err)
		}
		allProjects = append(allProjects, resp.Data...)
		if !resp.HasMore {
			break
		}
		after = resp.LastID
	}

	return &allProjects, nil
}

// listProject retrieves all projects with optional pagination and filtering.
func (c *Client) listProject(ctx context.Context, after string, limit int, includeArchived bool) (*ListProjectResponse, error) {
	query := url.Values{}
	if after != "" {
		query.Set("after", after)
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if includeArchived {
		query.Set("include_archived", "true")
	}

	respBody, err := c.doRequest(ctx, "GET", "/projects", query, nil)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	var result ListProjectResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}
	return &result, nil
}

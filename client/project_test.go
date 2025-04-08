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

func TestCreateProject(t *testing.T) {
	// Test data
	projectName := "test-project"
	expectedProject := Project{
		ID:        "proj_123",
		Object:    "project",
		Name:      projectName,
		CreatedAt: 1617123456,
		Status:    "active",
	}
	responseBody, _ := json.Marshal(expectedProject)

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify request
			if req.Method != "POST" {
				t.Errorf("Expected method to be POST, got %s", req.Method)
			}
			if req.URL.String() != "https://api.openai.com/v1/organization/projects" {
				t.Errorf("Expected URL to be %s, got %s", "https://api.openai.com/v1/organization/projects", req.URL.String())
			}

			// Verify request body
			body, _ := io.ReadAll(req.Body)
			var requestBody map[string]string
			if err := json.Unmarshal(body, &requestBody); err != nil {
				t.Errorf("Failed to unmarshal request body: %v", err)
				return nil, err
			}
			if requestBody["name"] != projectName {
				t.Errorf("Expected request body to contain name=%s, got %v", projectName, requestBody)
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

	// Test CreateProject
	project, err := client.CreateProject(context.Background(), projectName)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(*project, expectedProject) {
		t.Errorf("Expected project to be %+v, got %+v", expectedProject, *project)
	}
}

func TestGetProject_Found(t *testing.T) {
	// Test data
	projectName := "test-project"
	projects := []Project{
		{
			ID:        "proj_123",
			Object:    "project",
			Name:      "other-project",
			CreatedAt: 1617123456,
			Status:    "active",
		},
		{
			ID:        "proj_456",
			Object:    "project",
			Name:      projectName,
			CreatedAt: 1617123456,
			Status:    "active",
		},
	}
	listResponse := ListProjectResponse{
		Object:  "list",
		Data:    projects,
		FirstID: "proj_123",
		LastID:  "proj_456",
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
			if !strings.HasPrefix(req.URL.String(), "https://api.openai.com/v1/organization/projects") {
				t.Errorf("Expected URL to start with %s, got %s", "https://api.openai.com/v1/organization/projects", req.URL.String())
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

	// Test GetProject
	project, found, err := client.GetProject(context.Background(), projectName)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !found {
		t.Error("Expected project to be found, but it wasn't")
	}
	if project.ID != "proj_456" || project.Name != projectName {
		t.Errorf("Expected project with ID=proj_456 and Name=%s, got ID=%s and Name=%s", projectName, project.ID, project.Name)
	}
}

func TestGetProject_NotFound(t *testing.T) {
	// Test data
	projectName := "non-existent-project"
	projects := []Project{
		{
			ID:        "proj_123",
			Object:    "project",
			Name:      "other-project",
			CreatedAt: 1617123456,
			Status:    "active",
		},
	}
	listResponse := ListProjectResponse{
		Object:  "list",
		Data:    projects,
		FirstID: "proj_123",
		LastID:  "proj_123",
		HasMore: false,
	}
	responseBody, _ := json.Marshal(listResponse)

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
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

	// Test GetProject
	_, found, err := client.GetProject(context.Background(), projectName)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if found {
		t.Error("Expected project not to be found, but it was")
	}
}

func TestListProject(t *testing.T) {
	// Test data
	projects := []Project{
		{
			ID:        "proj_123",
			Object:    "project",
			Name:      "project-1",
			CreatedAt: 1617123456,
			Status:    "active",
		},
		{
			ID:        "proj_456",
			Object:    "project",
			Name:      "project-2",
			CreatedAt: 1617123456,
			Status:    "active",
		},
	}
	listResponse := ListProjectResponse{
		Object:  "list",
		Data:    projects,
		FirstID: "proj_123",
		LastID:  "proj_456",
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
			if query.Get("after") != "test-after" {
				t.Errorf("Expected 'after' query parameter to be 'test-after', got '%s'", query.Get("after"))
			}
			if query.Get("limit") != "50" {
				t.Errorf("Expected 'limit' query parameter to be '50', got '%s'", query.Get("limit"))
			}
			if query.Get("include_archived") != "true" {
				t.Errorf("Expected 'include_archived' query parameter to be 'true', got '%s'", query.Get("include_archived"))
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

	// Test listProject
	result, err := client.listProject(context.Background(), "test-after", 50, true)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(*result, listResponse) {
		t.Errorf("Expected result to be %+v, got %+v", listResponse, *result)
	}
}

func TestListProjects_Pagination(t *testing.T) {
	// Test data for first page
	firstPageProjects := []Project{
		{
			ID:        "proj_123",
			Object:    "project",
			Name:      "project-1",
			CreatedAt: 1617123456,
			Status:    "active",
		},
	}
	firstPageResponse := ListProjectResponse{
		Object:  "list",
		Data:    firstPageProjects,
		FirstID: "proj_123",
		LastID:  "proj_123",
		HasMore: true,
	}
	firstPageBody, _ := json.Marshal(firstPageResponse)

	// Test data for second page
	secondPageProjects := []Project{
		{
			ID:        "proj_456",
			Object:    "project",
			Name:      "project-2",
			CreatedAt: 1617123456,
			Status:    "active",
		},
	}
	secondPageResponse := ListProjectResponse{
		Object:  "list",
		Data:    secondPageProjects,
		FirstID: "proj_456",
		LastID:  "proj_456",
		HasMore: false,
	}
	secondPageBody, _ := json.Marshal(secondPageResponse)

	// Create mock HTTP client
	callCount := 0
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			callCount++

			// First call should have no 'after' parameter
			// Second call should have 'after' parameter set to 'proj_123'
			if callCount == 1 {
				if req.URL.Query().Get("after") != "" {
					t.Errorf("Expected no 'after' query parameter on first call, got '%s'", req.URL.Query().Get("after"))
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(string(firstPageBody))),
				}, nil
			} else {
				if req.URL.Query().Get("after") != "proj_123" {
					t.Errorf("Expected 'after' query parameter to be 'proj_123' on second call, got '%s'", req.URL.Query().Get("after"))
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

	// Test listProjects
	projects, err := client.listProjects(context.Background(), false)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have combined both pages
	expectedProjects := append(firstPageProjects, secondPageProjects...)
	if !reflect.DeepEqual(*projects, expectedProjects) {
		t.Errorf("Expected projects to be %+v, got %+v", expectedProjects, *projects)
	}

	// Should have made exactly 2 calls
	if callCount != 2 {
		t.Errorf("Expected 2 API calls, got %d", callCount)
	}
}

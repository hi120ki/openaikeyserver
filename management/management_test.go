package management

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hi120ki/monorepo/projects/openaikeyserver/client"
)

// MockClient is a mock implementation of client.APIClient
type MockClient struct {
	GetProjectFunc           func(ctx context.Context, projectName string) (*client.Project, bool, error)
	CreateProjectFunc        func(ctx context.Context, name string) (*client.Project, error)
	CreateServiceAccountFunc func(ctx context.Context, projectID string, name string) (*client.ServiceAccount, error)
	ListServiceAccountsFunc  func(ctx context.Context, projectID string) (*[]client.ServiceAccount, error)
	DeleteServiceAccountFunc func(ctx context.Context, projectID string, serviceAccountID string) (*client.DeletedServiceAccountResponse, error)
}

// Override methods with mock implementations
func (m *MockClient) GetProject(ctx context.Context, projectName string) (*client.Project, bool, error) {
	if m.GetProjectFunc != nil {
		return m.GetProjectFunc(ctx, projectName)
	}
	return nil, false, nil
}

func (m *MockClient) CreateProject(ctx context.Context, name string) (*client.Project, error) {
	if m.CreateProjectFunc != nil {
		return m.CreateProjectFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockClient) CreateServiceAccount(ctx context.Context, projectID string, name string) (*client.ServiceAccount, error) {
	if m.CreateServiceAccountFunc != nil {
		return m.CreateServiceAccountFunc(ctx, projectID, name)
	}
	return nil, nil
}

func (m *MockClient) ListServiceAccounts(ctx context.Context, projectID string) (*[]client.ServiceAccount, error) {
	if m.ListServiceAccountsFunc != nil {
		return m.ListServiceAccountsFunc(ctx, projectID)
	}
	return nil, nil
}

func (m *MockClient) DeleteServiceAccount(ctx context.Context, projectID string, serviceAccountID string) (*client.DeletedServiceAccountResponse, error) {
	if m.DeleteServiceAccountFunc != nil {
		return m.DeleteServiceAccountFunc(ctx, projectID, serviceAccountID)
	}
	return nil, nil
}

func TestNewManagement(t *testing.T) {
	// Test data
	client := &MockClient{}
	expiration := 24 * time.Hour

	// Test NewManagement
	m := NewManagement(client, expiration)

	// Verify result
	if m == nil {
		t.Fatal("Expected non-nil Management")
	}
}

func TestCreateAPIKey_ExistingProject(t *testing.T) {
	// Test data
	projectName := "test-project"
	serviceAccountName := "test-service-account"
	projectID := "proj_123"
	apiKeyValue := "sk-test-key"
	expiration := 24 * time.Hour

	// Create mock client
	mockClient := &MockClient{
		GetProjectFunc: func(ctx context.Context, name string) (*client.Project, bool, error) {
			if name != projectName {
				t.Errorf("Expected project name to be '%s', got '%s'", projectName, name)
			}
			return &client.Project{
				ID:   projectID,
				Name: projectName,
			}, true, nil
		},
		CreateServiceAccountFunc: func(ctx context.Context, projID string, name string) (*client.ServiceAccount, error) {
			if projID != projectID {
				t.Errorf("Expected project ID to be '%s', got '%s'", projectID, projID)
			}
			if name != serviceAccountName {
				t.Errorf("Expected service account name to be '%s', got '%s'", serviceAccountName, name)
			}
			return &client.ServiceAccount{
				ID:   "sa_123",
				Name: serviceAccountName,
				APIKey: struct {
					Object    string `json:"object"`
					Value     string `json:"value"`
					Name      string `json:"name"`
					CreatedAt int64  `json:"created_at"`
					ID        string `json:"id"`
				}{
					Value: apiKeyValue,
				},
			}, nil
		},
	}

	// Create management
	management := NewManagement(mockClient, expiration)

	// Test CreateAPIKey
	key, expirationTime, err := management.CreateAPIKey(context.Background(), projectName, serviceAccountName)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if key != apiKeyValue {
		t.Errorf("Expected API key to be '%s', got '%s'", apiKeyValue, key)
	}
	if expirationTime == nil {
		t.Error("Expected non-nil expiration time")
	}
}

func TestCreateAPIKey_NewProject(t *testing.T) {
	// Test data
	projectName := "test-project"
	serviceAccountName := "test-service-account"
	projectID := "proj_123"
	apiKeyValue := "sk-test-key"
	expiration := 24 * time.Hour

	// Create mock client
	mockClient := &MockClient{
		GetProjectFunc: func(ctx context.Context, name string) (*client.Project, bool, error) {
			if name != projectName {
				t.Errorf("Expected project name to be '%s', got '%s'", projectName, name)
			}
			return nil, false, nil
		},
		CreateProjectFunc: func(ctx context.Context, name string) (*client.Project, error) {
			if name != projectName {
				t.Errorf("Expected project name to be '%s', got '%s'", projectName, name)
			}
			return &client.Project{
				ID:   projectID,
				Name: projectName,
			}, nil
		},
		CreateServiceAccountFunc: func(ctx context.Context, projID string, name string) (*client.ServiceAccount, error) {
			if projID != projectID {
				t.Errorf("Expected project ID to be '%s', got '%s'", projectID, projID)
			}
			if name != serviceAccountName {
				t.Errorf("Expected service account name to be '%s', got '%s'", serviceAccountName, name)
			}
			return &client.ServiceAccount{
				ID:   "sa_123",
				Name: serviceAccountName,
				APIKey: struct {
					Object    string `json:"object"`
					Value     string `json:"value"`
					Name      string `json:"name"`
					CreatedAt int64  `json:"created_at"`
					ID        string `json:"id"`
				}{
					Value: apiKeyValue,
				},
			}, nil
		},
	}

	// Create management
	management := NewManagement(mockClient, expiration)

	// Test CreateAPIKey
	key, expirationTime, err := management.CreateAPIKey(context.Background(), projectName, serviceAccountName)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if key != apiKeyValue {
		t.Errorf("Expected API key to be '%s', got '%s'", apiKeyValue, key)
	}
	if expirationTime == nil {
		t.Error("Expected non-nil expiration time")
	}
}

func TestCreateAPIKey_GetProjectError(t *testing.T) {
	// Test data
	projectName := "test-project"
	serviceAccountName := "test-service-account"
	expiration := 24 * time.Hour
	expectedError := errors.New("get project error")

	// Create mock client
	mockClient := &MockClient{
		GetProjectFunc: func(ctx context.Context, name string) (*client.Project, bool, error) {
			return nil, false, expectedError
		},
	}

	// Create management
	management := NewManagement(mockClient, expiration)

	// Test CreateAPIKey
	_, _, err := management.CreateAPIKey(context.Background(), projectName, serviceAccountName)

	// Verify result
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCreateAPIKey_CreateProjectError(t *testing.T) {
	// Test data
	projectName := "test-project"
	serviceAccountName := "test-service-account"
	expiration := 24 * time.Hour
	expectedError := errors.New("create project error")

	// Create mock client
	mockClient := &MockClient{
		GetProjectFunc: func(ctx context.Context, name string) (*client.Project, bool, error) {
			return nil, false, nil
		},
		CreateProjectFunc: func(ctx context.Context, name string) (*client.Project, error) {
			return nil, expectedError
		},
	}

	// Create management
	management := NewManagement(mockClient, expiration)

	// Test CreateAPIKey
	_, _, err := management.CreateAPIKey(context.Background(), projectName, serviceAccountName)

	// Verify result
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCreateAPIKey_CreateServiceAccountError(t *testing.T) {
	// Test data
	projectName := "test-project"
	serviceAccountName := "test-service-account"
	projectID := "proj_123"
	expiration := 24 * time.Hour
	expectedError := errors.New("create service account error")

	// Create mock client
	mockClient := &MockClient{
		GetProjectFunc: func(ctx context.Context, name string) (*client.Project, bool, error) {
			return &client.Project{
				ID:   projectID,
				Name: projectName,
			}, true, nil
		},
		CreateServiceAccountFunc: func(ctx context.Context, projID string, name string) (*client.ServiceAccount, error) {
			return nil, expectedError
		},
	}

	// Create management
	management := NewManagement(mockClient, expiration)

	// Test CreateAPIKey
	_, _, err := management.CreateAPIKey(context.Background(), projectName, serviceAccountName)

	// Verify result
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCleanupAPIKey(t *testing.T) {
	// Test data
	projectName := "test-project"
	projectID := "proj_123"
	expiration := 24 * time.Hour
	now := time.Now()
	oldTime := now.Add(-2 * expiration).Unix()
	newTime := now.Add(-1 * time.Hour).Unix()

	// Create mock client
	mockClient := &MockClient{
		GetProjectFunc: func(ctx context.Context, name string) (*client.Project, bool, error) {
			if name != projectName {
				t.Errorf("Expected project name to be '%s', got '%s'", projectName, name)
			}
			return &client.Project{
				ID:   projectID,
				Name: projectName,
			}, true, nil
		},
		ListServiceAccountsFunc: func(ctx context.Context, projID string) (*[]client.ServiceAccount, error) {
			if projID != projectID {
				t.Errorf("Expected project ID to be '%s', got '%s'", projectID, projID)
			}
			return &[]client.ServiceAccount{
				{
					ID:        "sa_old",
					Name:      "old-service-account",
					CreatedAt: oldTime,
				},
				{
					ID:        "sa_new",
					Name:      "new-service-account",
					CreatedAt: newTime,
				},
			}, nil
		},
		DeleteServiceAccountFunc: func(ctx context.Context, projID string, serviceAccountID string) (*client.DeletedServiceAccountResponse, error) {
			if projID != projectID {
				t.Errorf("Expected project ID to be '%s', got '%s'", projectID, projID)
			}
			if serviceAccountID != "sa_old" {
				t.Errorf("Expected service account ID to be 'sa_old', got '%s'", serviceAccountID)
			}
			return &client.DeletedServiceAccountResponse{
				ID:      serviceAccountID,
				Deleted: true,
			}, nil
		},
	}

	// Create management
	management := NewManagement(mockClient, expiration)

	// Test CleanupAPIKey
	err := management.CleanupAPIKey(context.Background(), projectName)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCleanupAPIKey_GetProjectError(t *testing.T) {
	// Test data
	projectName := "test-project"
	expiration := 24 * time.Hour
	expectedError := errors.New("get project error")

	// Create mock client
	mockClient := &MockClient{
		GetProjectFunc: func(ctx context.Context, name string) (*client.Project, bool, error) {
			return nil, false, expectedError
		},
	}

	// Create management
	management := NewManagement(mockClient, expiration)

	// Test CleanupAPIKey
	err := management.CleanupAPIKey(context.Background(), projectName)

	// Verify result
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCleanupAPIKey_ProjectNotFound(t *testing.T) {
	// Test data
	projectName := "test-project"
	expiration := 24 * time.Hour

	// Create mock client
	mockClient := &MockClient{
		GetProjectFunc: func(ctx context.Context, name string) (*client.Project, bool, error) {
			return nil, false, nil
		},
	}

	// Create management
	management := NewManagement(mockClient, expiration)

	// Test CleanupAPIKey
	err := management.CleanupAPIKey(context.Background(), projectName)

	// Verify result
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCleanupAPIKey_ListServiceAccountsError(t *testing.T) {
	// Test data
	projectName := "test-project"
	projectID := "proj_123"
	expiration := 24 * time.Hour
	expectedError := errors.New("list service accounts error")

	// Create mock client
	mockClient := &MockClient{
		GetProjectFunc: func(ctx context.Context, name string) (*client.Project, bool, error) {
			return &client.Project{
				ID:   projectID,
				Name: projectName,
			}, true, nil
		},
		ListServiceAccountsFunc: func(ctx context.Context, projID string) (*[]client.ServiceAccount, error) {
			return nil, expectedError
		},
	}

	// Create management
	management := NewManagement(mockClient, expiration)

	// Test CleanupAPIKey
	err := management.CleanupAPIKey(context.Background(), projectName)

	// Verify result
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCleanupAPIKey_DeleteServiceAccountError(t *testing.T) {
	// Test data
	projectName := "test-project"
	projectID := "proj_123"
	expiration := 24 * time.Hour
	now := time.Now()
	oldTime := now.Add(-2 * expiration).Unix()
	expectedError := errors.New("delete service account error")

	// Create mock client
	mockClient := &MockClient{
		GetProjectFunc: func(ctx context.Context, name string) (*client.Project, bool, error) {
			return &client.Project{
				ID:   projectID,
				Name: projectName,
			}, true, nil
		},
		ListServiceAccountsFunc: func(ctx context.Context, projID string) (*[]client.ServiceAccount, error) {
			return &[]client.ServiceAccount{
				{
					ID:        "sa_old",
					Name:      "old-service-account",
					CreatedAt: oldTime,
				},
			}, nil
		},
		DeleteServiceAccountFunc: func(ctx context.Context, projID string, serviceAccountID string) (*client.DeletedServiceAccountResponse, error) {
			return nil, expectedError
		},
	}

	// Create management
	management := NewManagement(mockClient, expiration)

	// Test CleanupAPIKey
	err := management.CleanupAPIKey(context.Background(), projectName)

	// Verify result
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

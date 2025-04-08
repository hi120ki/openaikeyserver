package management

import (
	"context"
	"fmt"
	"time"

	"github.com/hi120ki/monorepo/projects/openaikeyserver/client"
)

// Manager defines the interface for API key management operations
type Manager interface {
	CreateAPIKey(ctx context.Context, projectName, serviceAccountName string) (string, *time.Time, error)
	CleanupAPIKey(ctx context.Context, projectName string) error
}

type Management struct {
	client     client.APIClient
	expiration time.Duration
}

func NewManagement(client client.APIClient, expiration time.Duration) *Management {
	return &Management{
		client:     client,
		expiration: expiration,
	}
}

func (m *Management) CreateAPIKey(ctx context.Context, projectName, serviceAccountName string) (string, *time.Time, error) {
	project, find, err := m.client.GetProject(ctx, projectName)
	if err != nil {
		return "", nil, fmt.Errorf("get project: %w", err)
	}
	if !find {
		project, err = m.client.CreateProject(ctx, projectName)
		if err != nil {
			return "", nil, fmt.Errorf("create project: %w", err)
		}
	}
	serviceAccount, err := m.client.CreateServiceAccount(ctx, project.ID, serviceAccountName)
	if err != nil {
		return "", nil, fmt.Errorf("create service account: %w", err)
	}
	expirationTime := time.Now().Add(m.expiration)
	return serviceAccount.APIKey.Value, &expirationTime, nil
}

func (m *Management) CleanupAPIKey(ctx context.Context, projectName string) error {
	project, find, err := m.client.GetProject(ctx, projectName)
	if err != nil {
		return fmt.Errorf("get project: %w", err)
	}
	if !find {
		return fmt.Errorf("find project %s", projectName)
	}
	serviceAccounts, err := m.client.ListServiceAccounts(ctx, project.ID)
	if err != nil {
		return fmt.Errorf("list service accounts: %w", err)
	}
	for _, serviceAccount := range *serviceAccounts {
		createdAt := time.Unix(serviceAccount.CreatedAt, 0)
		cutoff := time.Now().Add(-1 * m.expiration)
		if createdAt.Before(cutoff) {
			if _, err := m.client.DeleteServiceAccount(ctx, project.ID, serviceAccount.ID); err != nil {
				return fmt.Errorf("delete service account: %w", err)
			}
		}
	}
	return nil
}

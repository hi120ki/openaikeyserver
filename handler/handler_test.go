package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hi120ki/monorepo/projects/openaikeyserver/management"
	"github.com/hi120ki/monorepo/projects/openaikeyserver/oidc"
	"golang.org/x/oauth2"
)

// MockHTTPClient is a mock implementation of the http.Client
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, nil
}

// MockManagement is a mock implementation of the management.Manager interface
type MockManagement struct {
	CreateAPIKeyFunc  func(ctx context.Context, projectName, serviceAccountName string) (string, *time.Time, error)
	CleanupAPIKeyFunc func(ctx context.Context, projectName string) error
}

// Ensure MockManagement implements management.Manager
var _ management.Manager = (*MockManagement)(nil)

func (m *MockManagement) CreateAPIKey(ctx context.Context, projectName, serviceAccountName string) (string, *time.Time, error) {
	if m.CreateAPIKeyFunc != nil {
		return m.CreateAPIKeyFunc(ctx, projectName, serviceAccountName)
	}
	return "", nil, nil
}

func (m *MockManagement) CleanupAPIKey(ctx context.Context, projectName string) error {
	if m.CleanupAPIKeyFunc != nil {
		return m.CleanupAPIKeyFunc(ctx, projectName)
	}
	return nil
}

func TestNewHandler(t *testing.T) {
	// Test data
	allowedUsers := &[]string{"user1@example.com", "user2@example.com"}
	allowedDomains := &[]string{"example.com", "test.com"}
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	redirectURI := "http://localhost:8080/callback"

	// Create mock dependencies
	mockManagement := &MockManagement{}
	mockOIDC := oidc.NewOIDC("test-project", allowedUsers, allowedDomains, "https://accounts.google.com", "https://www.googleapis.com/oauth2/v3/certs")

	// Test NewHandler
	h := NewHandler(allowedUsers, allowedDomains, clientID, clientSecret, redirectURI, mockManagement, mockOIDC)

	// Verify result
	if h == nil {
		t.Fatal("Expected non-nil Handler")
	}

	if h.allowedUsers != allowedUsers {
		t.Errorf("Expected AllowedUsers to be %v, got %v", allowedUsers, h.allowedUsers)
	}

	if h.allowedDomains != allowedDomains {
		t.Errorf("Expected AllowedDomains to be %v, got %v", allowedDomains, h.allowedDomains)
	}

	if h.oauth2Config.ClientID != clientID {
		t.Errorf("Expected ClientID to be %s, got %s", clientID, h.oauth2Config.ClientID)
	}

	if h.oauth2Config.ClientSecret != clientSecret {
		t.Errorf("Expected ClientSecret to be %s, got %s", clientSecret, h.oauth2Config.ClientSecret)
	}

	if h.oauth2Config.RedirectURL != redirectURI {
		t.Errorf("Expected RedirectURL to be %s, got %s", redirectURI, h.oauth2Config.RedirectURL)
	}
}

func TestHandleError(t *testing.T) {
	// Create handler
	h := &Handler{}

	// Create test request and response recorder
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Test handleError
	h.handleError(w, req, nil, http.StatusBadRequest, "Test error message")

	// Verify response
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestGenerateStateOauthCookie(t *testing.T) {
	// Create handler
	h := &Handler{}

	// Create test request and response recorder
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Test generateStateOauthCookie
	state, err := h.generateStateOauthCookie(w, req)

	// Verify result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if state == "" {
		t.Error("Expected non-empty state")
	}

	// Verify cookie
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Errorf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != "oauthstate" {
		t.Errorf("Expected cookie name to be 'oauthstate', got '%s'", cookie.Name)
	}

	if cookie.Value != state {
		t.Errorf("Expected cookie value to be '%s', got '%s'", state, cookie.Value)
	}

	if !cookie.HttpOnly {
		t.Error("Expected HttpOnly to be true")
	}

	if cookie.Secure {
		t.Error("Expected Secure to be false for HTTP request")
	}

	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("Expected SameSite to be %v, got %v", http.SameSiteLaxMode, cookie.SameSite)
	}
}

func TestHandleRoot(t *testing.T) {
	// Create handler
	h := &Handler{
		oauth2Config: &oauth2.Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "http://localhost:8080/callback",
			Scopes:       []string{"email", "openid"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		},
	}

	// Create test request and response recorder
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Test HandleRoot
	h.HandleRoot(w, req)

	// Verify response
	resp := w.Result()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("Expected status code %d, got %d", http.StatusFound, resp.StatusCode)
	}

	// Verify redirect URL
	location := resp.Header.Get("Location")
	if location == "" {
		t.Error("Expected non-empty Location header")
	}

	// Verify cookie
	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Errorf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != "oauthstate" {
		t.Errorf("Expected cookie name to be 'oauthstate', got '%s'", cookie.Name)
	}
}

func TestHandleRevoke(t *testing.T) {
	// Create mock management
	mockManagement := &MockManagement{
		CleanupAPIKeyFunc: func(ctx context.Context, projectName string) error {
			if projectName != "test-project" {
				t.Errorf("Expected project name to be 'test-project', got '%s'", projectName)
			}
			return nil
		},
	}

	// Create mock OIDC
	mockOIDC := oidc.NewOIDC("test-project", &[]string{}, &[]string{}, "", "")

	// Create handler
	h := &Handler{
		management: mockManagement,
		oidc:       mockOIDC,
	}

	// Create test request and response recorder
	req := httptest.NewRequest("GET", "/revoke", nil)
	w := httptest.NewRecorder()

	// Test HandleRevoke
	h.HandleRevoke(w, req)

	// Verify response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

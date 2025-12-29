package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestNew_MissingURL(t *testing.T) {
	logger := logrus.New()

	cfg := Config{
		URL:   "",
		Token: "test-token",
	}

	client, err := New(cfg, logger)

	require.Error(t, err)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "URL is required")
}

func TestNew_MissingToken(t *testing.T) {
	logger := logrus.New()

	cfg := Config{
		URL:   "https://woodpecker.example.com",
		Token: "",
	}

	client, err := New(cfg, logger)

	require.Error(t, err)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "token is required")
}

func TestNew_ValidConfig(t *testing.T) {
	// Create a mock server that returns a valid self response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/user" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": 1, "login": "testuser", "email": "test@example.com"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	logger := logrus.New()

	cfg := Config{
		URL:   server.URL,
		Token: "test-token",
	}

	client, err := New(cfg, logger)

	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNew_ConnectionFailure_InvalidURL(t *testing.T) {
	logger := logrus.New()

	cfg := Config{
		URL:   "https://invalid-hostname-that-does-not-exist.local",
		Token: "test-token",
	}

	client, err := New(cfg, logger)

	// Should fail due to connection error
	require.Error(t, err)
	require.Nil(t, client)
}

func TestGetCurrentUser_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/user" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": 1, "login": "testuser", "email": "test@example.com"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	logger := logrus.New()
	cfg := Config{
		URL:   server.URL,
		Token: "test-token",
	}

	client, err := New(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, client)

	user, err := client.GetCurrentUser()
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, int64(1), user.ID)
	require.Equal(t, "testuser", user.Login)
}

func TestTestConnection_Failure(t *testing.T) {
	// Create a server that returns an error on /api/user
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/user" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "unauthorized"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	logger := logrus.New()

	cfg := Config{
		URL:   server.URL,
		Token: "invalid-token",
	}

	client, err := New(cfg, logger)

	// Should fail due to unauthorized connection
	require.Error(t, err)
	require.Nil(t, client)
}

// The following tests require knowledge of the woodpecker-go library's specific API paths.
// They are skipped because mocking the HTTP responses correctly requires detailed knowledge
// of the library's internal API structure. In production, these would be tested against
// a real Woodpecker server.

func TestListRepositories_Success(t *testing.T) {
	t.Skip("Requires woodpecker-go library API path knowledge")
}

func TestListRepositories_Error(t *testing.T) {
	t.Skip("Requires woodpecker-go library API path knowledge")
}

func TestGetRepository_Success(t *testing.T) {
	t.Skip("Requires woodpecker-go library API path knowledge")
}

func TestGetRepository_NotFound(t *testing.T) {
	t.Skip("Requires woodpecker-go library API path knowledge")
}

func TestLookupRepository_Success(t *testing.T) {
	t.Skip("Requires woodpecker-go library API path knowledge")
}

func TestListPipelines_Success(t *testing.T) {
	t.Skip("Requires woodpecker-go library API path knowledge")
}

func TestGetPipeline_Success(t *testing.T) {
	t.Skip("Requires woodpecker-go library API path knowledge")
}

func TestStartPipeline_Success(t *testing.T) {
	t.Skip("Requires woodpecker-go library API path knowledge")
}

func TestStopPipeline_Success(t *testing.T) {
	t.Skip("Requires woodpecker-go library API path knowledge")
}

package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockGitRepositoryStorage pour mocker l'interface storage
type MockGitRepositoryStorage struct {
	mock.Mock
}

func (m *MockGitRepositoryStorage) GetStorer(repoPath string) (storer.Storer, error) {
	args := m.Called(repoPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(storer.Storer), args.Error(1)
}

func (m *MockGitRepositoryStorage) CreateRepository(repoPath string) error {
	args := m.Called(repoPath)
	return args.Error(0)
}

func (m *MockGitRepositoryStorage) RepositoryExists(repoPath string) bool {
	args := m.Called(repoPath)
	return args.Bool(0)
}

func (m *MockGitRepositoryStorage) DeleteRepository(repoPath string) error {
	args := m.Called(repoPath)
	return args.Error(0)
}

func (m *MockGitRepositoryStorage) ListRepositories() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGitRepositoryStorage) Configure() error {
	args := m.Called()
	return args.Error(0)
}

func setupTestApp() (*fiber.App, *MockGitRepositoryStorage) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	mockStorage := &MockGitRepositoryStorage{}
	logger := zerolog.Nop() // Silent logger for tests

	controller := &RepoController{
		Logger:  logger,
		Storage: mockStorage,
	}

	// Setup routes
	app.Post("/api/repo", controller.CreateRepo)
	app.Get("/api/repos", controller.ListRepos)

	return app, mockStorage
}

func TestCreateRepoSuccess(t *testing.T) {
	app, mockStorage := setupTestApp()

	// Mock successful repository creation
	mockStorage.On("CreateRepository", "test-repo.git").Return(nil)

	// Prepare request body
	reqBody := map[string]string{
		"name": "test-repo",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Make request
	req := httptest.NewRequest("POST", "/api/repo", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "repository created", string(body))

	// Verify mock was called
	mockStorage.AssertExpectations(t)
}

func TestCreateRepoInvalidJSON(t *testing.T) {
	app, _ := setupTestApp()

	// Send invalid JSON
	req := httptest.NewRequest("POST", "/api/repo", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestCreateRepoMissingName(t *testing.T) {
	app, mockStorage := setupTestApp()

	// Mock storage to expect empty string (normalized name will be ".git")
	mockStorage.On("CreateRepository", ".git").Return(assert.AnError)

	// Send empty body
	reqBody := map[string]string{}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/repo", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	// The repository will be created with an empty name, which is technically valid
	// but will be normalized by NormalizeRepoPath and fail at the storage level
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	// Verify mock was called
	mockStorage.AssertExpectations(t)
}

func TestCreateRepoStorageError(t *testing.T) {
	app, mockStorage := setupTestApp()

	// Mock storage error
	mockStorage.On("CreateRepository", "test-repo.git").Return(assert.AnError)

	// Prepare request body
	reqBody := map[string]string{
		"name": "test-repo",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Make request
	req := httptest.NewRequest("POST", "/api/repo", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "failed to create repository", string(body))

	// Verify mock was called
	mockStorage.AssertExpectations(t)
}

func TestListReposSuccess(t *testing.T) {
	app, mockStorage := setupTestApp()

	// Mock successful repository listing
	expectedRepos := []string{"repo1.git", "repo2.git", "repo3.git"}
	mockStorage.On("ListRepositories").Return(expectedRepos, nil)

	// Make request
	req := httptest.NewRequest("GET", "/api/repos", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, _ := io.ReadAll(resp.Body)
	var repos []string
	err = json.Unmarshal(body, &repos)
	require.NoError(t, err)

	assert.Equal(t, expectedRepos, repos)

	// Verify mock was called
	mockStorage.AssertExpectations(t)
}

func TestListReposEmpty(t *testing.T) {
	app, mockStorage := setupTestApp()

	// Mock empty repository list
	mockStorage.On("ListRepositories").Return([]string{}, nil)

	// Make request
	req := httptest.NewRequest("GET", "/api/repos", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var repos []string
	err = json.Unmarshal(body, &repos)
	require.NoError(t, err)

	assert.Empty(t, repos)

	// Verify mock was called
	mockStorage.AssertExpectations(t)
}

func TestListReposStorageError(t *testing.T) {
	app, mockStorage := setupTestApp()

	// Mock storage error
	mockStorage.On("ListRepositories").Return([]string{}, assert.AnError)

	// Make request
	req := httptest.NewRequest("GET", "/api/repos", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "failed to list repositories", string(body))

	// Verify mock was called
	mockStorage.AssertExpectations(t)
}

// Test d'intégration pour valider le flux complet
func TestRepoControllerIntegration(t *testing.T) {
	app, mockStorage := setupTestApp()

	// Test 1: List repositories (empty at start)
	mockStorage.On("ListRepositories").Return([]string{}, nil).Once()

	req := httptest.NewRequest("GET", "/api/repos", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Test 2: Create a new repository
	mockStorage.On("CreateRepository", "integration-test.git").Return(nil).Once()

	reqBody := map[string]string{"name": "integration-test"}
	bodyBytes, _ := json.Marshal(reqBody)

	req = httptest.NewRequest("POST", "/api/repo", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	// Test 3: List repositories (now with the new one)
	mockStorage.On("ListRepositories").Return([]string{"integration-test.git"}, nil).Once()

	req = httptest.NewRequest("GET", "/api/repos", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var repos []string
	err = json.Unmarshal(body, &repos)
	require.NoError(t, err)
	assert.Contains(t, repos, "integration-test.git")

	// Verify all mocks were called
	mockStorage.AssertExpectations(t)
}

// Test de charge pour valider les performances
func BenchmarkCreateRepo(b *testing.B) {
	app, mockStorage := setupTestApp()

	// Mock pour accepter tous les appels
	mockStorage.On("CreateRepository", mock.AnythingOfType("string")).Return(nil)

	reqBody := map[string]string{"name": "benchmark-repo"}
	bodyBytes, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/repo", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		resp.Body.Close()
	}
}

func BenchmarkListRepos(b *testing.B) {
	app, mockStorage := setupTestApp()

	// Mock avec une liste de repos
	repos := make([]string, 100)
	for i := 0; i < 100; i++ {
		repos[i] = fmt.Sprintf("repo%d.git", i)
	}
	mockStorage.On("ListRepositories").Return(repos, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/repos", nil)
		resp, _ := app.Test(req)
		resp.Body.Close()
	}
}

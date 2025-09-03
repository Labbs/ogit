package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/labbs/git-server-s3/internal/api/router"
	"github.com/labbs/git-server-s3/internal/config"
	"github.com/labbs/git-server-s3/pkg/storage/local"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLocalStorageIntegration tests the API with local storage to avoid S3 complications
func TestLocalStorageIntegration(t *testing.T) {
	// Create a temporary directory for tests
	tempDir, err := os.MkdirTemp("", "git-server-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test configuration with local storage
	config.Storage.Type = "local"
	config.Storage.Local.Path = tempDir

	// Silent logger for tests
	logger := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel)

	// Initialize local storage
	localStorage := local.NewLocalStorage(logger)
	err = localStorage.Configure()
	require.NoError(t, err)

	// Create the Fiber app for tests
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Configure routes
	routerConfig := &router.Config{
		Fiber:   app,
		Logger:  logger,
		Storage: localStorage,
	}
	router.NewRepoRouter(routerConfig)
	router.NewGitRouter(routerConfig)

	// Test 1: List repositories (empty at start)
	t.Run("list_empty_repositories", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/repos", nil)
		resp, err := app.Test(req, 5*1000) // 5 second timeout
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var repos []string
		err = json.Unmarshal(body, &repos)
		require.NoError(t, err)
		assert.Empty(t, repos)
	})

	// Test 2: Create a new repository
	var testRepoName = "integration-test-repo"
	t.Run("create_repository", func(t *testing.T) {
		reqBody := map[string]string{
			"name": testRepoName,
		}
		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/repo", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5*1000)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "repository created", string(body))
	})

	// Test 3: Verify that the repository exists on the filesystem
	t.Run("verify_repository_created", func(t *testing.T) {
		expectedPath := filepath.Join(tempDir, testRepoName+".git")
		assert.DirExists(t, expectedPath)

		// Verify that it's a bare Git repository
		configPath := filepath.Join(expectedPath, "config")
		assert.FileExists(t, configPath)

		headPath := filepath.Join(expectedPath, "HEAD")
		assert.FileExists(t, headPath)
	})

	// Test 4: List repositories (now with the new one)
	t.Run("list_repositories_with_new_repo", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/repos", nil)
		resp, err := app.Test(req, 5*1000)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var repos []string
		err = json.Unmarshal(body, &repos)
		require.NoError(t, err)

		assert.Len(t, repos, 1)
		assert.Contains(t, repos, testRepoName+".git")
	})

	// Test 5: Test the Git info/refs endpoint
	t.Run("git_info_refs", func(t *testing.T) {
		url := fmt.Sprintf("/%s.git/info/refs?service=git-upload-pack", testRepoName)
		req := httptest.NewRequest("GET", url, nil)

		resp, err := app.Test(req, 5*1000)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/x-git-upload-pack-advertisement", resp.Header.Get("Content-Type"))

		// The body should contain Git protocol data
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.NotEmpty(t, body)
	})

	// Test 6: Create another repository to test multiple repository management
	var secondRepoName = "second-repo"
	t.Run("create_second_repository", func(t *testing.T) {
		reqBody := map[string]string{
			"name": secondRepoName,
		}
		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/repo", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5*1000)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// Test 7: Verify that both repositories appear
	t.Run("list_multiple_repositories", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/repos", nil)
		resp, err := app.Test(req, 5*1000)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var repos []string
		err = json.Unmarshal(body, &repos)
		require.NoError(t, err)

		assert.Len(t, repos, 2)
		assert.Contains(t, repos, testRepoName+".git")
		assert.Contains(t, repos, secondRepoName+".git")
	})
}

// TestErrorCases tests error cases
func TestErrorCases(t *testing.T) {
	// Create a temporary directory for tests
	tempDir, err := os.MkdirTemp("", "git-server-error-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test configuration
	config.Storage.Type = "local"
	config.Storage.Local.Path = tempDir

	logger := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel)
	localStorage := local.NewLocalStorage(logger)
	err = localStorage.Configure()
	require.NoError(t, err)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	routerConfig := &router.Config{
		Fiber:   app,
		Logger:  logger,
		Storage: localStorage,
	}
	router.NewRepoRouter(routerConfig)
	router.NewGitRouter(routerConfig)

	// Test 1: Request with invalid JSON
	t.Run("invalid_json", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/repo", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5*1000)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Test 2: Create the same repository twice
	t.Run("duplicate_repository", func(t *testing.T) {
		repoName := "duplicate-repo"

		// Première création (doit réussir)
		reqBody := map[string]string{"name": repoName}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/repo", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5*1000)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Second creation (should fail)
		req = httptest.NewRequest("POST", "/api/repo", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, 5*1000)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return an error because the repo already exists
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	// Test 3: Access a repository that doesn't exist
	t.Run("nonexistent_repository", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/nonexistent-repo.git/info/refs?service=git-upload-pack", nil)

		resp, err := app.Test(req, 5*1000)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// BenchmarkAPIEndpoints tests the performance of main endpoints
func BenchmarkAPIEndpoints(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "git-server-benchmark-*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	config.Storage.Type = "local"
	config.Storage.Local.Path = tempDir

	logger := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel)
	localStorage := local.NewLocalStorage(logger)
	err = localStorage.Configure()
	require.NoError(b, err)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	routerConfig := &router.Config{
		Fiber:   app,
		Logger:  logger,
		Storage: localStorage,
	}
	router.NewRepoRouter(routerConfig)

	b.Run("ListRepositories", func(b *testing.B) {
		// Create a few repositories for the benchmark
		for i := 0; i < 10; i++ {
			reqBody := map[string]string{"name": fmt.Sprintf("benchmark-repo-%d", i)}
			bodyBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/api/repo", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req)
			resp.Body.Close()
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/api/repos", nil)
			resp, _ := app.Test(req)
			resp.Body.Close()
		}
	})

	b.Run("CreateRepository", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			reqBody := map[string]string{"name": fmt.Sprintf("bench-repo-%d-%d", time.Now().UnixNano(), i)}
			bodyBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/api/repo", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req)
			resp.Body.Close()
		}
	})
}

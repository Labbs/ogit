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

// TestLocalStorageIntegration teste l'API avec le storage local pour éviter les complications S3
func TestLocalStorageIntegration(t *testing.T) {
	// Créer un répertoire temporaire pour les tests
	tempDir, err := os.MkdirTemp("", "git-server-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Configuration de test avec storage local
	config.Storage.Type = "local"
	config.Storage.Local.Path = tempDir

	// Logger silencieux pour les tests
	logger := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel)

	// Initialiser le storage local
	localStorage := local.NewLocalStorage(logger)
	err = localStorage.Configure()
	require.NoError(t, err)

	// Créer l'app Fiber pour les tests
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Configuration des routes
	routerConfig := &router.Config{
		Fiber:   app,
		Logger:  logger,
		Storage: localStorage,
	}
	router.NewRepoRouter(routerConfig)
	router.NewGitRouter(routerConfig)

	// Test 1: Lister les repositories (vide au début)
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

	// Test 2: Créer un nouveau repository
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

	// Test 3: Vérifier que le repository existe sur le filesystem
	t.Run("verify_repository_created", func(t *testing.T) {
		expectedPath := filepath.Join(tempDir, testRepoName+".git")
		assert.DirExists(t, expectedPath)

		// Vérifier que c'est un repo Git bare
		configPath := filepath.Join(expectedPath, "config")
		assert.FileExists(t, configPath)

		headPath := filepath.Join(expectedPath, "HEAD")
		assert.FileExists(t, headPath)
	})

	// Test 4: Lister les repositories (maintenant avec le nouveau)
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

	// Test 5: Tester l'endpoint Git info/refs
	t.Run("git_info_refs", func(t *testing.T) {
		url := fmt.Sprintf("/%s.git/info/refs?service=git-upload-pack", testRepoName)
		req := httptest.NewRequest("GET", url, nil)

		resp, err := app.Test(req, 5*1000)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/x-git-upload-pack-advertisement", resp.Header.Get("Content-Type"))

		// Le body devrait contenir des données Git protocol
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.NotEmpty(t, body)
	})

	// Test 6: Créer un autre repository pour tester la gestion multiple
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

	// Test 7: Vérifier que les deux repositories apparaissent
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

// TestErrorCases teste les cas d'erreur
func TestErrorCases(t *testing.T) {
	// Créer un répertoire temporaire pour les tests
	tempDir, err := os.MkdirTemp("", "git-server-error-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Configuration de test
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

	// Test 1: Requête avec JSON invalide
	t.Run("invalid_json", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/repo", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5*1000)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Test 2: Créer deux fois le même repository
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

		// Deuxième création (doit échouer)
		req = httptest.NewRequest("POST", "/api/repo", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, 5*1000)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Devrait retourner une erreur car le repo existe déjà
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	// Test 3: Accès à un repository qui n'existe pas
	t.Run("nonexistent_repository", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/nonexistent-repo.git/info/refs?service=git-upload-pack", nil)

		resp, err := app.Test(req, 5*1000)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// BenchmarkAPIEndpoints teste les performances des endpoints principaux
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
		// Créer quelques repos pour le benchmark
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

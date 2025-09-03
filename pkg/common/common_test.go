package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractRepoPathFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		suffix   string
		expected string
	}{
		{
			name:     "simple git URL",
			url:      "/test-repo.git/info/refs",
			suffix:   "/info/refs",
			expected: "test-repo.git",
		},
		{
			name:     "nested path",
			url:      "/organization/project.git/git-upload-pack",
			suffix:   "/git-upload-pack",
			expected: "organization/project.git",
		},
		{
			name:     "root path",
			url:      "/repo.git/info/refs",
			suffix:   "/info/refs",
			expected: "repo.git",
		},
		{
			name:     "no suffix match",
			url:      "/test-repo.git/other/path",
			suffix:   "/info/refs",
			expected: "",
		},
		{
			name:     "empty URL",
			url:      "",
			suffix:   "/info/refs",
			expected: "",
		},
		{
			name:     "URL without .git",
			url:      "/test-repo/info/refs",
			suffix:   "/info/refs",
			expected: "test-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractRepoPathFromURL(tt.url, tt.suffix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeRepoPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple repo name",
			input:    "test-repo",
			expected: "test-repo.git",
		},
		{
			name:     "repo with .git extension",
			input:    "test-repo.git",
			expected: "test-repo.git",
		},
		{
			name:     "repo with leading slash",
			input:    "/test-repo",
			expected: "test-repo.git",
		},
		{
			name:     "repo with trailing slash",
			input:    "test-repo/",
			expected: "test-repo.git",
		},
		{
			name:     "repo with both slashes",
			input:    "/test-repo/",
			expected: "test-repo.git",
		},
		{
			name:     "organization/project format",
			input:    "organization/project",
			expected: "organization/project.git",
		},
		{
			name:     "organization/project with .git",
			input:    "organization/project.git",
			expected: "organization/project.git",
		},
		{
			name:     "empty input",
			input:    "",
			expected: ".git",
		},
		{
			name:     "only slashes",
			input:    "///",
			expected: ".git",
		},
		{
			name:     "complex path",
			input:    "/company/team/project/",
			expected: "company/team/project.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeRepoPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests pour valider des cas d'usage réels
func TestNormalizeRepoPathRealExamples(t *testing.T) {
	// Cas d'usage typiques qu'on peut rencontrer
	examples := map[string]string{
		"my-awesome-project":       "my-awesome-project.git",
		"frontend-app":             "frontend-app.git",
		"backend-api.git":          "backend-api.git",
		"company/mobile-app":       "company/mobile-app.git",
		"github.com/user/repo":     "github.com/user/repo.git",
		"/tmp/test-repo":           "tmp/test-repo.git",
		"project-with-dashes":      "project-with-dashes.git",
		"project_with_underscores": "project_with_underscores.git",
		"ProjectWithCamelCase":     "ProjectWithCamelCase.git",
		"project.with.dots":        "project.with.dots.git",
		"123-numeric-start":        "123-numeric-start.git",
	}

	for input, expected := range examples {
		t.Run("normalize_"+input, func(t *testing.T) {
			result := NormalizeRepoPath(input)
			assert.Equal(t, expected, result)
		})
	}
}

func TestExtractRepoPathRealExamples(t *testing.T) {
	// Cas d'usage typiques pour l'extraction de chemin
	examples := []struct {
		url      string
		suffix   string
		expected string
	}{
		{
			url:      "/my-project.git/info/refs",
			suffix:   "/info/refs",
			expected: "my-project.git",
		},
		{
			url:      "/company/mobile-app.git/git-upload-pack",
			suffix:   "/git-upload-pack",
			expected: "company/mobile-app.git",
		},
		{
			url:      "/github.com/user/repo.git/git-receive-pack",
			suffix:   "/git-receive-pack",
			expected: "github.com/user/repo.git",
		},
		{
			url:      "/deep/nested/path/project.git/info/refs",
			suffix:   "/info/refs",
			expected: "deep/nested/path/project.git",
		},
	}

	for _, example := range examples {
		t.Run("extract_"+example.expected, func(t *testing.T) {
			result := ExtractRepoPathFromURL(example.url, example.suffix)
			assert.Equal(t, example.expected, result)
		})
	}
}

// Tests de performance pour s'assurer que les fonctions sont rapides
func BenchmarkNormalizeRepoPath(b *testing.B) {
	testInput := "organization/very-long-project-name-with-many-characters"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NormalizeRepoPath(testInput)
	}
}

func BenchmarkExtractRepoPathFromURL(b *testing.B) {
	testURL := "/organization/very-long-project-name.git/info/refs"
	testSuffix := "/info/refs"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractRepoPathFromURL(testURL, testSuffix)
	}
}

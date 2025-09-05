// Package common provides utility functions shared across the Git server.
package common

import "strings"

// ExtractRepoPathFromURL extracts the repository path from a Git Smart HTTP URL.
// It removes the specified suffix and any leading/trailing slashes to get the clean repository path.
//
// Parameters:
//   - urlPath: The full URL path from the HTTP request (e.g., "/myrepo.git/info/refs")
//   - suffix: The suffix to remove (e.g., "/info/refs", "/git-upload-pack")
//
// Returns:
//   - The clean repository path (e.g., "myrepo.git") or empty string if invalid
//
// Example:
//
//	ExtractRepoPathFromURL("/myrepo.git/info/refs", "/info/refs") → "myrepo.git"
//	ExtractRepoPathFromURL("/path/to/repo.git/git-upload-pack", "/git-upload-pack") → "path/to/repo.git"
func ExtractRepoPathFromURL(urlPath, suffix string) string {
	// Find the last occurrence of the suffix in the URL path
	idx := strings.LastIndex(urlPath, suffix)
	if idx < 0 {
		return ""
	}

	// Extract the base path by removing the suffix and trailing slash
	base := strings.TrimSuffix(urlPath[:idx], "/")

	// Remove leading slash to get relative path
	rel := strings.TrimPrefix(base, "/")
	if rel == "" {
		return ""
	}

	return rel
}

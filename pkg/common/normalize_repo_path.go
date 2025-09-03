package common

import "strings"

// NormalizeRepoPath normalizes a repository path to ensure proper Git bare repository naming.
// It trims whitespace and ensures the path ends with ".git" suffix, which is the standard
// convention for bare Git repositories.
//
// Parameters:
//   - repoPath: The raw repository path/name from user input
//
// Returns:
//   - The normalized repository path with ".git" suffix
//
// Examples:
//
//	NormalizeRepoPath("myrepo") → "myrepo.git"
//	NormalizeRepoPath("myrepo.git") → "myrepo.git"
//	NormalizeRepoPath("  myrepo  ") → "myrepo.git"
func NormalizeRepoPath(repoPath string) string {
	// Remove leading and trailing whitespace
	repoPath = strings.TrimSpace(repoPath)

	// Ensure the repository path ends with .git suffix
	if !strings.HasSuffix(repoPath, ".git") {
		repoPath += ".git"
	}

	return repoPath
}

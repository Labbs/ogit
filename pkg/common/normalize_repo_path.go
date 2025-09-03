package common

import "strings"

func NormalizeRepoPath(repoPath string) string {
	repoPath = strings.TrimSpace(repoPath)
	if !strings.HasSuffix(repoPath, ".git") {
		repoPath += ".git"
	}
	return repoPath
}

package common

import "strings"

func ExtractRepoPathFromURL(urlPath, suffix string) string {
	idx := strings.LastIndex(urlPath, suffix)
	if idx < 0 {
		return ""
	}
	base := strings.TrimSuffix(urlPath[:idx], "/")
	rel := strings.TrimPrefix(base, "/")
	if rel == "" {
		return ""
	}
	return rel
}

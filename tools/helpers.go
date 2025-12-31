package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/denysvitali/woodpecker-ci-mcp/internal/client"
)

// Helper functions for argument extraction

// getRepoID resolves a repository ID from arguments.
// It tries (in order):
// 1. repo_id from arguments
// 2. repo_name from arguments (looks up the repository)
// 3. git remote inference (if neither repo_id nor repo_name is provided)
func getRepoID(wclient *client.Client, arguments map[string]interface{}) (int64, error) {
	// Try repo_id first
	if repoID, ok := arguments["repo_id"]; ok {
		if repoIDFloat, ok := repoID.(float64); ok {
			return int64(repoIDFloat), nil
		}
		return 0, fmt.Errorf("repo_id must be a number")
	}

	// Try repo_name second
	if repoName, ok := arguments["repo_name"]; ok {
		if repoNameStr, ok := repoName.(string); ok {
			repo, err := wclient.LookupRepository(repoNameStr)
			if err != nil {
				return 0, fmt.Errorf("failed to lookup repository: %w", err)
			}
			return repo.ID, nil
		}
		return 0, fmt.Errorf("repo_name must be a string")
	}

	// Try to infer from git remote as last resort
	repoName, err := getRepoNameFromRemote()
	if err == nil {
		repo, lookupErr := wclient.LookupRepository(repoName)
		if lookupErr == nil {
			return repo.ID, nil
		}
		return 0, fmt.Errorf("failed to lookup inferred repository %s: %w", repoName, lookupErr)
	}

	return 0, fmt.Errorf("either repo_id, repo_name must be provided, or git remote must be available")
}

// getBool returns the boolean value for a key, or defaultValue if not present or invalid type
func getBool(arguments map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := arguments[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

// getString returns the string value for a key, or defaultValue if not present or invalid type
func getString(arguments map[string]interface{}, key string, defaultValue string) string {
	if val, ok := arguments[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// getNumber returns the number value for a key, or defaultValue if not present or invalid type
func getNumber(arguments map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := arguments[key]; ok {
		if numVal, ok := val.(float64); ok {
			return numVal
		}
	}
	return defaultValue
}

// requireNumber returns the number value for a key, or an error if not present or invalid type
func requireNumber(arguments map[string]interface{}, key string) (float64, error) {
	if val, ok := arguments[key]; ok {
		if numVal, ok := val.(float64); ok {
			return numVal, nil
		}
		return 0, fmt.Errorf("%s must be a number", key)
	}
	return 0, fmt.Errorf("%s is required", key)
}

// checkContextCancelled returns a cancellation error result if the context is done
func checkContextCancelled(ctx context.Context) *mcp.CallToolResult {
	select {
	case <-ctx.Done():
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "request cancelled",
				},
			},
			IsError: true,
		}
	default:
		return nil
	}
}

// getRepoNameFromRemote infers the repository name from git remote
// It strips the github.com part and returns the rest (e.g., foo/bar)
func getRepoNameFromRemote() (string, error) {
	// Get the remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))

	// Remove common git prefixes/suffixes
	remoteURL = strings.TrimPrefix(remoteURL, "https://")
	remoteURL = strings.TrimPrefix(remoteURL, "http://")
	remoteURL = strings.TrimPrefix(remoteURL, "git@")
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	// Handle SSH format (git@github.com:foo/bar)
	if strings.Contains(remoteURL, ":") {
		parts := strings.SplitN(remoteURL, ":", 2)
		if len(parts) == 2 {
			remoteURL = parts[1]
		}
	}

	// Handle HTTPS format (github.com/foo/bar)
	parts := strings.Split(remoteURL, "/")
	if len(parts) >= 2 {
		// Find the part after the domain
		for i, part := range parts {
			if strings.Contains(part, ".") && i+1 < len(parts) {
				// This looks like a domain, take the rest
				return strings.Join(parts[i+1:], "/"), nil
			}
		}
		// If no domain found, assume the last 2 parts are owner/repo
		if len(parts) >= 2 {
			return strings.Join(parts[len(parts)-2:], "/"), nil
		}
	}

	return "", fmt.Errorf("could not parse repository name from remote URL: %s", remoteURL)
}

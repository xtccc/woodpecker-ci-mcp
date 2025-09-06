package tools

import (
	"fmt"

	"github.com/denysvitali/woodpecker-ci-mcp/internal/client"
)

// Helper functions for argument extraction

func getRepoID(wclient *client.Client, arguments map[string]interface{}) (int64, error) {
	if repoID, ok := arguments["repo_id"]; ok {
		if repoIDFloat, ok := repoID.(float64); ok {
			return int64(repoIDFloat), nil
		}
		return 0, fmt.Errorf("repo_id must be a number")
	}

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

	return 0, fmt.Errorf("either repo_id or repo_name must be provided")
}

func getBool(arguments map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := arguments[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

func getString(arguments map[string]interface{}, key string, defaultValue string) string {
	if val, ok := arguments[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func getNumber(arguments map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := arguments[key]; ok {
		if numVal, ok := val.(float64); ok {
			return numVal
		}
	}
	return defaultValue
}

func requireNumber(arguments map[string]interface{}, key string) (float64, error) {
	if val, ok := arguments[key]; ok {
		if numVal, ok := val.(float64); ok {
			return numVal, nil
		}
		return 0, fmt.Errorf("%s must be a number", key)
	}
	return 0, fmt.Errorf("%s is required", key)
}

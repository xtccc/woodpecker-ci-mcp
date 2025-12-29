package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetBool_WithValidBool(t *testing.T) {
	arguments := map[string]interface{}{
		"enabled": true,
	}

	result := getBool(arguments, "enabled", false)
	require.True(t, result)
}

func TestGetBool_WithMissingKey_ReturnsDefault(t *testing.T) {
	arguments := map[string]interface{}{}

	result := getBool(arguments, "missing", true)
	require.True(t, result)
}

func TestGetBool_WithWrongType_ReturnsDefault(t *testing.T) {
	arguments := map[string]interface{}{
		"enabled": "true",
	}

	result := getBool(arguments, "enabled", false)
	require.False(t, result)
}

func TestGetBool_WithFalseValue(t *testing.T) {
	arguments := map[string]interface{}{
		"enabled": false,
	}

	result := getBool(arguments, "enabled", true)
	require.False(t, result)
}

func TestGetString_WithValidString(t *testing.T) {
	arguments := map[string]interface{}{
		"name": "test-value",
	}

	result := getString(arguments, "name", "default")
	require.Equal(t, "test-value", result)
}

func TestGetString_WithMissingKey_ReturnsDefault(t *testing.T) {
	arguments := map[string]interface{}{}

	result := getString(arguments, "missing", "default")
	require.Equal(t, "default", result)
}

func TestGetString_WithWrongType_ReturnsDefault(t *testing.T) {
	arguments := map[string]interface{}{
		"name": float64(123),
	}

	result := getString(arguments, "name", "default")
	require.Equal(t, "default", result)
}

func TestGetString_WithEmptyString(t *testing.T) {
	arguments := map[string]interface{}{
		"name": "",
	}

	result := getString(arguments, "name", "default")
	require.Equal(t, "", result)
}

func TestGetNumber_WithValidNumber(t *testing.T) {
	arguments := map[string]interface{}{
		"count": float64(42.5),
	}

	result := getNumber(arguments, "count", 0)
	require.Equal(t, float64(42.5), result)
}

func TestGetNumber_WithMissingKey_ReturnsDefault(t *testing.T) {
	arguments := map[string]interface{}{}

	result := getNumber(arguments, "missing", 10.0)
	require.Equal(t, float64(10.0), result)
}

func TestGetNumber_WithWrongType_ReturnsDefault(t *testing.T) {
	arguments := map[string]interface{}{
		"count": "42",
	}

	result := getNumber(arguments, "count", 0)
	require.Equal(t, float64(0), result)
}

func TestGetNumber_WithIntegerFloat(t *testing.T) {
	arguments := map[string]interface{}{
		"count": float64(100),
	}

	result := getNumber(arguments, "count", 0)
	require.Equal(t, float64(100), result)
}

func TestRequireNumber_WithValidNumber(t *testing.T) {
	arguments := map[string]interface{}{
		"value": float64(42),
	}

	result, err := requireNumber(arguments, "value")

	require.NoError(t, err)
	require.Equal(t, float64(42), result)
}

func TestRequireNumber_WithMissingKey_ReturnsError(t *testing.T) {
	arguments := map[string]interface{}{}

	result, err := requireNumber(arguments, "missing")

	require.Error(t, err)
	require.Equal(t, float64(0), result)
	require.Contains(t, err.Error(), "missing is required")
}

func TestRequireNumber_WithWrongType_ReturnsError(t *testing.T) {
	arguments := map[string]interface{}{
		"value": "not-a-number",
	}

	result, err := requireNumber(arguments, "value")

	require.Error(t, err)
	require.Equal(t, float64(0), result)
	require.Contains(t, err.Error(), "value must be a number")
}

func TestRequireNumber_WithZero(t *testing.T) {
	arguments := map[string]interface{}{
		"value": float64(0),
	}

	result, err := requireNumber(arguments, "value")

	require.NoError(t, err)
	require.Equal(t, float64(0), result)
}

// TestGetRepoID_RequiresClient tests that getRepoID requires a valid client.Client
// We cannot test this without integration with the actual woodpecker-go client
func TestGetRepoID_RequiresClient(t *testing.T) {
	// This test documents that getRepoID requires a *client.Client
	// For proper testing, use the client_test.go with httptest
	t.Skip("getRepoID requires *client.Client - tested via client integration tests")
}

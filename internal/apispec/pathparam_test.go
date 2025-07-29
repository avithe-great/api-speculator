package apispec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnifyParameterizedPathIfApplicable(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		isSpec   bool
		expected string
	}{
		{
			name:     "static path - no change",
			input:    "/users/list",
			isSpec:   false,
			expected: "/users/list",
		},
		{
			name:     "numeric ID in path",
			input:    "/users/123",
			isSpec:   false,
			expected: "/users/{param1}",
		},
		{
			name:     "UUID in path",
			input:    "/orders/550e8400-e29b-41d4-a716-446655440000",
			isSpec:   false,
			expected: "/orders/{param1}",
		},
		{
			name:     "mixed alphanumeric long string",
			input:    "/data/abc12345xyz",
			isSpec:   false,
			expected: "/data/{param1}",
		},
		{
			name:     "short alphanumeric (not treated as param)",
			input:    "/data/ab12",
			isSpec:   false,
			expected: "/data/ab12",
		},
		{
			name:     "multiple parameters in path",
			input:    "/users/123/orders/550e8400-e29b-41d4-a716-446655440000",
			isSpec:   false,
			expected: "/users/{param1}/orders/{param2}",
		},
		{
			name:     "already parameterized spec path",
			input:    "/users/{userId}",
			isSpec:   true,
			expected: "/users/{userId}",
		},
		{
			name:     "spec path with multiple params",
			input:    "/users/{userId}/orders/{orderId}",
			isSpec:   true,
			expected: "/users/{userId}/orders/{orderId}",
		},
		{
			name:     "root path",
			input:    "/",
			isSpec:   false,
			expected: "/",
		},
		{
			name:     "empty path",
			input:    "",
			isSpec:   false,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnifyParameterizedPathIfApplicable(tt.input, tt.isSpec)
			assert.Equal(t, tt.expected, result)
		})
	}
}

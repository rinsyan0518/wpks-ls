package lsp

import (
	"testing"
)

func TestParseInitializationOptions(t *testing.T) {
	tests := []struct {
		name                  string
		initializationOptions any
		expected              ServerOptions
	}{
		{
			name:                  "nil input",
			initializationOptions: nil,
			expected: ServerOptions{
				CheckAllOnInitialized: false,
			},
		},
		{
			name:                  "empty map",
			initializationOptions: map[string]any{},
			expected: ServerOptions{
				CheckAllOnInitialized: false,
			},
		},
		{
			name: "checkAllOnInitialized is true",
			initializationOptions: map[string]any{
				"checkAllOnInitialized": true,
			},
			expected: ServerOptions{
				CheckAllOnInitialized: true,
			},
		},
		{
			name: "checkAllOnInitialized is false",
			initializationOptions: map[string]any{
				"checkAllOnInitialized": false,
			},
			expected: ServerOptions{
				CheckAllOnInitialized: false,
			},
		},
		{
			name: "checkAllOnInitialized is not bool type",
			initializationOptions: map[string]any{
				"checkAllOnInitialized": "true", // string instead of bool
			},
			expected: ServerOptions{
				CheckAllOnInitialized: false, // should use default value
			},
		},
		{
			name: "checkAllOnInitialized key does not exist",
			initializationOptions: map[string]any{
				"someOtherKey": "value",
			},
			expected: ServerOptions{
				CheckAllOnInitialized: false,
			},
		},
		{
			name: "multiple fields with checkAllOnInitialized true",
			initializationOptions: map[string]any{
				"checkAllOnInitialized": true,
				"otherField":            "ignored",
				"anotherField":          123,
			},
			expected: ServerOptions{
				CheckAllOnInitialized: true,
			},
		},
		{
			name: "multiple fields with checkAllOnInitialized false",
			initializationOptions: map[string]any{
				"checkAllOnInitialized": false,
				"otherField":            "ignored",
				"anotherField":          123,
			},
			expected: ServerOptions{
				CheckAllOnInitialized: false,
			},
		},
		{
			name:                  "input is not a map",
			initializationOptions: "not a map",
			expected: ServerOptions{
				CheckAllOnInitialized: false,
			},
		},
		{
			name:                  "input is number",
			initializationOptions: 42,
			expected: ServerOptions{
				CheckAllOnInitialized: false,
			},
		},
		{
			name:                  "input is slice",
			initializationOptions: []string{"a", "b", "c"},
			expected: ServerOptions{
				CheckAllOnInitialized: false,
			},
		},
		{
			name: "input is map[string]interface{} (equivalent to map[string]any)",
			initializationOptions: map[string]interface{}{
				"checkAllOnInitialized": true,
			},
			expected: ServerOptions{
				CheckAllOnInitialized: true, // map[string]interface{} is equivalent to map[string]any
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseInitializationOptions(tt.initializationOptions)

			if result.CheckAllOnInitialized != tt.expected.CheckAllOnInitialized {
				t.Errorf("ParseInitializationOptions() CheckAllOnInitialized = %v, expected %v",
					result.CheckAllOnInitialized, tt.expected.CheckAllOnInitialized)
			}
		})
	}
}

// Benchmark test to ensure performance is reasonable
func BenchmarkParseInitializationOptions(b *testing.B) {
	input := map[string]any{
		"checkAllOnInitialized": true,
		"otherField":            "value",
		"anotherField":          123,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseInitializationOptions(input)
	}
}

func BenchmarkParseInitializationOptionsNil(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseInitializationOptions(nil)
	}
}

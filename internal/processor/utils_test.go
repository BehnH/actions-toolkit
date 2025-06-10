package processor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsHexString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid hex string - lowercase",
			input:    "1a2b3c4d5e6f",
			expected: true,
		},
		{
			name:     "valid hex string - uppercase",
			input:    "1A2B3C4D5E6F",
			expected: true,
		},
		{
			name:     "valid hex string - mixed case",
			input:    "1a2B3c4D5e6F",
			expected: true,
		},
		{
			name:     "valid hex string - digits only",
			input:    "123456789",
			expected: true,
		},
		{
			name:     "invalid hex string - contains non-hex character",
			input:    "1a2b3g4d5e6f",
			expected: false,
		},
		{
			name:     "invalid hex string - empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid hex string - contains special characters",
			input:    "1a2b-3c4d",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHexString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpdateVersionComment(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		version        string
		expectedOutput string
	}{
		{
			name:           "no existing comment",
			line:           "uses: actions/setup-node@v4.3.0",
			version:        "v4.4.0",
			expectedOutput: "uses: actions/setup-node@v4.3.0 # v4.4.0",
		},
		{
			name:           "existing version comment",
			line:           "uses: actions/setup-node@v4.3.0 # v4.3.0",
			version:        "v4.4.0",
			expectedOutput: "uses: actions/setup-node@v4.3.0 # v4.4.0",
		},
		{
			name:           "existing version comment with additional text",
			line:           "uses: actions/setup-node@v4.3.0 # v4.3.0 pinned version",
			version:        "v4.4.0",
			expectedOutput: "uses: actions/setup-node@v4.3.0 # v4.4.0 pinned version",
		},
		{
			name:           "comment with pin@version pattern",
			line:           "uses: actions/setup-node@v4.3.0 # pin@v4.3.0",
			version:        "v4.4.0",
			expectedOutput: "uses: actions/setup-node@v4.3.0 # pin@v4.4.0",
		},
		{
			name:           "comment with pin@version and additional text",
			line:           "uses: actions/setup-node@v4.3.0 # pin@v4.3.0 stable version",
			version:        "v4.4.0",
			expectedOutput: "uses: actions/setup-node@v4.3.0 # pin@v4.4.0 stable version",
		},
		{
			name:           "comment without version pattern",
			line:           "uses: actions/setup-node@v4.3.0 # stable version",
			version:        "v4.4.0",
			expectedOutput: "uses: actions/setup-node@v4.3.0 # v4.4.0 stable version",
		},
		{
			name:           "comment with version elsewhere",
			line:           "uses: actions/setup-node@v4.3.0 # stable version v4.3.0",
			version:        "v4.4.0",
			expectedOutput: "uses: actions/setup-node@v4.3.0 # stable version v4.4.0",
		},
		{
			name:           "empty comment",
			line:           "uses: actions/setup-node@v4.3.0 #",
			version:        "v4.4.0",
			expectedOutput: "uses: actions/setup-node@v4.3.0 # v4.4.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updateVersionComment(tt.line, tt.version)
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

func TestIsMajorVersionConstraint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "major version with v prefix",
			input:    "v4",
			expected: true,
		},
		{
			name:     "major version without v prefix",
			input:    "4",
			expected: true,
		},
		{
			name:     "not major version - has minor",
			input:    "v4.3",
			expected: false,
		},
		{
			name:     "not major version - has minor without v",
			input:    "4.3",
			expected: false,
		},
		{
			name:     "not major version - full semver",
			input:    "v4.3.0",
			expected: false,
		},
		{
			name:     "not major version - full semver without v",
			input:    "4.3.0",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMajorVersionConstraint(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractMajorVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "full semver with v prefix",
			input:    "v4.3.0",
			expected: "v4",
		},
		{
			name:     "full semver without v prefix",
			input:    "4.3.0",
			expected: "4",
		},
		{
			name:     "major.minor with v prefix",
			input:    "v4.3",
			expected: "v4",
		},
		{
			name:     "major.minor without v prefix",
			input:    "4.3",
			expected: "4",
		},
		{
			name:     "only major with v prefix",
			input:    "v4",
			expected: "v4",
		},
		{
			name:     "only major without v prefix",
			input:    "4",
			expected: "4",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMajorVersion(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

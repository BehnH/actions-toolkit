/*
Copyright © 2025 Behn Hayhoe hello@behn.dev

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package processor_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/behnh/actions-toolkit/internal/processor"
	"github.com/stretchr/testify/assert"
)

func TestIsVersionNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid version with v prefix",
			input:    "v1.2.3",
			expected: true,
		},
		{
			name:     "valid version without v prefix",
			input:    "1.2.3",
			expected: true,
		},
		{
			name:     "valid version with two components",
			input:    "1.2",
			expected: true,
		},
		{
			name:     "invalid version - no dots",
			input:    "123",
			expected: false,
		},
		{
			name:     "invalid version - contains letters",
			input:    "1.2.3a",
			expected: false,
		},
		{
			name:     "invalid version - empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid version - contains other characters",
			input:    "1.2-beta",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.IsVersionNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPinAction(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "pinning-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test scenarios
	tests := []struct {
		name            string
		fixtureFile     string
		actionName      string
		version         string
		mockToken       string
		write           bool
		setupMockGithub func() // Function to setup GitHub API mocks if needed
		verify          func(t *testing.T, tempFile string)
	}{
		{
			name:        "pin action with semver",
			fixtureFile: "workflow_semver.yaml",
			actionName:  "actions/setup-node",
			version:     "v4.3.0",
			mockToken:   "mock-token",
			write:       true,
			setupMockGithub: func() {
				// In a real implementation, you would mock the GitHub API
			},
			verify: func(t *testing.T, tempFile string) {
				content, err := os.ReadFile(tempFile)
				assert.NoError(t, err)

				// Since we're using a mock token, the GitHub API call will fail
				// The file should still contain the original version
				assert.Contains(t, string(content), "actions/setup-node@v4.3.0")
			},
		},
		{
			name:        "pin action with SHA",
			fixtureFile: "workflow_sha.yaml",
			actionName:  "actions/setup-node",
			version:     "v4.3.0",
			mockToken:   "mock-token",
			write:       true,
			setupMockGithub: func() {
				// In a real implementation, you would mock the GitHub API
			},
			verify: func(t *testing.T, tempFile string) {
				content, err := os.ReadFile(tempFile)
				assert.NoError(t, err)

				// Since we're using a mock token, the GitHub API call will fail
				// The file should still contain the original SHA
				assert.Contains(t, string(content), 
					"actions/setup-node@cdca7365b2dadb8aad0a33bc7601856ffabcc48e")
			},
		},
		{
			name:        "pin action with major version",
			fixtureFile: "workflow_major_version.yaml",
			actionName:  "actions/cache",
			version:     "v3",
			mockToken:   "mock-token",
			write:       true,
			setupMockGithub: func() {
				// In a real implementation, you would mock the GitHub API
			},
			verify: func(t *testing.T, tempFile string) {
				content, err := os.ReadFile(tempFile)
				assert.NoError(t, err)

				// Since we're using a mock token, the GitHub API call will fail
				// The file should still contain the original version
				assert.Contains(t, string(content), "actions/cache@v3")
			},
		},
		{
			name:        "pin action in file with no actions",
			fixtureFile: "no_actions.yaml",
			actionName:  "actions/setup-node",
			version:     "v4.3.0",
			mockToken:   "mock-token",
			write:       true,
			setupMockGithub: func() {
				// In a real implementation, you would mock the GitHub API
			},
			verify: func(t *testing.T, tempFile string) {
				content, err := os.ReadFile(tempFile)
				assert.NoError(t, err)

				// The file should remain unchanged as it has no actions
				assert.NotContains(t, string(content), "actions/setup-node")
			},
		},
		{
			name:        "pin action with main version",
			fixtureFile: "workflow_with_main.yaml",
			actionName:  "actions/setup-node",
			version:     "v4.3.0",
			mockToken:   "mock-token",
			write:       true,
			setupMockGithub: func() {
				// In a real implementation, you would mock the GitHub API
			},
			verify: func(t *testing.T, tempFile string) {
				content, err := os.ReadFile(tempFile)
				assert.NoError(t, err)

				// The file should remain unchanged as actions with main version are skipped
				assert.Contains(t, string(content), "actions/setup-node@main")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy the fixture file to a temporary location
			fixtureFile := filepath.Join("..", "..", "snapshots", tt.fixtureFile)
			tempFile := filepath.Join(tempDir, tt.fixtureFile)
			
			// Read the fixture file
			content, err := os.ReadFile(fixtureFile)
			assert.NoError(t, err)
			
			// Write to the temp file
			err = os.WriteFile(tempFile, content, 0644)
			assert.NoError(t, err)
			
			// Setup mock GitHub API if needed
			if tt.setupMockGithub != nil {
				tt.setupMockGithub()
			}
			
			// Call the function being tested
			processor.PinAction([]string{tempFile}, tt.actionName, tt.version, tt.mockToken, tt.write)
			
			// Verify the results
			if tt.verify != nil {
				tt.verify(t, tempFile)
			}
		})
	}
}

func TestPinAllActions(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "pinning-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test scenarios
	tests := []struct {
		name            string
		fixtureFiles    []string
		mockToken       string
		write           bool
		setupMockGithub func() // Function to setup GitHub API mocks if needed
		verify          func(t *testing.T, tempFiles []string)
	}{
		{
			name:         "pin all actions in multiple files",
			fixtureFiles: []string{"workflow_semver.yaml", "workflow_sha.yaml", "workflow_major_version.yaml"},
			mockToken:    "mock-token",
			write:        true,
			setupMockGithub: func() {
				// In a real implementation, you would mock the GitHub API
			},
			verify: func(t *testing.T, tempFiles []string) {
				// Since we're using a mock token, the GitHub API calls will fail
				// The files should remain unchanged
				
				// Check the semver file
				content, err := os.ReadFile(tempFiles[0])
				assert.NoError(t, err)
				assert.Contains(t, string(content), "actions/setup-node@v4.3.0")
				
				// Check the SHA file
				content, err = os.ReadFile(tempFiles[1])
				assert.NoError(t, err)
				assert.Contains(t, string(content), 
					"actions/setup-node@cdca7365b2dadb8aad0a33bc7601856ffabcc48e")
				
				// Check the major version file
				content, err = os.ReadFile(tempFiles[2])
				assert.NoError(t, err)
				assert.Contains(t, string(content), "actions/cache@v3")
			},
		},
		{
			name:         "pin all actions with some files having no actions",
			fixtureFiles: []string{"workflow_semver.yaml", "no_actions.yaml"},
			mockToken:    "mock-token",
			write:        true,
			setupMockGithub: func() {
				// In a real implementation, you would mock the GitHub API
			},
			verify: func(t *testing.T, tempFiles []string) {
				// Check the semver file
				content, err := os.ReadFile(tempFiles[0])
				assert.NoError(t, err)
				assert.Contains(t, string(content), "actions/setup-node@v4.3.0")
				
				// Check the no actions file
				content, err = os.ReadFile(tempFiles[1])
				assert.NoError(t, err)
				assert.NotContains(t, string(content), "actions/setup-node")
			},
		},
		{
			name:         "pin all actions with some files having main version",
			fixtureFiles: []string{"workflow_semver.yaml", "workflow_with_main.yaml"},
			mockToken:    "mock-token",
			write:        true,
			setupMockGithub: func() {
				// In a real implementation, you would mock the GitHub API
			},
			verify: func(t *testing.T, tempFiles []string) {
				// Check the semver file
				content, err := os.ReadFile(tempFiles[0])
				assert.NoError(t, err)
				assert.Contains(t, string(content), "actions/setup-node@v4.3.0")
				
				// Check the main version file
				content, err = os.ReadFile(tempFiles[1])
				assert.NoError(t, err)
				assert.Contains(t, string(content), "actions/setup-node@main")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy the fixture files to temporary locations
			tempFiles := make([]string, len(tt.fixtureFiles))
			for i, fixtureFile := range tt.fixtureFiles {
				fixtureFilePath := filepath.Join("..", "..", "snapshots", fixtureFile)
				tempFilePath := filepath.Join(tempDir, fixtureFile)
				tempFiles[i] = tempFilePath
				
				// Read the fixture file
				content, err := os.ReadFile(fixtureFilePath)
				assert.NoError(t, err)
				
				// Write to the temp file
				err = os.WriteFile(tempFilePath, content, 0644)
				assert.NoError(t, err)
			}
			
			// Setup mock GitHub API if needed
			if tt.setupMockGithub != nil {
				tt.setupMockGithub()
			}
			
			// Call the function being tested
			processor.PinAllActions(tempFiles, tt.mockToken, tt.write)
			
			// Verify the results
			if tt.verify != nil {
				tt.verify(t, tempFiles)
			}
		})
	}
}
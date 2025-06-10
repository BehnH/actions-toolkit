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

package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetYAMLFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "yaml-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := map[string]bool{
		"config.yml":      true,  // Should be found
		"data.yaml":       true,  // Should be found
		"script.sh":       false, // Should not be found
		"readme.md":       false, // Should not be found
		"nested/test.yml": true,  // Should be found (in subdirectory)
	}

	// Create the files
	for filename, _ := range testFiles {
		fullPath := filepath.Join(tempDir, filename)
		// Create directory if needed
		dir := filepath.Dir(fullPath)
		if dir != tempDir {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create directory %s: %v", dir, err)
			}
		}
		// Create the file
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", fullPath, err)
		}
	}

	// Test the function
	yamlFiles, err := GetYAMLFiles(tempDir)
	if err != nil {
		t.Fatalf("GetYAMLFiles returned an error: %v", err)
	}

	// Check if all expected YAML files were found
	foundFiles := make(map[string]bool)
	for _, file := range yamlFiles {
		relPath, err := filepath.Rel(tempDir, file)
		if err != nil {
			t.Fatalf("Failed to get relative path: %v", err)
		}
		foundFiles[relPath] = true
	}

	// Verify results
	for filename, shouldBeFound := range testFiles {
		_, found := foundFiles[filename]
		if shouldBeFound && !found {
			t.Errorf("Expected to find %s, but it was not found", filename)
		} else if !shouldBeFound && found {
			t.Errorf("Did not expect to find %s, but it was found", filename)
		}
	}

	// Test with non-existent directory
	_, err = GetYAMLFiles(filepath.Join(tempDir, "nonexistent"))
	if err == nil {
		t.Error("Expected error for non-existent directory, but got nil")
	}

	// Test with a file path instead of a directory
	filePath := filepath.Join(tempDir, "config.yml")
	_, err = GetYAMLFiles(filePath)
	if err == nil {
		t.Error("Expected error when passing a file path instead of a directory, but got nil")
	}
}

func TestReadFile(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test-file-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write test content to the file
	testContent := []byte("This is a test file content")
	if _, err := tempFile.Write(testContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test reading the file
	content, err := ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("ReadFile returned an error: %v", err)
	}

	// Verify the content
	if string(content) != string(testContent) {
		t.Errorf("Expected content %q, got %q", string(testContent), string(content))
	}

	// Test with non-existent file
	_, err = ReadFile(tempFile.Name() + ".nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent file, but got nil")
	}
}

func TestParseYAMLForUses(t *testing.T) {
	// Test case 1: YAML with jobs.*.steps.* path
	yamlWithJobs := []byte(`
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-node@v1
        with:
          node-version: '14'
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v2
`)

	// Test case 2: YAML with runs.steps.* path
	yamlWithRuns := []byte(`
runs:
  using: composite
  steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-java@v1
      with:
        java-version: '11'
`)

	// Test case 3: YAML with both paths
	yamlWithBoth := []byte(`
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
runs:
  using: composite
  steps:
    - uses: actions/setup-go@v3
`)

	// Test case 4: Invalid YAML
	invalidYAML := []byte(`
this is not valid yaml
  - foo: bar
    baz
`)

	tests := []struct {
		name     string
		yaml     []byte
		expected []string
		wantErr  bool
	}{
		{
			name: "YAML with jobs.*.steps.* path",
			yaml: yamlWithJobs,
			expected: []string{
				"actions/checkout@v2",
				"actions/setup-node@v1",
				"actions/checkout@v3",
				"actions/setup-python@v2",
			},
			wantErr: false,
		},
		{
			name: "YAML with runs.steps.* path",
			yaml: yamlWithRuns,
			expected: []string{
				"actions/checkout@v2",
				"actions/setup-java@v1",
			},
			wantErr: false,
		},
		{
			name: "YAML with both paths",
			yaml: yamlWithBoth,
			expected: []string{
				"actions/checkout@v2",
				"actions/setup-go@v3",
			},
			wantErr: false,
		},
		{
			name:     "Invalid YAML",
			yaml:     invalidYAML,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseYAMLForUses(tt.yaml)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseYAMLForUses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Check if all expected values are present
				if len(got) != len(tt.expected) {
					t.Errorf("ParseYAMLForUses() got %d values, want %d", len(got), len(tt.expected))
					return
				}

				// Create a map for easier comparison
				expectedMap := make(map[string]bool)
				for _, v := range tt.expected {
					expectedMap[v] = true
				}

				for _, v := range got {
					if !expectedMap[v] {
						t.Errorf("ParseYAMLForUses() got unexpected value %q", v)
					}
				}
			}
		})
	}
}

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
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	yamlv3 "gopkg.in/yaml.v3"
)

// GetYAMLFiles returns a slice of paths to all YAML files (with .yml or .yaml extensions)
// in the specified directory.
func GetYAMLFiles(dirPath string) ([]string, error) {
	var yamlFiles []string

	// Check if the directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrNotExist
	}

	slog.Debug("Getting YAML files from directory", "directory", dirPath)

	// Walk through the directory
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}
		// Check if the file has a YAML extension
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yml" || ext == ".yaml" {
			yamlFiles = append(yamlFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	slog.Debug("Found YAML files", "files", yamlFiles)

	return yamlFiles, nil
}

func GetAllFiles(dirPath string) ([]string, error) {
	var files []string

	// Check if the directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrNotExist
	}

	// Walk through the directory
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// ReadFile reads the content of a file and returns it as a byte slice.
func ReadFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// ParseYAMLForUses parses a YAML file and extracts all 'uses' values from steps
// at paths matching 'jobs.*.steps.*' or 'runs.steps.*'.
func ParseYAMLForUses(content []byte) ([]string, error) {
	var data map[string]interface{}
	if err := yamlv3.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	var usesValues []string

	// Check for jobs.*.steps.* path
	if jobs, ok := data["jobs"].(map[string]interface{}); ok {
		for _, job := range jobs {
			if jobMap, ok := job.(map[string]interface{}); ok {
				if steps, ok := jobMap["steps"].([]interface{}); ok {
					for _, step := range steps {
						if stepMap, ok := step.(map[string]interface{}); ok {
							if uses, ok := stepMap["uses"].(string); ok {
								usesValues = append(usesValues, uses)
							}
						}
					}
				}
			}
		}
	}

	// Check for runs.steps.* path
	if runs, ok := data["runs"].(map[string]interface{}); ok {
		if steps, ok := runs["steps"].([]interface{}); ok {
			for _, step := range steps {
				if stepMap, ok := step.(map[string]interface{}); ok {
					if uses, ok := stepMap["uses"].(string); ok {
						usesValues = append(usesValues, uses)
					}
				}
			}
		}
	}

	return usesValues, nil
}

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

package processor

import (
	"fmt"
	"github.com/behnh/actions-toolkit/internal/file"
	"github.com/behnh/actions-toolkit/internal/github"
	"log/slog"
	"os"
	"strings"
)

func FindActionsInFile(filePath string) []string {
	// Read the file
	content, err := file.ReadFile(filePath)
	if err != nil {
		slog.Error("Failed to read file", "file", filePath, "error", err)
		return nil
	}

	// Parse the file for 'uses' values
	usesValues, err := file.ParseYAMLForUses(content)
	if err != nil {
		slog.Error("Failed to parse file", "file", filePath, "error", err)
		return nil
	}

	var actions []string

	// Find the specified action in the uses values
	for _, uses := range usesValues {
		parts := strings.Split(uses, "@")
		slog.Debug(fmt.Sprintf("Found uses value: %s", uses))
		if len(parts) != 2 {
			continue
		}

		// If the action version is not "main", add it to the list
		if parts[1] != "main" {
			actions = append(actions, strings.TrimSpace(parts[0]))
		}
	}

	return actions
}

func FindActionsInFiles(files []string) []string {
	// Use a map to store unique actions
	actionMap := make(map[string]bool)

	for _, f := range files {
		for _, action := range FindActionsInFile(f) {
			actionMap[action] = true
		}
	}

	// Convert map keys back to slice
	actions := make([]string, 0, len(actionMap))
	for action := range actionMap {
		actions = append(actions, action)
	}

	return actions
}

func UpdateAction(filePath, actionName, token string, write bool) {
	// Read the file
	content, err := file.ReadFile(filePath)
	if err != nil {
		slog.Error("Failed to read file", "file", filePath, "error", err)
		return
	}

	contentModified := false
	contentStr := string(content)

	// Parse the file for 'uses' values
	usesValues, err := file.ParseYAMLForUses(content)
	if err != nil {
		slog.Error("Failed to parse file", "file", filePath, "error", err)
		return
	}

	// Find the specified action in the uses values
	for _, uses := range usesValues {
		// Check if this uses value matches the action we're looking for
		if strings.HasPrefix(uses, actionName+"@") {
			parts := strings.Split(uses, "@")
			slog.Debug(fmt.Sprintf("Found uses value: %s", uses))
			if len(parts) != 2 {
				continue
			}

			currentVersion := parts[1]
			slog.Info("Found action", "action", actionName, "version", currentVersion, "file", filePath)

			// Skip if version is "main"
			if currentVersion == "main" {
				slog.Info("Skipping action with 'main' version", "action", actionName, "file", filePath)
				continue
			}

			// Get the latest release with SHA
			latestRelease, latestSHA, err := github.GetLatestReleaseWithSHA(token, actionName, currentVersion)
			if err != nil {
				slog.Error("Failed to get latest release", "action", actionName, "error", err)
				continue
			}

			if latestRelease == "" {
				slog.Info("No release found for action", "action", actionName)
				continue
			}

			// Compare versions
			if currentVersion != latestRelease || (len(currentVersion) == 40 && isHexString(currentVersion) && latestSHA != "" && currentVersion != latestSHA) {
				slog.Info("Update available",
					"action", actionName,
					"current", currentVersion,
					"latest", latestRelease,
					"latestSHA", latestSHA,
					"file", filePath)

				// Update the file if write is true
				if write {
					var newContent string

					// Check if the current version is an SHA (40 hex characters)
					isSHA := len(currentVersion) == 40 && isHexString(currentVersion)

					if isSHA {
						// Look for the line with this SHA in the file
						lines := strings.Split(contentStr, "\n")
						newLines := make([]string, len(lines))

						for i, line := range lines {
							if strings.Contains(line, actionName+"@"+currentVersion) {
								// Replace the SHA with the latest SHA
								updatedAction := fmt.Sprintf("%s@%s", actionName, latestSHA)
								updatedLine := strings.Replace(line, actionName+"@"+currentVersion, updatedAction, 1)

								slog.Debug("Updated line", "originalLine", line, "updatedLine", updatedLine)

								// Update the comment with the new version using the shared function
								newLines[i] = updateVersionComment(updatedLine, latestRelease)
								slog.Debug("Comment updated", "before", updatedLine, "after", newLines[i])
							} else {
								newLines[i] = line
							}
						}

						newContent = strings.Join(newLines, "\n")
						contentStr = newContent
						contentModified = true
					} else {
						// Check if the current version is a major version constraint
						if isMajorVersionConstraint(currentVersion) {
							// Extract the major version from the latest release
							latestMajorVersion := extractMajorVersion(latestRelease)
							// Replace the version in the file content, preserving the major version constraint
							newContent = strings.ReplaceAll(contentStr,
								actionName+"@"+currentVersion,
								actionName+"@"+latestMajorVersion)

							slog.Debug("Updating major version constraint",
								"action", actionName,
								"from", currentVersion,
								"to", latestMajorVersion)
						} else {
							// Replace the version in the file content with the full version
							newContent = strings.ReplaceAll(contentStr,
								actionName+"@"+currentVersion,
								actionName+"@"+latestRelease)

							slog.Debug("Updating full version",
								"action", actionName,
								"from", currentVersion,
								"to", latestRelease)
						}
						contentStr = newContent
						contentModified = true
					}

					slog.Info("Updated action in memory",
						"action", actionName,
						"from", currentVersion,
						"to", latestRelease,
						"file", filePath)
				} else {
					slog.Info("Dry run - not updating file",
						"action", actionName,
						"file", filePath,
						"hint", "Use --write/-w to update files")
				}
			} else {
				slog.Info("Action is already up to date",
					"action", actionName,
					"version", currentVersion,
					"file", filePath)
			}
		}
	}

	// Write the updated content back to the file if it was modified and write mode is enabled
	if contentModified && write {
		err = os.WriteFile(filePath, []byte(contentStr), 0644)
		if err != nil {
			slog.Error("Failed to write file", "file", filePath, "error", err)
			return
		}
		slog.Info("Successfully wrote file with all updates", "file", filePath)
	}
}

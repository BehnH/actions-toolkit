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
	"github.com/aymanbagabas/go-udiff"
	"os"
	"strings"

	"github.com/behnh/actions-toolkit/internal/file"
	"github.com/behnh/actions-toolkit/internal/github"
	"log/slog"
)

func PinAllActions(files []string, token string, write bool) {
	// Get all unique actions from the files
	actions := FindActionsInFiles(files)
	slog.Info("Found actions to pin", "count", len(actions), "actions", actions)

	// Process each file for each action
	for _, f := range files {
		// Read the file
		content, err := file.ReadFile(f)
		if err != nil {
			slog.Error("Failed to read file", "file", f, "error", err)
			continue
		}

		contentModified := false
		contentStr := string(content)

		// Parse the file for 'uses' values
		usesValues, err := file.ParseYAMLForUses(content)
		if err != nil {
			slog.Error("Failed to parse file", "file", f, "error", err)
			continue
		}

		// Process each uses value in the file
		for _, uses := range usesValues {
			parts := strings.Split(uses, "@")
			if len(parts) != 2 {
				continue
			}

			actionName := strings.TrimSpace(parts[0])
			currentVersion := parts[1]

			// Skip if version is "main"
			if currentVersion == "main" {
				slog.Debug("Skipping action with 'main' version", "action", actionName, "file", f)
				continue
			}

			slog.Debug("Processing action", "action", actionName, "version", currentVersion, "file", f)

			// Get the latest release with SHA
			latestRelease, latestSHA, err := github.GetLatestReleaseWithSHA(token, actionName, currentVersion)
			if err != nil {
				slog.Error("Failed to get latest release", "action", actionName, "error", err)
				continue
			}

			if latestRelease == "" || latestSHA == "" {
				slog.Info("No release or SHA found for action", "action", actionName)
				continue
			}

			// Check if current version is already a SHA
			isSHA := len(currentVersion) == 40 && isHexString(currentVersion)

			// Check if update is needed
			if isSHA && currentVersion == latestSHA {
				slog.Info("Action is already pinned to latest SHA", "action", actionName, "file", f)
				continue
			}

			// Update the action to use the SHA
			lines := strings.Split(contentStr, "\n")
			newLines := make([]string, len(lines))

			for i, line := range lines {
				if strings.Contains(line, actionName+"@"+currentVersion) {
					// Replace with SHA
					updatedAction := fmt.Sprintf("%s@%s", actionName, latestSHA)
					updatedLine := strings.Replace(line, actionName+"@"+currentVersion, updatedAction, 1)
					newLines[i] = updateVersionComment(updatedLine, latestRelease)

					slog.Debug("Updated line", "originalLine", line, "updatedLine", newLines[i])
				} else {
					newLines[i] = line
				}
			}

			contentStr = strings.Join(newLines, "\n")
			contentModified = true

			slog.Debug("Updated action in memory",
				"action", actionName,
				"from", currentVersion,
				"to", latestSHA,
				"version", latestRelease,
				"file", f)
		}

		// Write changes to file if needed
		if contentModified && write {
			err = os.WriteFile(f, []byte(contentStr), 0644)
			if err != nil {
				slog.Error("Failed to write file", "file", f, "error", err)
				continue
			}
			slog.Info("Successfully updated file with pinned actions", "file", f)
		} else if contentModified {
			slog.Info(fmt.Sprintf("Dry run - not updating file. Would have applied:\n%s\n", udiff.Unified(f, f, string(content), contentStr)))
		} else {
			slog.Info("No changes to file", "file", f)
		}
	}
}

// IsVersionNumber checks if a string looks like a version number (e.g., 1.2.3 or 1.2)
func IsVersionNumber(s string) bool {
	// Remove 'v' prefix if it exists to standardize the check
	if strings.HasPrefix(s, "v") {
		s = s[1:]
	}

	// Check if the string contains digits and dots only
	for _, c := range s {
		if (c < '0' || c > '9') && c != '.' {
			return false
		}
	}

	// Must contain at least one dot for a version number
	return strings.Contains(s, ".")
}

func PinAction(files []string, actionName string, version, token string, write bool) {
	for _, f := range files {
		// Read the file
		content, err := file.ReadFile(f)
		if err != nil {
			slog.Error("Failed to read file", "file", f, "error", err)
			return
		}

		contentModified := false
		contentStr := string(content)

		// Parse the file for 'uses' values
		usesValues, err := file.ParseYAMLForUses(content)
		if err != nil {
			slog.Error("Failed to parse file", "file", f, "error", err)
			return
		}

		// Get the SHA for the specific version
		_, latestSHA, err := github.GetLatestReleaseWithSHA(token, actionName, version)
		if err != nil {
			slog.Error("Failed to get SHA for version", "action", actionName, "version", version, "error", err)
			return
		}

		if latestSHA == "" {
			slog.Error("No SHA found for version", "action", actionName, "version", version)
			return
		}

		// Find the specified action in the uses values
		for _, uses := range usesValues {
			// Check if this uses value matches the action we're looking for
			if strings.HasPrefix(uses, actionName+"@") {
				parts := strings.Split(uses, "@")
				if len(parts) != 2 {
					continue
				}

				currentVersion := parts[1]
				slog.Info("Found action", "action", actionName, "version", currentVersion, "file", f)

				// Skip if version is "main"
				if currentVersion == "main" {
					slog.Info("Skipping action with 'main' version", "action", actionName, "file", f)
					continue
				}

				// Update to the specified SHA
				lines := strings.Split(contentStr, "\n")
				newLines := make([]string, len(lines))

				for i, line := range lines {
					if strings.Contains(line, actionName+"@"+currentVersion) {
						// Replace with SHA
						updatedAction := fmt.Sprintf("%s@%s", actionName, latestSHA)
						updatedLine := strings.Replace(line, actionName+"@"+currentVersion, updatedAction, 1)
						newLines[i] = updateVersionComment(updatedLine, version)

						slog.Debug("Updated line", "originalLine", line, "updatedLine", newLines[i])
					} else {
						newLines[i] = line
					}
				}

				contentStr = strings.Join(newLines, "\n")
				contentModified = true

				if write {
					slog.Info("Updated action in memory",
						"action", actionName,
						"from", currentVersion,
						"to", latestSHA,
						"version", version,
						"file", f)
				} else {
					slog.Info("Dry run - not updating file",
						"action", actionName,
						"file", f,
						"hint", "Use --write/-w to update files")
				}
			}
		}

		// Write the updated content back to the file if it was modified and write is enabled
		if contentModified && write {
			err = os.WriteFile(f, []byte(contentStr), 0644)
			if err != nil {
				slog.Error("Failed to write file", "file", f, "error", err)
				return
			}
			slog.Info("Successfully wrote file with pinned action", "file", f)
		}
	}
}

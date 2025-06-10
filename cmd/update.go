/*
Copyright © 2025 Behn Hayhoe hello@behn.dev
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/behnh/actions-toolkit/internal/file"
	"github.com/behnh/actions-toolkit/internal/github"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update GitHub Actions to their latest versions",
	Long: `Update GitHub Actions to their latest versions in workflow files.
You can specify a specific action to update, or update all actions in a file or directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		actionName, _ := cmd.Flags().GetString("action")
		dirPath, _ := cmd.Flags().GetString("dir")
		filePath, _ := cmd.Flags().GetString("file")
		write, _ := cmd.Flags().GetBool("write")
		token, _ := cmd.Flags().GetString("token")

		if actionName == "" {
			slog.Error("Action name is required")
			return
		}

		var filesToProcess []string
		var err error

		if filePath != "" {
			filesToProcess = []string{filePath}
		} else if dirPath != "" {
			filesToProcess, err = file.GetYAMLFiles(dirPath)
			if err != nil {
				slog.Error("Failed to get YAML files", "error", err)
				return
			}
		} else {
			slog.Error("Either --dir or --file must be specified")
			return
		}

		for _, f := range filesToProcess {
			processFile(f, actionName, token, write)
		}
	},
}

func isHexString(s string) bool {
	match, _ := regexp.MatchString("^[0-9a-fA-F]+$", s)
	return match
}

// isMajorVersionConstraint checks if a version string is a major version constraint,
// For example, "v4" or "4" are major version constraints, but "v4.3.0" is not
func isMajorVersionConstraint(version string) bool {
	if version == "" {
		return false
	}

	// Remove the 'v' prefix if it exists
	versionStr := version
	if strings.HasPrefix(versionStr, "v") {
		versionStr = versionStr[1:]
	}

	// Check if the version contains any dots
	return !strings.Contains(versionStr, ".")
}

// extractMajorVersion extracts the major version from a full version string
// For example, "v3.5.0" -> "v3", "v1.2.3" -> "v1"
func extractMajorVersion(version string) string {
	if version == "" {
		return ""
	}

	// Remove the 'v' prefix if it exists
	versionStr := version
	if strings.HasPrefix(versionStr, "v") {
		versionStr = versionStr[1:]
	}

	// Split by dots and take the first part
	parts := strings.Split(versionStr, ".")
	if len(parts) == 0 {
		return version // Return original if no parts
	}

	// Add 'v' prefix back if it was there
	if strings.HasPrefix(version, "v") {
		return "v" + parts[0]
	}

	return parts[0]
}

// processFile reads a file, finds actions, and updates them if needed
func processFile(filePath, actionName, token string, write bool) {
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

								// Check if there's a comment after the SHA
								parts := strings.SplitN(updatedLine, "#", 2)
								if len(parts) > 1 {
									// Update the comment with the new version
									newLines[i] = strings.TrimRight(parts[0], " ") + " # " + latestRelease
								} else {
									// Add a comment with the new version
									newLines[i] = updatedLine + " # " + latestRelease
								}
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

func init() {
	rootCmd.AddCommand(updateCmd)

	// Add flags specific to the update command
	updateCmd.Flags().String("action", "", "Action name to update (required)")
	updateCmd.Flags().String("dir", "", "Directory containing workflow files")
	updateCmd.Flags().String("file", "", "Specific workflow file to update")
	updateCmd.Flags().BoolP("write", "w", false, "Write changes to files (default is dry run)")

	// Mark action as required
	updateCmd.MarkFlagRequired("action")
}

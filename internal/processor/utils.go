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
	"regexp"
	"strings"
)

func isHexString(s string) bool {
	match, _ := regexp.MatchString("^[0-9a-fA-F]+$", s)
	return match
}

func updateVersionComment(line string, version string) string {
	parts := strings.SplitN(line, "#", 2)
	baseContent := strings.TrimRight(parts[0], " ")

	// No existing comment
	if len(parts) <= 1 {
		return baseContent + " # " + version
	}

	// We have an existing comment to update
	commentText := strings.TrimSpace(parts[1])
	commentParts := strings.Fields(commentText)

	// Check for different comment patterns
	if len(commentParts) > 0 {
		// Case 1: Comment starts with a version number (v4.3.0 or 4.3.0)
  if strings.HasPrefix(commentParts[0], "v") || IsVersionNumber(commentParts[0]) {
			// Replace the version part, keep any additional text
			if len(commentParts) > 1 {
				newComment := version + " " + strings.Join(commentParts[1:], " ")
				return baseContent + " # " + newComment
			} else {
				// Just replace the version
				return baseContent + " # " + version
			}
		} else if strings.Contains(commentParts[0], "@") {
			// Case 2: Comment contains a pin@v4 pattern
			pinParts := strings.Split(commentParts[0], "@")
			if len(pinParts) == 2 {
				// Replace the version part but keep the prefix
				prefix := pinParts[0]
				if len(commentParts) > 1 {
					// There are additional words after pin@v4
					newComment := prefix + "@" + version + " " + strings.Join(commentParts[1:], " ")
					return baseContent + " # " + newComment
				} else {
					// Just the pin@version in the comment
					return baseContent + " # " + prefix + "@" + version
				}
			} else {
				// Malformed pin@version, just prepend the version
				return baseContent + " # " + version + " " + commentText
			}
		} else {
			// Case 3: Comment doesn't match any known version pattern
			// Check if there's already a version-looking string anywhere in the comment
			foundVersion := false
			for i, part := range commentParts {
    if (strings.HasPrefix(part, "v") && IsVersionNumber(part[1:])) || IsVersionNumber(part) {
					// Replace this part with the new version
					commentParts[i] = version
					foundVersion = true
					break
				}
			}

			if foundVersion {
				return baseContent + " # " + strings.Join(commentParts, " ")
			} else {
				// If no version found, prepend the new version
				return baseContent + " # " + version + " " + commentText
			}
		}
	} else {
		// Empty comment
		return baseContent + " # " + version
	}
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

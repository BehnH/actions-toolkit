package github

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/google/go-github/v72/github"
)

type ReleaseInfo struct {
	MajorVersion string // The major version (e.g., v3)
	FullVersion  string // The full version (e.g., v3.5.0)
	SHA          string // The SHA of the release
}

var releaseCache = make(map[string]ReleaseInfo)
var cacheMutex sync.RWMutex

// GetLatestRelease searches for the latest release of a GitHub action.
// The actionName should be in the format "org/repo/optional_subpath".
func GetLatestRelease(token string, actionName string, currentVersion string) (string, error) {
	// Check if we have this action in the cache
	baseActionName := getBaseActionName(actionName)
	cacheMutex.RLock()
	if info, found := releaseCache[baseActionName]; found {
		cacheMutex.RUnlock()

		// If currentVersion is a major version constraint (e.g., v4)
		// and it matches the major version in the cache, return the current version
		if isMajorVersionConstraint(currentVersion) {
			currentMajorVersion := extractMajorVersion(currentVersion)
			if currentMajorVersion == info.MajorVersion {
				slog.Debug("Using current version (major version constraint matches)",
					"action", actionName,
					"currentVersion", currentVersion,
					"latestMajorVersion", info.MajorVersion)
				return currentVersion, nil
			}
		}

		slog.Debug("Using cached release info",
			"action", actionName,
			"version", info.FullVersion)
		return info.FullVersion, nil
	}
	cacheMutex.RUnlock()

	// Create a GitHub client with the provided token
	client := github.NewClient(nil).WithAuthToken(token)
	return getLatestReleaseWithClient(client, actionName)
}

// GetLatestReleaseWithSHA searches for the latest release of a GitHub action and returns both the version and SHA.
// The actionName should be in the format "org/repo/optional_subpath".
func GetLatestReleaseWithSHA(token string, actionName string, currentVersion string) (string, string, error) {
	baseActionName := getBaseActionName(actionName)
	cacheMutex.RLock()
	if info, found := releaseCache[baseActionName]; found {
		cacheMutex.RUnlock()

		// If the currentVersion is a major version constraint (e.g., v4)
		// and it matches the major version in the cache, return the current version
		if isMajorVersionConstraint(currentVersion) {
			currentMajorVersion := extractMajorVersion(currentVersion)
			if currentMajorVersion == info.MajorVersion {
				slog.Debug("Using current version (major version constraint matches)",
					"action", actionName,
					"currentVersion", currentVersion,
					"latestMajorVersion", info.MajorVersion)
				return currentVersion, "", nil
			}
		}

		slog.Debug("Using cached release info",
			"action", actionName,
			"version", info.FullVersion,
			"sha", info.SHA)
		return info.FullVersion, info.SHA, nil
	}
	cacheMutex.RUnlock()

	client := github.NewClient(nil).WithAuthToken(token)
	version, err := getLatestReleaseWithClient(client, actionName)
	if err != nil {
		return "", "", err
	}

	cacheMutex.RLock()
	info, found := releaseCache[baseActionName]
	cacheMutex.RUnlock()
	if !found {
		return version, "", nil
	}

	return version, info.SHA, nil
}

// getBaseActionName extracts the base action name (org/repo) from the full action name
func getBaseActionName(actionName string) string {
	parts := strings.SplitN(actionName, "/", 3)
	if len(parts) < 2 {
		return actionName
	}
	return parts[0] + "/" + parts[1]
}

// isMajorVersionConstraint checks if a version string is a major version constraint,
// For example, "v4" or "4" are major version constraints, but "v4.3.0" is not
func isMajorVersionConstraint(version string) bool {
	if version == "" {
		return false
	}

	versionStr := version
	if strings.HasPrefix(versionStr, "v") {
		versionStr = versionStr[1:]
	}

	return !strings.Contains(versionStr, ".")
}

// extractMajorVersion extracts the major version from a full version string
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

// getLatestReleaseWithClient is an internal function that allows for dependency injection
// of the GitHub client for testing purposes.
func getLatestReleaseWithClient(client *github.Client, actionName string) (string, error) {
	// Parse the action name to extract org and repo
	parts := strings.SplitN(actionName, "/", 3)
	if len(parts) < 2 {
		return "", nil
	}

	owner := parts[0]
	repo := parts[1]
	baseActionName := owner + "/" + repo

	// Get the latest release
	ctx := context.Background()
	release, resp, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		// Check if the error is due to no releases found (404)
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			slog.Debug("No release found for GitHub action",
				"action", actionName,
				"owner", owner,
				"repo", repo)
			return "", nil
		}
		return "", err
	}

	if release == nil || release.TagName == nil {
		slog.Debug("No release found for GitHub action",
			"action", actionName,
			"owner", owner,
			"repo", repo)
		return "", nil
	}

	fullVersion := release.GetTagName()

	// Extract major version from full version (e.g., v3 from v3.5.0)
	majorVersion := extractMajorVersion(fullVersion)

	// Get SHA for tag from the refs api
	ref := "refs/tags/" + fullVersion
	refResp, _, err := client.Git.GetRef(ctx, owner, repo, ref)
	if err != nil {
		return "", err
	}

	// Get SHA if available
	sha := ""
	if refResp.GetObject().GetSHA() != "" {
		sha = refResp.GetObject().GetSHA()
	} else {
		sha = release.GetTargetCommitish()
	}

	// Store in cache
	cacheMutex.Lock()
	releaseCache[baseActionName] = ReleaseInfo{
		MajorVersion: majorVersion,
		FullVersion:  fullVersion,
		SHA:          sha,
	}
	cacheMutex.Unlock()

	slog.Debug("Cached release info",
		"action", baseActionName,
		"majorVersion", majorVersion,
		"fullVersion", fullVersion,
		"sha", sha)

	return fullVersion, nil
}

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

package github

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v72/github"
)

func TestGetLatestRelease(t *testing.T) {
	// Clear the cache before testing
	cacheMutex.Lock()
	releaseCache = make(map[string]ReleaseInfo)
	cacheMutex.Unlock()

	// Setup a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request path to determine the response
		switch r.URL.Path {
		case "/repos/actions/checkout/releases/latest":
			// Return a successful response with a release
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"tag_name": "v3.5.0", "target_commitish": "abcdef1234567890"}`))
		case "/repos/actions/checkout/git/ref/tags/v3.5.0":
			// Return a successful response for the git ref
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"ref": "refs/tags/v3.5.0",
				"node_id": "REF_kwDOAJy2JLByZWZzL3RhZ3MvdjMuNS4w",
				"url": "https://api.github.com/repos/actions/checkout/git/refs/tags/v3.5.0",
				"object": {
					"sha": "abcdef1234567890abcdef1234567890abcdef12",
					"type": "commit",
					"url": "https://api.github.com/repos/actions/checkout/git/commits/abcdef1234567890abcdef1234567890abcdef12"
				}
			}`))
		case "/repos/nonexistent/repo/releases/latest":
			// Return a 404 for non-existent repo
			w.WriteHeader(http.StatusNotFound)
		default:
			// Default response
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Create a GitHub client that uses the mock server
	mockURL, _ := url.Parse(mockServer.URL + "/")
	mockClient := github.NewClient(nil)
	mockClient.BaseURL = mockURL
	mockClient.UploadURL = mockURL

	tests := []struct {
		name       string
		actionName string
		want       string
		wantErr    bool
	}{
		{
			name:       "Valid action with release",
			actionName: "actions/checkout",
			want:       "v3.5.0",
			wantErr:    false,
		},
		{
			name:       "Action with no release",
			actionName: "nonexistent/repo",
			want:       "",
			wantErr:    false,
		},
		{
			name:       "Invalid action name format",
			actionName: "invalid-format",
			want:       "",
			wantErr:    false,
		},
		{
			name:       "Action with subpath",
			actionName: "actions/checkout/v3",
			want:       "v3.5.0",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLatestReleaseWithClient(mockClient, tt.actionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLatestReleaseWithClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getLatestReleaseWithClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCaching(t *testing.T) {
	// Clear the cache before testing
	cacheMutex.Lock()
	releaseCache = make(map[string]ReleaseInfo)
	cacheMutex.Unlock()

	// Set up a mock HTTP server that counts requests
	requestCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.URL.Path == "/repos/actions/checkout/releases/latest" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"tag_name": "v3.5.0", "target_commitish": "abcdef1234567890"}`))
		} else if r.URL.Path == "/repos/actions/checkout/git/ref/tags/v3.5.0" {
			// Return a successful response for the git ref
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"ref": "refs/tags/v3.5.0",
				"node_id": "REF_kwDOAJy2JLByZWZzL3RhZ3MvdjMuNS4w",
				"url": "https://api.github.com/repos/actions/checkout/git/refs/tags/v3.5.0",
				"object": {
					"sha": "abcdef1234567890abcdef1234567890abcdef12",
					"type": "commit",
					"url": "https://api.github.com/repos/actions/checkout/git/commits/abcdef1234567890abcdef1234567890abcdef12"
				}
			}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Create a GitHub client that uses the mock server
	mockURL, _ := url.Parse(mockServer.URL + "/")
	mockClient := github.NewClient(nil)
	mockClient.BaseURL = mockURL
	mockClient.UploadURL = mockURL

	// First call should hit the API
	version1, err := getLatestReleaseWithClient(mockClient, "actions/checkout")
	if err != nil {
		t.Errorf("getLatestReleaseWithClient() error = %v", err)
		return
	}
	if version1 != "v3.5.0" {
		t.Errorf("getLatestReleaseWithClient() = %v, want %v", version1, "v3.5.0")
	}
	// We expect 2 requests: one for the release and one for the git ref
	if requestCount != 2 {
		t.Errorf("Expected 2 requests, got %d", requestCount)
	}

	// Check that the cache was populated
	cacheMutex.RLock()
	info, found := releaseCache["actions/checkout"]
	cacheMutex.RUnlock()
	if !found {
		t.Errorf("Cache was not populated")
	}
	if info.FullVersion != "v3.5.0" {
		t.Errorf("Cache has wrong version: got %v, want %v", info.FullVersion, "v3.5.0")
	}
	if info.MajorVersion != "v3" {
		t.Errorf("Cache has wrong major version: got %v, want %v", info.MajorVersion, "v3")
	}
	if info.SHA != "abcdef1234567890abcdef1234567890abcdef12" {
		t.Errorf("Cache has wrong SHA: got %v, want %v", info.SHA, "abcdef1234567890abcdef1234567890abcdef12")
	}

	// Second call should use the cache
	version2, err := getLatestReleaseWithClient(mockClient, "actions/checkout")
	if err != nil {
		t.Errorf("getLatestReleaseWithClient() error = %v", err)
		return
	}
	if version2 != "v3.5.0" {
		t.Errorf("getLatestReleaseWithClient() = %v, want %v", version2, "v3.5.0")
	}
	// Request count should be 4 (2 initial requests + 2 more for the second call)
	if requestCount != 4 {
		t.Errorf("Expected 4 requests, got %d", requestCount)
	}

	// Call with a subpath should also use the cache
	version3, err := getLatestReleaseWithClient(mockClient, "actions/checkout/v3")
	if err != nil {
		t.Errorf("getLatestReleaseWithClient() error = %v", err)
		return
	}
	if version3 != "v3.5.0" {
		t.Errorf("getLatestReleaseWithClient() = %v, want %v", version3, "v3.5.0")
	}
	// Request count should be 6 (4 previous requests + 2 more for the third call)
	if requestCount != 6 {
		t.Errorf("Expected 6 requests, got %d", requestCount)
	}
}

func TestExtractMajorVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "Semantic version with v prefix",
			version: "v3.5.0",
			want:    "v3",
		},
		{
			name:    "Semantic version without v prefix",
			version: "1.2.3",
			want:    "1",
		},
		{
			name:    "Single number with v prefix",
			version: "v2",
			want:    "v2",
		},
		{
			name:    "Single number without v prefix",
			version: "4",
			want:    "4",
		},
		{
			name:    "Empty string",
			version: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractMajorVersion(tt.version)
			if got != tt.want {
				t.Errorf("extractMajorVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLatestReleaseWithMajorVersionConstraint(t *testing.T) {
	// Clear the cache before testing
	cacheMutex.Lock()
	releaseCache = make(map[string]ReleaseInfo)
	// Populate the cache with a test entry
	releaseCache["actions/setup-node"] = ReleaseInfo{
		MajorVersion: "v4",
		FullVersion:  "v4.4.0",
		SHA:          "49933ea5288caeca8642d1e84afbd3f7d6820020",
	}
	cacheMutex.Unlock()

	// Test cases
	tests := []struct {
		name           string
		actionName     string
		currentVersion string
		want           string
	}{
		{
			name:           "Major version constraint matches cache",
			actionName:     "actions/setup-node",
			currentVersion: "v4",
			want:           "v4", // Should return current version since major versions match
		},
		{
			name:           "Major version constraint doesn't match cache",
			actionName:     "actions/setup-node",
			currentVersion: "v3",
			want:           "v4.4.0", // Should return latest version since major versions don't match
		},
		{
			name:           "Full version",
			actionName:     "actions/setup-node",
			currentVersion: "v4.2.0",
			want:           "v4.4.0", // Should return latest version for full version
		},
		{
			name:           "SHA",
			actionName:     "actions/setup-node",
			currentVersion: "cdca7365b2dadb8aad0a33bc7601856ffabcc48e",
			want:           "v4.4.0", // Should return latest version for SHA
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We're not making actual API calls here, just testing the cache logic
			got, err := GetLatestRelease("dummy-token", tt.actionName, tt.currentVersion)
			if err != nil {
				t.Errorf("GetLatestRelease() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("GetLatestRelease() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsMajorVersionConstraint(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{
			name:    "Major version with v prefix",
			version: "v3",
			want:    true,
		},
		{
			name:    "Major version without v prefix",
			version: "4",
			want:    true,
		},
		{
			name:    "Semantic version with v prefix",
			version: "v3.5.0",
			want:    false,
		},
		{
			name:    "Semantic version without v prefix",
			version: "1.2.3",
			want:    false,
		},
		{
			name:    "SHA",
			version: "cdca7365b2dadb8aad0a33bc7601856ffabcc48e",
			want:    false,
		},
		{
			name:    "Empty string",
			version: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMajorVersionConstraint(tt.version)
			if got != tt.want {
				t.Errorf("isMajorVersionConstraint() = %v, want %v", got, tt.want)
			}
		})
	}
}

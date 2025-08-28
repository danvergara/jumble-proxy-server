package github

import (
	"testing"
)

func TestGithubClient_GetURLResourceType(t *testing.T) {
	client := &GithubClient{client: nil} // client can be nil since it's not used

	tests := []struct {
		name        string
		url         string
		expected    URLResourceInfo
		expectError bool
	}{
		{
			name: "user profile",
			url:  "https://github.com/jung-kurt",
			expected: URLResourceInfo{
				Type:   User,
				Owner:  "jung-kurt",
				Repo:   "",
				Number: 0,
			},
			expectError: false,
		},
		{
			name: "repository",
			url:  "https://github.com/jung-kurt/gofpdf",
			expected: URLResourceInfo{
				Type:   Repository,
				Owner:  "jung-kurt",
				Repo:   "gofpdf",
				Number: 0,
			},
			expectError: false,
		},
		{
			name: "issue",
			url:  "https://github.com/jung-kurt/gofpdf/issues/336",
			expected: URLResourceInfo{
				Type:   Issue,
				Owner:  "jung-kurt",
				Repo:   "gofpdf",
				Number: 336,
			},
			expectError: false,
		},
		{
			name: "pull request",
			url:  "https://github.com/golang/go/pull/12345",
			expected: URLResourceInfo{
				Type:   PullRequest,
				Owner:  "golang",
				Repo:   "go",
				Number: 12345,
			},
			expectError: false,
		},
		{
			name: "repository with trailing slash",
			url:  "https://github.com/owner/repo/",
			expected: URLResourceInfo{
				Type:   Repository,
				Owner:  "owner",
				Repo:   "repo",
				Number: 0,
			},
			expectError: false,
		},
		{
			name: "user with trailing slash",
			url:  "https://github.com/username/",
			expected: URLResourceInfo{
				Type:   User,
				Owner:  "username",
				Repo:   "",
				Number: 0,
			},
			expectError: false,
		},
		{
			name: "release tag",
			url:  "https://github.com/owner/repo/releases/tag/v1.0.0",
			expected: URLResourceInfo{
				Type:    Release,
				Owner:   "owner",
				Repo:    "repo",
				Number:  0,
				Version: "v1.0.0",
			},
			expectError: false,
		},
		{
			name: "commits",
			url:  "https://github.com/owner/repo/commit/abc123",
			expected: URLResourceInfo{
				Type:   Commit,
				Owner:  "owner",
				Repo:   "repo",
				Number: 0,
				SHA:    "fe114c64733d850007f181bb029d9cc2237efe0f",
			},
			expectError: false,
		},
		{
			name: "unknown resource type - wiki",
			url:  "https://github.com/owner/repo/wiki",
			expected: URLResourceInfo{
				Type:   Unknown,
				Owner:  "",
				Repo:   "",
				Number: 0,
			},
			expectError: false,
		},
		{
			name: "empty path",
			url:  "https://github.com/",
			expected: URLResourceInfo{
				Type:   Unknown,
				Owner:  "",
				Repo:   "",
				Number: 0,
			},
			expectError: false,
		},
		{
			name: "root github",
			url:  "https://github.com",
			expected: URLResourceInfo{
				Type:   Unknown,
				Owner:  "",
				Repo:   "",
				Number: 0,
			},
			expectError: false,
		},
		{
			name: "issue with invalid number",
			url:  "https://github.com/owner/repo/issues/abc",
			expected: URLResourceInfo{
				Type:   Unknown,
				Owner:  "owner",
				Repo:   "repo",
				Number: 0,
			},
			expectError: true,
		},
		{
			name: "pull request with invalid number",
			url:  "https://github.com/owner/repo/pull/xyz",
			expected: URLResourceInfo{
				Type:   Unknown,
				Owner:  "owner",
				Repo:   "repo",
				Number: 0,
			},
			expectError: true,
		},
		{
			name: "release with semantic version",
			url:  "https://github.com/golang/go/releases/tag/go1.21.0",
			expected: URLResourceInfo{
				Type:    Release,
				Owner:   "golang",
				Repo:    "go",
				Number:  0,
				Version: "go1.21.0",
			},
			expectError: false,
		},
		{
			name: "release without v prefix",
			url:  "https://github.com/owner/repo/releases/tag/1.0.0",
			expected: URLResourceInfo{
				Type:    Release,
				Owner:   "owner",
				Repo:    "repo",
				Number:  0,
				Version: "1.0.0",
			},
			expectError: false,
		},
		{
			name: "release with complex tag",
			url:  "https://github.com/owner/repo/releases/tag/release-2023.10.15",
			expected: URLResourceInfo{
				Type:    Release,
				Owner:   "owner",
				Repo:    "repo",
				Number:  0,
				Version: "release-2023.10.15",
			},
			expectError: false,
		},
		{
			name: "release with trailing slash",
			url:  "https://github.com/owner/repo/releases/tag/v2.0.0/",
			expected: URLResourceInfo{
				Type:    Release,
				Owner:   "owner",
				Repo:    "repo",
				Number:  0,
				Version: "v2.0.0",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.getResourceFromURL(tt.url)

			if tt.expectError && err == nil {
				t.Errorf("GetURLResourceType() expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("GetURLResourceType() unexpected error: %v", err)
			}

			if result.Type != tt.expected.Type {
				t.Errorf(
					"GetURLResourceType() Type = %v, expected %v",
					result.Type,
					tt.expected.Type,
				)
			}

			if result.Owner != tt.expected.Owner {
				t.Errorf(
					"GetURLResourceType() Owner = %v, expected %v",
					result.Owner,
					tt.expected.Owner,
				)
			}

			if result.Repo != tt.expected.Repo {
				t.Errorf(
					"GetURLResourceType() Repo = %v, expected %v",
					result.Repo,
					tt.expected.Repo,
				)
			}

			if result.Number != tt.expected.Number {
				t.Errorf(
					"GetURLResourceType() Number = %v, expected %v",
					result.Number,
					tt.expected.Number,
				)
			}

			if result.Version != tt.expected.Version {
				t.Errorf(
					"GetURLResourceType() Version = %v, expected %v",
					result.Version,
					tt.expected.Version,
				)
			}
		})
	}
}

func TestResourceType_String(t *testing.T) {
	tests := []struct {
		resourceType ResourceType
		expected     string
	}{
		{User, "user"},
		{Repository, "repository"},
		{Issue, "issue"},
		{PullRequest, "pull_request"},
		{Release, "release"},
		{Commit, "commit"},
		{Unknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.resourceType.String()
			if result != tt.expected {
				t.Errorf("ResourceType.String() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

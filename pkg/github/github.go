package github

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/google/go-github/v74/github"
)

type ResourceType int

const (
	Unknown ResourceType = iota
	User
	Repository
	Issue
	PullRequest
)

func (rt ResourceType) String() string {
	switch rt {
	case User:
		return "user"
	case Repository:
		return "repository"
	case Issue:
		return "issue"
	case PullRequest:
		return "pull_request"
	default:
		return "unknown"
	}
}

type URLResourceInfo struct {
	Type   ResourceType
	Owner  string
	Repo   string
	Number int // For issues and PRs
}

type GithubClient struct {
	client *github.Client
}

// IsGitHubURL method determines if a given URL belongs to GitHub.
func (gc *GithubClient) IsGitHubURL(url string) bool {
	return strings.Contains(strings.ToLower(url), "github.com") ||
		strings.Contains(strings.ToLower(url), "api.github.com") ||
		strings.Contains(strings.ToLower(url), "gist.github.com")
}

// GetURLResourceType determines the type of GitHub resource from a URL
func (gc *GithubClient) GetURLResourceType(rawURL string) (URLResourceInfo, error) {
	result := URLResourceInfo{}

	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return result, err
	}

	// Clean and split the path
	path := strings.Trim(parsedURL.Path, "/")
	if path == "" {
		result.Type = Unknown
		return result, nil
	}

	pathParts := strings.Split(path, "/")

	switch len(pathParts) {
	case 1:
		// https://github.com/username
		result.Type = User
		result.Owner = pathParts[0]

	case 2:
		// https://github.com/owner/repo
		result.Type = Repository
		result.Owner = pathParts[0]
		result.Repo = pathParts[1]

	case 4:
		// https://github.com/owner/repo/issues/123
		// https://github.com/owner/repo/pull/456
		result.Owner = pathParts[0]
		result.Repo = pathParts[1]

		if pathParts[2] == "issues" {
			result.Type = Issue
		} else if pathParts[2] == "pull" {
			result.Type = PullRequest
		} else {
			result.Type = Unknown
			return result, nil
		}

		// Parse the number
		if number, err := strconv.Atoi(pathParts[3]); err == nil {
			result.Number = number
		} else {
			result.Type = Unknown
			return result, err
		}

	default:
		// Other GitHub URLs (releases, commits, etc.)
		result.Type = Unknown
	}

	return result, nil
}

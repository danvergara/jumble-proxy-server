package github

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v74/github"
)

type ResourceType int

const (
	Unknown ResourceType = iota
	User
	Repository
	Issue
	PullRequest
	Release
	Commit
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
	case Release:
		return "release"
	case Commit:
		return "commit"
	default:
		return "unknown"
	}
}

type URLResourceInfo struct {
	Type    ResourceType
	Owner   string
	Repo    string
	Number  int // For issues and PRs
	SHA     string
	Version string
}

type GitHubResponse struct {
	Title     string
	Body      string
	imgageSrc string
}

type GithubClient struct {
	client *github.Client
}

func New(apiKey string) *GithubClient {
	c := github.NewClient(nil).WithAuthToken(apiKey)

	return &GithubClient{client: c}
}

// IsGitHubURL method determines if a given URL belongs to GitHub.
func (gc *GithubClient) IsGitHubURL(url string) bool {
	return strings.Contains(strings.ToLower(url), "github.com") ||
		strings.Contains(strings.ToLower(url), "api.github.com") ||
		strings.Contains(strings.ToLower(url), "gist.github.com")
}

// getURLResourceType determines the type of GitHub resource from a URL
func (gc *GithubClient) getResourceFromURL(rawURL string) (URLResourceInfo, error) {
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
		// https://github.com/owner/repo/commit/fe114c64733d850007f181bb029d9cc2237efe0f
		result.Owner = pathParts[0]
		result.Repo = pathParts[1]

		switch pathParts[2] {
		case "issues":
			result.Type = Issue
		case "pull":
			result.Type = PullRequest
		case "commit":
			result.Type = Commit
		default:
			result.Type = Unknown
			return result, nil
		}

		switch pathParts[2] {
		case "issues", "pull":
			if number, err := strconv.Atoi(pathParts[3]); err == nil {
				result.Number = number
			} else {
				result.Type = Unknown
				return result, err
			}
		case "commit":
			result.SHA = pathParts[3]
		}
		// Parse the number
	case 5:
		// https://github.com/owner/repo/releases/tag/v0.1.0
		result.Owner = pathParts[0]
		result.Repo = pathParts[1]

		if pathParts[2] == "releases" {
			result.Type = Release
		} else {
			result.Type = Unknown
			return result, nil
		}

		result.Version = pathParts[4]
	default:
		result.Type = Unknown
	}

	return result, nil
}

// queryGitHubResource returns the title/name and a description/bio of the resource, based on what it is asked for.
func (gc *GithubClient) queryGitHubResource(
	ctx context.Context,
	rawURL string,
) (GitHubResponse, error) {
	resp := GitHubResponse{}

	resourceInfo, err := gc.getResourceFromURL(rawURL)
	if err != nil {
		return resp, err
	}

	hash := time.Now().Unix()
	baseURL := fmt.Sprintf(
		"https://opengraph.githubassets.com/%d/%s/%s",
		hash,
		resourceInfo.Owner,
		resourceInfo.Repo,
	)

	switch resourceInfo.Type {
	case User:
		u, _, err := gc.client.Users.Get(ctx, resourceInfo.Owner)
		if err != nil {
			return resp, err
		}

		if u != nil {
			resp.Title = u.GetName()
			resp.Body = u.GetBio()
			resp.imgageSrc = u.GetAvatarURL()
		} else {
			return resp, fmt.Errorf("error getting the GitHub user %s", resourceInfo.Owner)
		}

		return resp, nil
	case Repository:
		repo, _, err := gc.client.Repositories.Get(ctx, resourceInfo.Owner, resourceInfo.Repo)
		if err != nil {
			return resp, err
		}

		if repo != nil {
			resp.Title = repo.GetFullName()
			resp.Body = repo.GetDescription()
			resp.imgageSrc = baseURL
		} else {
			return resp, fmt.Errorf("error getting the GitHub repository %s", resourceInfo.Repo)
		}

		return resp, nil
	case PullRequest:
		pr, _, err := gc.client.PullRequests.Get(
			ctx,
			resourceInfo.Owner,
			resourceInfo.Repo,
			resourceInfo.Number,
		)
		if err != nil {
			return resp, err
		}

		if pr != nil {
			resp.Title = pr.GetTitle()
			resp.Body = pr.GetBody()
			resp.imgageSrc = fmt.Sprintf("%s/pull/%d", baseURL, resourceInfo.Number)
		} else {
			return resp, fmt.Errorf("error getting the GitHub pull request #%d from %s repository", resourceInfo.Number, resourceInfo.Repo)
		}

		return resp, nil
	case Issue:
		issue, _, err := gc.client.Issues.Get(
			ctx,
			resourceInfo.Owner,
			resourceInfo.Repo,
			resourceInfo.Number,
		)
		if err != nil {
			return resp, err
		}

		if issue != nil {
			resp.Title = issue.GetTitle()
			resp.Body = issue.GetBody()
			resp.imgageSrc = fmt.Sprintf("%s/issues/%d", baseURL, resourceInfo.Number)
		} else {
			return resp, fmt.Errorf("error getting the GitHub issue #%d from %s repository", resourceInfo.Number, resourceInfo.Repo)
		}

		return resp, nil
	case Commit:
		commit, _, err := gc.client.Repositories.GetCommit(
			ctx,
			resourceInfo.Owner,
			resourceInfo.Repo,
			resourceInfo.SHA,
			nil,
		)
		if err != nil {
			return resp, err
		}

		if commit != nil {
			shortSHA := commit.GetSHA()
			if len(shortSHA) > 7 {
				shortSHA = shortSHA[:7]
			}
			description := fmt.Sprintf("%s/%s@%s", resourceInfo.Owner, resourceInfo.Repo, shortSHA)
			resp.Title = commit.GetCommit().GetMessage()
			resp.Body = description
			resp.imgageSrc = fmt.Sprintf("%s/commit/%s", baseURL, resourceInfo.SHA)
		} else {
			return resp, fmt.Errorf("error getting the GitHub commit %s from %s repository", resourceInfo.SHA, resourceInfo.Repo)
		}
		return resp, nil
	case Release:
		release, _, err := gc.client.Repositories.GetReleaseByTag(
			ctx,
			resourceInfo.Owner,
			resourceInfo.Repo,
			resourceInfo.Version,
		)
		if err != nil {
			return resp, err
		}

		if release != nil {
			resp.Title = fmt.Sprintf("%s %s", resourceInfo.Repo, release.GetTagName())
			resp.Body = release.GetBody()
			resp.imgageSrc = fmt.Sprintf("%s/releases/tag/%s", baseURL, resourceInfo.Version)
		} else {
			return resp, fmt.Errorf("error getting the GitHub release %s from %s repository", resourceInfo.Version, resourceInfo.Repo)
		}

		return resp, nil
	default:
		if resourceInfo.Owner != "" && resourceInfo.Repo != "" {
			repo, _, err := gc.client.Repositories.Get(ctx, resourceInfo.Owner, resourceInfo.Repo)
			if err != nil {
				return resp, err
			}

			if repo != nil {
				resp.Title = repo.GetFullName()
				resp.Body = repo.GetDescription()
				resp.imgageSrc = baseURL
			} else {
				return resp, fmt.Errorf("error getting the GitHub repository %s", resourceInfo.Repo)
			}

			return resp, nil
		} else {
			return resp, fmt.Errorf("resource type unknown %s", resourceInfo.Type)
		}
	}
}

func (gc *GithubClient) GenerateGithubOpenGraph(
	ctx context.Context,
	rawURL string,
) (string, error) {
	resp, err := gc.queryGitHubResource(ctx, rawURL)
	if err != nil {
		return "", err
	}

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>%s</title>
    <meta property="og:title" content="%s">
    <meta property="og:description" content="%s">
    <meta property="og:url" content="%s">
    <meta property="og:image" content="%s">
    <meta property="og:image:width" content="1200">
    <meta property="og:image:height" content="630">
    <meta property="og:type" content="website">
    <meta property="og:site_name" content="GitHub">
</head>
<body>
    <h1>%s</h1>
    <p>%s</p>
    <img src="%s" alt="Preview" style="max-width: 100%%; height: auto;">
</body>
</html>`,
		resp.Title, resp.Title, resp.Body, rawURL, resp.imgageSrc,
		resp.Title, resp.Body, resp.imgageSrc)

	return htmlContent, nil
}

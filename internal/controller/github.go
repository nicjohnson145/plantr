package controller

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"regexp"

	"github.com/carlmjohnson/requests"
	"github.com/go-git/go-billy/v5/helper/iofs"
	gmemfs "github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	ghttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/nicjohnson145/hlp"
	"github.com/rs/zerolog"
)

type GithubGitClientConfig struct {
	Logger zerolog.Logger
	Token  string
}

func NewGithubGitClient(conf GithubGitClientConfig) (*GithubGitClient, error) {
	if conf.Token == "" {
		return nil, fmt.Errorf("token is required")
	}

	return &GithubGitClient{
		log:   conf.Logger,
		token: conf.Token,
	}, nil
}

var _ GitClient = (*GithubGitClient)(nil)

type GithubGitClient struct {
	log    zerolog.Logger
	client *http.Client

	token string
}

func (g *GithubGitClient) parseUrl(url string) (string, string, error) {
	exp := regexp.MustCompile(`^(https://github.com/|git@github.com:)(?P<owner>[a-zA-Z0-9_\-]+)/(?P<repo>[a-zA-Z0-9_\-]+).git`)
	got := hlp.ExtractNamedMatches(exp, exp.FindStringSubmatch(url))
	if got["owner"] == "" {
		return "", "", fmt.Errorf("unable to extract owner from URL")
	}
	if got["repo"] == "" {
		return "", "", fmt.Errorf("unable to extract repo from URL")
	}
	return got["owner"], got["repo"], nil
}

func (g *GithubGitClient) GetLatestCommit(url string) (string, error) {
	owner, repo, err := g.parseUrl(url)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %w", err)
	}

	type latestCommit struct {
		SHA string `json:"sha"`
	}

	var resp []latestCommit
	var errResp map[string]any

	err = requests.
		URL("https://api.github.com").
		Pathf("repos/%v/%v/commits", owner, repo).
		Param("per_page", "1").
		Header("accept", "application/vnd.github+json").
		Header("Authorization", fmt.Sprintf("Bearer %v", g.token)).
		ToJSON(&resp).
		ErrorJSON(&errResp).
		Client(g.client).
		Fetch(context.Background())

	if err != nil {
		g.log.Debug().Interface("body", errResp).Msg("error response body")
		return "", fmt.Errorf("error querying commits: %w", err)
	}

	return resp[0].SHA, nil
}

func (g *GithubGitClient) CloneAtCommit(url string, commit string) (fs.FS, error) {
	bfs := gmemfs.New()
	r, err := git.Clone(memory.NewStorage(), bfs, &git.CloneOptions{
		Auth: &ghttp.BasicAuth{
			Username: "__token__",
			Password: g.token,
		},
		URL:   url,
		Depth: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("error cloning: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, fmt.Errorf("error getting worktree: %w", err)
	}

	if err := w.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(commit)}); err != nil {
		return nil, fmt.Errorf("error checking out commit: %w", err)
	}

	return iofs.New(bfs), nil
}

func (g *GithubGitClient) GetLatestRelease(url string) (string, error) {
	owner, repo, err := g.parseUrl(url)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %w", err)
	}

	type respType struct {
		TagName string `json:"tag_name"`
	}

	var resp respType
	var errResp map[string]any

	err = requests.
		URL("https://api.github.com").
		Pathf("repos/%v/%v/releases/latest", owner, repo).
		Header("accept", "application/vnd.github+json").
		Header("Authorization", fmt.Sprintf("Bearer %v", g.token)).
		ToJSON(&resp).
		ErrorJSON(&errResp).
		Fetch(context.Background())
	if err != nil {
		g.log.Debug().Interface("body", errResp).Msg("error response body")
		return "", fmt.Errorf("error getting latest release: %w", err)
	}

	return resp.TagName, nil
}

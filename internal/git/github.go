package git

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/carlmjohnson/requests"
	"github.com/rs/zerolog"
)

type GithubConfig struct {
	Logger zerolog.Logger

	Token string
	Owner string
	Repo  string
}

func NewGithub(conf GithubConfig) (*Github, error) {
	if conf.Owner == "" {
		return nil, fmt.Errorf("owner is required")
	}
	if conf.Repo == "" {
		return nil, fmt.Errorf("repo is required")
	}
	if conf.Token == "" {
		return nil, fmt.Errorf("token is required")
	}

	return &Github{
		log:   conf.Logger,
		owner: conf.Owner,
		repo:  conf.Repo,
		token: conf.Token,
	}, nil
}

type Github struct {
	log    zerolog.Logger
	client *http.Client

	token string
	owner string
	repo  string
}

type githubLatestCommitResponse struct {
	SHA string `json:"sha"`
}

func (g *Github) GetLatestCommit() (string, error) {
	var resp []githubLatestCommitResponse
	var errResp map[string]any

	err := requests.
		URL("https://api.github.com").
		Pathf("repos/%v/%v/commits", g.owner, g.repo).
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

func (g *Github) CloneAtCommit(commit string) (fs.FS, error) {
	panic("not implemented") // TODO: Implement
}

package git

import (
	"fmt"
	"io/fs"
	"regexp"

	"github.com/nicjohnson145/hlp"
	"github.com/nicjohnson145/plantr/internal/config"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var (
	githubURLPattern = regexp.MustCompile(`^(https://github.com/|git@github.com:)(?P<owner>[a-zA-Z0-9_\-]+)/(?P<repo>[a-zA-Z0-9_\-]+).git`)
)

type Client interface {
	GetLatestCommit() (string, error)
	CloneAtCommit(commit string) (fs.FS, error)
}

func NewFromEnv(logger zerolog.Logger) (Client, error) {
	kind, err := config.ParseGitKind(viper.GetString(config.GitType))
	if err != nil {
		return nil, err
	}

	switch kind {
	case config.GitKindGithub:
		owner, repo := parseGithubURL(viper.GetString(config.GitUrl))
		gh, err := NewGithub(GithubConfig{
			Logger: logger,
			Owner:  owner,
			Repo:   repo,
			Token:  viper.GetString(config.GitAccessToken),
		})
		if err != nil {
			return nil, fmt.Errorf("error initializing GitHub client: %w", err)
		}
		return gh, nil
	default:
		return nil, fmt.Errorf("unhandled kind '%v'", kind)
	}
}

func parseGithubURL(url string) (string, string) {
	got := hlp.ExtractNamedMatches(githubURLPattern, githubURLPattern.FindStringSubmatch(url))
	return got["owner"], got["repo"]
}

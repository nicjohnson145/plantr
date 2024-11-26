package controller

import (
	"fmt"
	"io/fs"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type GitClient interface {
	GetLatestCommit(url string) (string, error)
	CloneAtCommit(url string, commit string) (fs.FS, error)
	GetLatestRelease(url string) (string, error)
}

func NewGitFromEnv(logger zerolog.Logger) (GitClient, error) {
	kind, err := ParseGitKind(viper.GetString(GitType))
	if err != nil {
		return nil, err
	}

	switch kind {
	case GitKindGithub:
		gh, err := NewGithubGitClient(GithubGitClientConfig{
			Logger: logger,
			Token:  viper.GetString(GitAccessToken),
		})
		if err != nil {
			return nil, fmt.Errorf("error initializing GitHub client: %w", err)
		}
		return gh, nil
	case GitKindStatic:
		s, err := NewStaticGitClient(StaticGitClientConfig{
			Logger:       logger,
			CheckoutPath: viper.GetString(GitStaticCheckoutPath),
		})
		if err != nil {
			return nil, fmt.Errorf("error initializing static client: %w", err)
		}
		return s, nil
	default:
		return nil, fmt.Errorf("unhandled kind '%v'", kind)
	}
}

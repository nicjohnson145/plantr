package controller

import (
	"fmt"
	"io/fs"

	"github.com/nicjohnson145/plantr/internal/config"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type GitClient interface {
	GetLatestCommit(url string) (string, error)
	CloneAtCommit(url string, commit string) (fs.FS, error)
	GetLatestRelease(url string) (string, error)
}

func NewGitFromEnv(logger zerolog.Logger) (GitClient, error) {
	kind, err := config.ParseGitKind(viper.GetString(config.GitType))
	if err != nil {
		return nil, err
	}

	switch kind {
	case config.GitKindGithub:
		gh, err := NewGithubGitClient(GithubGitClientConfig{
			Logger: logger,
			Token:  viper.GetString(config.GitAccessToken),
		})
		if err != nil {
			return nil, fmt.Errorf("error initializing GitHub client: %w", err)
		}
		return gh, nil
	case config.GitKindStatic:
		s, err := NewStaticGitClient(StaticGitClientConfig{
			Logger:       logger,
			CheckoutPath: viper.GetString(config.GitStaticCheckoutPath),
		})
		if err != nil {
			return nil, fmt.Errorf("error initializing static client: %w", err)
		}
		return s, nil
	default:
		return nil, fmt.Errorf("unhandled kind '%v'", kind)
	}
}

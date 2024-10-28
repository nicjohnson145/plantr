package git

import (
	"fmt"
	"io/fs"

	"github.com/nicjohnson145/plantr/internal/config"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
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
		gh, err := NewGithub(GithubConfig{
			Logger: logger,
		})
		if err != nil {
			return nil, fmt.Errorf("error initializing GitHub client: %w", err)
		}
		return gh, nil
	default:
		return nil, fmt.Errorf("unhandled kind '%v'", kind)
	}
}

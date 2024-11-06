package git

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/rs/zerolog"
)

type StaticConfig struct {
	Logger       zerolog.Logger
	CheckoutPath string
}

func NewStatic(conf StaticConfig) (*Static, error) {
	if conf.CheckoutPath == "" {
		return nil, fmt.Errorf("must specify checkout path")
	}

	return &Static{
		log:          conf.Logger,
		checkoutPath: conf.CheckoutPath,
	}, nil
}

type Static struct {
	log          zerolog.Logger
	checkoutPath string
}

func (s *Static) GetLatestCommit(url string) (string, error) {
	return "not-a-real-commit", nil
}

func (s *Static) CloneAtCommit(url string, commit string) (fs.FS, error) {
	return os.DirFS(s.checkoutPath), nil
}

func (s *Static) GetLatestRelease(url string) (string, error) {
	return "fake-tag", nil
}

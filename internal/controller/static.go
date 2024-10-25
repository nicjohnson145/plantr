package controller

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/rs/zerolog"
)

type StaticGitClientConfig struct {
	Logger       zerolog.Logger
	CheckoutPath string
}

func NewStaticGitClient(conf StaticGitClientConfig) (*StaticGitClient, error) {
	if conf.CheckoutPath == "" {
		return nil, fmt.Errorf("must specify checkout path")
	}

	return &StaticGitClient{
		log:          conf.Logger,
		checkoutPath: conf.CheckoutPath,
	}, nil
}

type StaticGitClient struct {
	log          zerolog.Logger
	checkoutPath string
}

func (s *StaticGitClient) GetLatestCommit(url string) (string, error) {
	return "not-a-real-commit", nil
}

func (s *StaticGitClient) CloneAtCommit(url string, commit string) (fs.FS, error) {
	return os.DirFS(s.checkoutPath), nil
}

func (s *StaticGitClient) GetLatestRelease(url string) (string, error) {
	return "fake-tag", nil
}

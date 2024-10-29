package controller

import (
	"context"
	"errors"
	"fmt"

	"sync"

	"connectrpc.com/connect"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/v1"
	"github.com/nicjohnson145/plantr/internal/git"
	"github.com/nicjohnson145/plantr/internal/parsing"
	"github.com/nicjohnson145/plantr/internal/storage"
	"github.com/rs/zerolog"
)

var (
	ErrNoNodeIDError = errors.New("node_id is required")
)

type ControllerConfig struct {
	Logger        zerolog.Logger
	GitClient     git.Client
	StorageClient storage.Client
	RepoURL       string
	JWTSigningKey []byte
}

func NewController(conf ControllerConfig) (*Controller, error) {
	ctrl := &Controller{
		log:           conf.Logger,
		git:           conf.GitClient,
		store:         conf.StorageClient,
		repoUrl:       conf.RepoURL,
		jwtSigningKey: conf.JWTSigningKey,
	}

	return ctrl, nil
}

type Controller struct {
	log           zerolog.Logger
	git           git.Client
	store         storage.Client
	repoUrl       string
	jwtSigningKey []byte

	mu     *sync.RWMutex
	config *pbv1.Config
}

func (c *Controller) logAndHandleError(err error, msg string) error {
	str := "an error occurred"
	if msg != "" {
		str = msg
	}

	c.log.Err(err).Msg(str)

	switch true {
	case errors.Is(err, ErrNoNodeIDError):
		return connect.NewError(connect.CodeInvalidArgument, err)
	default:
		return err
	}
}

func (c *Controller) ensureConfig() error {
	if c.config != nil {
		c.log.Debug().Msg("config already present, nothing to do")
		return nil
	}

	c.log.Debug().Msg("no config loaded, loading now")

	c.log.Trace().Msg("fetching latest commit")
	latest, err := c.git.GetLatestCommit(c.repoUrl)
	if err != nil {
		return fmt.Errorf("error getting latest commit: %w", err)
	}

	c.log.Trace().Msg("cloning repo")
	repoFS, err := c.git.CloneAtCommit(c.repoUrl, latest)
	if err != nil {
		return fmt.Errorf("error cloning repo: %w", err)
	}

	config, err := parsing.ParseRepoFS(repoFS)
	if err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}

	c.mu.Lock()
	c.config = config
	c.mu.Unlock()

	return nil
}

func (c *Controller) Login(ctx context.Context, req *connect.Request[pbv1.LoginRequest]) (*connect.Response[pbv1.LoginResponse], error) {
	if err := c.validateLogin(req.Msg); err != nil {
		return nil, c.logAndHandleError(err, "error validating")
	}

	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("method unimplemented"))
}

func (c *Controller) validateLogin(req *pbv1.LoginRequest) error {
	if req.NodeId == "" {
		return ErrNoNodeIDError
	}

	return nil
}

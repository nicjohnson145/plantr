package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"sync"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt"
	hsqlx "github.com/nicjohnson145/hlp/sqlx"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/v1"
	"github.com/nicjohnson145/plantr/internal/encryption"
	"github.com/nicjohnson145/plantr/internal/git"
	"github.com/nicjohnson145/plantr/internal/interceptors"
	"github.com/nicjohnson145/plantr/internal/parsing"
	"github.com/nicjohnson145/plantr/internal/storage"
	"github.com/nicjohnson145/plantr/internal/token"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
)

var (
	ErrNoNodeIDError                = errors.New("node_id is required")
	ErrNoChallengeIDError           = errors.New("challenge_id required")
	ErrNoChallengeValueError        = errors.New("challenge_value required")
	ErrUnknownNodeIDError           = errors.New("unknown node_id")
	ErrUnknownChallengeIDError      = errors.New("unknown challenge_id")
	ErrIncorrectChallengeValueError = errors.New("incorrect challenge_value")
)

type ControllerConfig struct {
	Logger        zerolog.Logger
	GitClient     git.Client
	StorageClient storage.Client
	RepoURL       string
	JWTSigningKey []byte
	JWTDuration   time.Duration

	NowFunc func() time.Time // for unit tests
}

func NewController(conf ControllerConfig) (*Controller, error) {
	repoUrl := conf.RepoURL
	if !strings.HasSuffix(repoUrl, ".git") {
		repoUrl = repoUrl + ".git"
	}
	ctrl := &Controller{
		log:           conf.Logger,
		git:           conf.GitClient,
		store:         conf.StorageClient,
		repoUrl:       repoUrl,
		jwtSigningKey: conf.JWTSigningKey,
		jwtDuration:   conf.JWTDuration,
		nowFunc:       conf.NowFunc,
		mu:            &sync.RWMutex{},
	}

	if ctrl.nowFunc == nil {
		ctrl.nowFunc = func() time.Time {
			return time.Now().UTC()
		}
	}

	return ctrl, nil
}

type Controller struct {
	log           zerolog.Logger
	git           git.Client
	store         storage.Client
	repoUrl       string
	jwtSigningKey []byte
	jwtDuration   time.Duration

	mu     *sync.RWMutex
	config *pbv1.Config

	nowFunc func() time.Time // for unit tests
}

func (c *Controller) now() time.Time {
	return c.nowFunc()
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
	case errors.Is(err, ErrNoChallengeIDError):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, ErrNoChallengeValueError):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, ErrUnknownNodeIDError):
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("permission denied"))
	case errors.Is(err, ErrUnknownChallengeIDError):
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("permission denied"))
	case errors.Is(err, ErrIncorrectChallengeValueError):
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("permission denied"))
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

	c.log.Trace().Msgf("fetching latest commit for %v", c.repoUrl)
	latest, err := c.git.GetLatestCommit(c.repoUrl)
	if err != nil {
		return fmt.Errorf("error getting latest commit: %w", err)
	}

	c.log.Trace().Msg("cloning repo")
	repoFS, err := c.git.CloneAtCommit(c.repoUrl, latest)
	if err != nil {
		return fmt.Errorf("error cloning repo: %w", err)
	}

	c.log.Trace().Msg("parsing config from cloned repo")
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

	// Get the node that we're supposed to be logging in
	if err := c.ensureConfig(); err != nil {
		return nil, c.logAndHandleError(err, "error ensuring config")
	}

	var node *pbv1.Node
	for _, n := range c.config.Nodes {
		if n.Id == req.Msg.NodeId {
			node = n
		}
	}
	if node == nil {
		return nil, c.logAndHandleError(ErrUnknownNodeIDError, "unable to find node matching ID")
	}

	// If we've only got a node id, create a challenge and send it back
	if req.Msg.ChallengeId == nil {
		challengeID := ulid.Make().String()
		challengeValue := ulid.Make().String()

		encryptedValue, err := encryption.EncryptValue(challengeValue, node.PublicKey)
		if err != nil {
			return nil, c.logAndHandleError(err, "error encrypting challenge")
		}

		challenge := &storage.Challenge{
			ID:    challengeID,
			Value: challengeValue,
		}
		if err := c.store.WriteChallenge(ctx, challenge); err != nil {
			return nil, c.logAndHandleError(err, "error inserting challenge")
		}

		return connect.NewResponse(&pbv1.LoginResponse{
			ChallengeId:     &challengeID,
			SealedChallenge: &encryptedValue,
		}), nil
	}

	// Otherwise, lets validate their challenge response and issue them a token
	storedChallenge, err := c.store.ReadChallenge(ctx, *req.Msg.ChallengeId)
	if err != nil {
		if errors.Is(err, hsqlx.ErrNotFoundError) {
			return nil, c.logAndHandleError(ErrUnknownChallengeIDError, "challenge id not found")
		}
		return nil, c.logAndHandleError(err, "error reading challenge")
	}
	if storedChallenge.Value != *req.Msg.ChallengeValue {
		return nil, c.logAndHandleError(ErrIncorrectChallengeValueError, "incorrect challenge value")
	}

	token, err := token.GenerateJWT(c.jwtSigningKey, token.Token{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: c.now().Add(c.jwtDuration).Unix(),
		},
		NodeID: req.Msg.NodeId,
	})
	if err != nil {
		return nil, c.logAndHandleError(err, "error generating JWT")
	}

	return connect.NewResponse(&pbv1.LoginResponse{
		Token: &token,
	}), nil
}

func (c *Controller) validateLogin(req *pbv1.LoginRequest) error {
	if req.NodeId == "" {
		return ErrNoNodeIDError
	}

	if req.ChallengeId != nil && req.ChallengeValue == nil {
		return ErrNoChallengeValueError
	}

	if req.ChallengeValue != nil && req.ChallengeId == nil {
		return ErrNoChallengeIDError
	}

	return nil
}

func (c *Controller) GetSyncData(ctx context.Context, req *connect.Request[pbv1.GetSyncDataRequest]) (*connect.Response[pbv1.GetSyncDataReponse], error) {
	token, err := interceptors.ClaimsFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	c.log.Info().Msgf("parsing sync data for %v", token.NodeID)
	if err := c.ensureConfig(); err != nil {
		return nil, c.logAndHandleError(err, "error ensuring config")
	}

	return connect.NewResponse(&pbv1.GetSyncDataReponse{
		Seeds: []*pbv1.Seed{},
	}), nil
}

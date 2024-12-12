package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"connectrpc.com/connect"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/agent/v1"
	controllerv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/nicjohnson145/plantr/gen/plantr/controller/v1/controllerv1connect"
	"github.com/nicjohnson145/plantr/internal/encryption"
	"github.com/nicjohnson145/plantr/internal/interceptors"
	"github.com/rs/zerolog"
)

var (
	ErrSyncInProgressError = errors.New("sync already in progress")
)

type AgentConfig struct {
	Logger            zerolog.Logger
	NodeID            string
	PrivateKey        string
	ControllerAddress string
	NowFunc           func() time.Time
	HTTPClient        *http.Client
}

func NewAgent(conf AgentConfig) *Agent {
	a := &Agent{
		log:               conf.Logger,
		nodeID:            conf.NodeID,
		privateKey:        conf.PrivateKey,
		controllerAddress: conf.ControllerAddress,
		mu:                &sync.Mutex{},
		nowFunc:           conf.NowFunc,
		httpClient:        conf.HTTPClient,
	}

	if a.nowFunc == nil {
		a.nowFunc = func() time.Time {
			return time.Now().UTC()
		}
	}

	return a
}

type Agent struct {
	log               zerolog.Logger
	nodeID            string
	privateKey        string
	controllerAddress string
	mu                *sync.Mutex

	token           string
	tokenExpiration time.Time
	nowFunc         func() time.Time
	httpClient      *http.Client
}

func (a *Agent) logAndHandleError(err error, msg string) error {
	str := "an error occurred"
	if msg != "" {
		str = msg
	}

	a.log.Err(err).Msg(str)

	switch true {
	default:
		return err
	}
}

func (a *Agent) Sync(req *pbv1.SyncRequest) (*pbv1.SyncResponse, error) {
	if !a.mu.TryLock() {
		return nil, ErrSyncInProgressError
	}
	defer a.mu.Unlock()

	a.log.Info().Msg("beginning sync")

	a.log.Debug().Msg("fetching sync data")
	client := controllerv1connect.NewControllerServiceClient(http.DefaultClient, a.controllerAddress)
	token, err := a.getAccessToken(client)
	if err != nil {
		return nil, a.logAndHandleError(err, "error getting access token")
	}

	client = controllerv1connect.NewControllerServiceClient(
		http.DefaultClient,
		a.controllerAddress,
		connect.WithInterceptors(interceptors.NewClientAuthInterceptor(token)),
	)
	resp, err := client.GetSyncData(context.Background(), connect.NewRequest(&controllerv1.GetSyncDataRequest{}))
	if err != nil {
		return nil, a.logAndHandleError(err, "error getting sync data")
	}

	if err := a.executeSeeds(resp.Msg.Seeds); err != nil {
		return nil, a.logAndHandleError(err, "error executing seeds")
	}

	a.log.Info().Msg("sync completed successfully")
	return nil, nil
}

func (a *Agent) getAccessToken(client controllerv1connect.ControllerServiceClient) (string, error) {
	if a.token != "" && a.tokenExpiration.After(a.nowFunc().Add(5*time.Minute)) {
		a.log.Debug().Msg("token still valid, reusing")
		return a.token, nil
	}

	a.log.Debug().Msg("token missing or close/after expiration, attempting login")
	resp, err := client.Login(context.Background(), connect.NewRequest(&controllerv1.LoginRequest{
		NodeId: a.nodeID,
	}))
	if err != nil {
		return "", fmt.Errorf("error getting login challenge: %w", err)
	}

	challenge, err := encryption.DecryptValue(*resp.Msg.SealedChallenge, a.privateKey)
	if err != nil {
		return "", fmt.Errorf("error decrypting login challenge: %w", err)
	}

	resp, err = client.Login(context.Background(), connect.NewRequest(&controllerv1.LoginRequest{
		NodeId:         a.nodeID,
		ChallengeId:    resp.Msg.ChallengeId,
		ChallengeValue: &challenge,
	}))
	if err != nil {
		return "", fmt.Errorf("error getting login token: %w", err)
	}

	a.token = *resp.Msg.Token
	// TODO: actually unpack token and get its expiration
	a.tokenExpiration = a.nowFunc().Add(24 * time.Hour)

	return a.token, nil
}

func (a *Agent) executeSeeds(seeds []*controllerv1.Seed) error {
	var errs []error

	for _, seed := range seeds {
		switch concrete := seed.Element.(type) {
		case *controllerv1.Seed_ConfigFile:
			if err := a.executeSeed_configFile(concrete.ConfigFile); err != nil {
				errs = append(errs, fmt.Errorf("error executing config file: %w", err))
			}
		case *controllerv1.Seed_GithubRelease:
			if err := a.executeSeed_githubRelease(concrete.GithubRelease); err != nil {
				errs = append(errs, fmt.Errorf("error executing github release: %w", err))
			}
		default:
			a.log.Warn().Msgf("dropping unknown seed type %T", concrete)
		}
	}

	return errors.Join(errs...)
}

func (a *Agent) executeSeed_configFile(seed *controllerv1.ConfigFile) error {
	// TODO: inventory tracking
	if err := os.MkdirAll(filepath.Dir(seed.Destination), 0775); err != nil {
		return fmt.Errorf("error creating containing dir: %w", err)
	}
	// TODO: configurable permissions
	if err := os.WriteFile(seed.Destination, []byte(seed.Content), 0644); err != nil { //nolint:gosec // ignore until configurable permissions
		return fmt.Errorf("error creating file: %w", err)
	}
	return nil
}

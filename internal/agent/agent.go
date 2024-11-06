package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/v1"
	"github.com/nicjohnson145/plantr/gen/plantr/v1/plantrv1connect"
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
}

func NewAgent(conf AgentConfig) *Agent {
	a := &Agent{
		log:               conf.Logger,
		nodeID:            conf.NodeID,
		privateKey:        conf.PrivateKey,
		controllerAddress: conf.ControllerAddress,
		mu:                &sync.Mutex{},
		nowFunc:           conf.NowFunc,
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
	client := plantrv1connect.NewControllerServiceClient(http.DefaultClient, a.controllerAddress)
	token, err := a.getAccessToken(client)
	if err != nil {
		return nil, a.logAndHandleError(err, "error getting access token")
	}

	client = plantrv1connect.NewControllerServiceClient(
		http.DefaultClient,
		a.controllerAddress,
		connect.WithInterceptors(interceptors.NewClientAuthInterceptor(token)),
	)
	resp, err := client.GetSyncData(context.Background(), connect.NewRequest(&pbv1.GetSyncDataRequest{}))
	if err != nil {
		return nil, a.logAndHandleError(err, "error getting sync data")
	}

	_ = resp

	a.log.Info().Msg("sync completed successfully")
	return nil, nil
}

func (a *Agent) getAccessToken(client plantrv1connect.ControllerServiceClient) (string, error) {
	if a.token != "" && a.tokenExpiration.After(a.nowFunc().Add(5*time.Minute)) {
		a.log.Debug().Msg("token still valid, reusing")
		return a.token, nil
	}

	a.log.Debug().Msg("token missing or close/after expiration, attempting login")
	resp, err := client.Login(context.Background(), connect.NewRequest(&pbv1.LoginRequest{
		NodeId: a.nodeID,
	}))
	if err != nil {
		return "", fmt.Errorf("error getting login challenge: %w", err)
	}

	challenge, err := encryption.DecryptValue(*resp.Msg.SealedChallenge, a.privateKey)
	if err != nil {
		return "", fmt.Errorf("error decrypting login challenge: %w", err)
	}

	resp, err = client.Login(context.Background(), connect.NewRequest(&pbv1.LoginRequest{
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

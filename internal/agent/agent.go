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
}

func NewAgent(conf AgentConfig) *Agent {
	return &Agent{
		log:               conf.Logger,
		nodeID:            conf.NodeID,
		privateKey:        conf.PrivateKey,
		controllerAddress: conf.ControllerAddress,
		mu:                &sync.Mutex{},
	}
}

type Agent struct {
	log               zerolog.Logger
	nodeID            string
	privateKey        string
	controllerAddress string
	mu                *sync.Mutex

	token           string
	tokenExpiration time.Time
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

	_ = token

	a.log.Info().Msg("sync completed successfully")
	return nil, nil
}

func (a *Agent) getAccessToken(client plantrv1connect.ControllerServiceClient) (string, error) {
	return "", nil
}

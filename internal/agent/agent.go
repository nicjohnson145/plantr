package agent

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/v1"
	"github.com/rs/zerolog"
)

type AgentConfig struct {
	Logger zerolog.Logger
}

func NewAgent(conf AgentConfig) *Agent {
	return &Agent{
		log: conf.Logger,
	}
}

type Agent struct {
	log zerolog.Logger
}

func (a *Agent) Sync(ctx context.Context, req *connect.Request[pbv1.SyncRequest]) (*connect.Response[pbv1.SyncResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not done yet"))
}

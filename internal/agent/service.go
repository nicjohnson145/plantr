package agent

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/v1"
	"github.com/rs/zerolog"
)

type ServiceConfig struct {
	Logger zerolog.Logger
	Agent  *Agent
}

func NewService(conf ServiceConfig) *Service {
	return &Service{
		log:   conf.Logger,
		agent: conf.Agent,
	}
}

type Service struct {
	log   zerolog.Logger
	agent *Agent
}

func (s *Service) logAndHandleError(err error, msg string) error {
	str := "an error occurred"
	if msg != "" {
		str = msg
	}

	s.log.Err(err).Msg(str)

	switch true {
	case errors.Is(err, ErrSyncInProgressError):
		return connect.NewError(connect.CodeUnavailable, err)
	default:
		return err
	}
}

func (s *Service) Sync(ctx context.Context, req *connect.Request[pbv1.SyncRequest]) (*connect.Response[pbv1.SyncResponse], error) {
	resp, err := s.agent.Sync(req.Msg)
	if err != nil {
		return nil, s.logAndHandleError(err, "error syncing")
	}
	return connect.NewResponse(resp), nil
}

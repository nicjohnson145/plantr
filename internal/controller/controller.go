package controller

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/v1"
	"github.com/rs/zerolog"
)

type ControllerConfig struct {
	Logger zerolog.Logger
}

func NewController(conf ControllerConfig) *Controller {
	return &Controller{
		log: conf.Logger,
	}
}

type Controller struct {
	log zerolog.Logger
}

func (c *Controller) Login(ctx context.Context, req *connect.Request[pbv1.LoginRequest]) (*connect.Response[pbv1.LoginResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("method unimplemented"))
}

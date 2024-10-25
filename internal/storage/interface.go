package storage

import (
	"context"
)

type Client interface {
	RegisterHost(ctx context.Context, host *Host) (*Host, error)
}


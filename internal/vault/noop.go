package vault

import (
	"github.com/rs/zerolog"
	"github.com/oklog/ulid/v2"
)

type NoopConfig struct {
	Logger zerolog.Logger
}

func NewNoop(conf NoopConfig) *Noop {
	return &Noop{
		log: conf.Logger,
	}
}

type Noop struct {
	log zerolog.Logger
}

func (n *Noop) GetSecretVersion() (string, error) {
	n.log.Debug().Msg("noop vault client, returning random secret version")
	return ulid.Make().String(), nil
}

func (n *Noop) ReadSecretData() (map[string]any, error) {
	n.log.Debug().Msg("noop vault client, returning static secret data")
	return map[string]any{
		"foo": "static-foo-value",
	}, nil
}


package controller

import (
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
)

type NoopVaultConfig struct {
	Logger zerolog.Logger
}

func NewNoopVault(conf NoopVaultConfig) *NoopVault {
	return &NoopVault{
		log: conf.Logger,
	}
}

var _ VaultClient = (*NoopVault)(nil)

type NoopVault struct {
	log zerolog.Logger
}

func (n *NoopVault) GetSecretVersion() (string, error) {
	n.log.Debug().Msg("noop vault client, returning random secret version")
	return ulid.Make().String(), nil
}

func (n *NoopVault) ReadSecretData() (map[string]any, error) {
	n.log.Debug().Msg("noop vault client, returning static secret data")
	return map[string]any{
		"foo": "static-foo-value",
	}, nil
}

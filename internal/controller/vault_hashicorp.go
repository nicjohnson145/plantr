package controller

import (
	"github.com/rs/zerolog"
)

type HashicorpVaultConfig struct {
	Logger zerolog.Logger
}

func NewHashicorpVault(conf HashicorpVaultConfig) *HashicorpVault {
	return &HashicorpVault{
		log: conf.Logger,
	}
}

var _ VaultClient = (*HashicorpVault)(nil)

type HashicorpVault struct {
	log zerolog.Logger
}

func (h *HashicorpVault) GetSecretVersion() (string, error) {
	panic("not implemented") // TODO: Implement
}

func (h *HashicorpVault) ReadSecretData() (map[string]any, error) {
	panic("not implemented") // TODO: Implement
}

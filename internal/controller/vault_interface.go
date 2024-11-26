package controller

import (
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type VaultClient interface {
	GetSecretVersion() (string, error)
	ReadSecretData() (map[string]any, error)
}

func NewVaultFromEnv(logger zerolog.Logger) (VaultClient, error) {
	if !viper.GetBool(VaultEnabled) {
		return NewNoopVault(NoopVaultConfig{
			Logger: logger,
		}), nil
	}

	return NewHashicorpVault(HashicorpVaultConfig{
		Logger: logger,
	}), nil
}

package controller

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type VaultClient interface {
	ReadSecretData(ctx context.Context) (map[string]any, error)
}

func NewVaultFromEnv(logger zerolog.Logger) (VaultClient, error) {
	if !viper.GetBool(VaultEnabled) {
		return NewNoopVault(NoopVaultConfig{
			Logger: logger,
		}), nil
	}

	return NewHashicorpVault(HashicorpVaultConfig{
		Logger:     logger,
		Address:    viper.GetString(VaultHashicorpAddress),
		Username:   viper.GetString(VaultHashicorpUsername),
		Password:   viper.GetString(VaultHashicorpPassword),
		SecretPath: viper.GetString(VaultHashicorpSecretPath),
		TokenTTL:   viper.GetDuration(VaultHashicorpTTL),
	}), nil
}

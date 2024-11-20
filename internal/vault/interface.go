package vault

import (
	"github.com/nicjohnson145/plantr/internal/config"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type Client interface {
	GetSecretVersion() (string, error)
	ReadSecretData() (map[string]any, error)
}

func NewFromEnv(logger zerolog.Logger) (Client, error) {
	if !viper.GetBool(config.VaultEnabled) {
		return NewNoop(NoopConfig{
			Logger: logger,
		}), nil
	}

	return NewHashicorp(HashicorpConfig{
		Logger: logger,
	}), nil
}

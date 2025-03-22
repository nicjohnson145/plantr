package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/rs/zerolog"
)

type HashicorpVaultConfig struct {
	Logger     zerolog.Logger
	Address    string
	Username   string
	Password   string
	SecretPath string
	TokenTTL   time.Duration
	Now        func() time.Time
}

func NewHashicorpVault(conf HashicorpVaultConfig) *HashicorpVault {
	now := conf.Now
	if now == nil {
		now = func() time.Time {
			return time.Now()
		}
	}

	return &HashicorpVault{
		log:        conf.Logger,
		address:    conf.Address,
		username:   conf.Username,
		password:   conf.Password,
		secretPath: conf.SecretPath,

		tokenTTL: conf.TokenTTL,
		now:      now,
	}
}

var _ VaultClient = (*HashicorpVault)(nil)

type HashicorpVault struct {
	log        zerolog.Logger
	address    string
	username   string
	password   string
	secretPath string

	tokenTTL time.Duration
	now      func() time.Time

	client         *vault.Client
	tokenGenerated time.Time
}

func (h *HashicorpVault) getClient(ctx context.Context) (*vault.Client, error) {
	if h.client == nil {
		h.log.Debug().Msg("nil client, creating one")
		client, err := vault.New(
			vault.WithAddress(h.address),
			vault.WithRequestTimeout(5*time.Second),
		)
		if err != nil {
			return nil, fmt.Errorf("error creating vault client: %w", err)
		}

		h.client = client
	}

	now := h.now()
	if now.After(h.tokenGenerated.Add(h.tokenTTL)) {
		h.log.Debug().Msg("token TTL expired, re-authenticating")
		resp, err := h.client.Auth.UserpassLogin(ctx, h.username, schema.UserpassLoginRequest{Password: h.password})
		if err != nil {
			return nil, fmt.Errorf("error authenticating to vault: %w", err)
		}
		if err := h.client.SetToken(resp.Auth.ClientToken); err != nil {
			return nil, fmt.Errorf("error setting client token: %w", err)
		}
		h.tokenGenerated = now
	}

	return h.client, nil
}

func (h *HashicorpVault) ReadSecretData(ctx context.Context) (map[string]any, error) {
	client, err := h.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting vault client: %w", err)
	}

	h.log.Debug().Msg("executing vault read")
	resp, err := client.Read(ctx, h.secretPath)
	if err != nil {
		return nil, fmt.Errorf("error getting secret data: %w", err)
	}

	return resp.Data["data"].(map[string]any), nil
}

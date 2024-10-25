package controller

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHashicorpVaultBenchTests(t *testing.T) {
	if os.Getenv("HASHICORP_BENCH_TESTS") == "" {
		t.Skipf("skipping bench tests due to HASHICORP_BENCH_TESTS not set")
	}

	newVault := func() *HashicorpVault {
		return NewHashicorpVault(HashicorpVaultConfig{
			Address:    "CHANGE_ME",
			Username:   "CHANGE_ME",
			Password:   "CHANGE_ME",
			SecretPath: "CHANGE_ME",
			TokenTTL:   1 * time.Hour,
		})
	}

	t.Run("get secret version", func(t *testing.T) {
		vault := newVault()
		resp, err := vault.ReadSecretData(context.Background())
		require.NoError(t, err)
		require.Equal(t, map[string]any{"foo": "vault-foo-value"}, resp)
	})
}

package controller

import (
	"context"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestStorageIntegration(t *testing.T) {
	execute := func(t *testing.T) {
		store, cleanup, err := NewStorageClientFromEnv(zerolog.New(os.Stdout).Level(zerolog.Disabled))
		require.NoError(t, err)
		t.Cleanup(cleanup)

		ctx := context.Background()

		const (
			challengeID    = "some-challenge-id"
			challengeValue = "some-challenge-value"
		)

		// Write a challenge
		require.NoError(t, store.WriteChallenge(ctx, &Challenge{
			ID: challengeID,
			Value: challengeValue,
		}))

		// Read that challenge back
		gotChallenge, err := store.ReadChallenge(ctx, challengeID)
		require.NoError(t, err)
		require.Equal(t, challengeValue, gotChallenge.Value)
	}

	t.Run("sqlite", func(t *testing.T) {
		t.Cleanup(func() {
			viper.Reset()
		})
		tmp := t.TempDir()
		viper.Set(StorageType, StorageKindSqlite)
		viper.Set(SqliteDBPath, tmp+"/controller.db")

		execute(t)
	})
}

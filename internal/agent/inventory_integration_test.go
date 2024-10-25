package agent

import (
	"context"
	"os"
	"testing"

	"github.com/nicjohnson145/hlp"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestInventoryIntegration(t *testing.T) {
	execute := func(t *testing.T) {
		store, cleanup, err := NewInventoryClientFromEnv(zerolog.New(os.Stdout).Level(zerolog.Disabled))
		require.NoError(t, err)
		t.Cleanup(cleanup)

		ctx := context.Background()

		const (
			hash = "some-hash-string"
			path = "some-path"
		)

		// Read a row thats not there
		row, err := store.GetRow(ctx, hash)
		require.NoError(t, err)
		require.Nil(t, row)

		// Write a row
		require.NoError(t, store.WriteRow(ctx, InventoryRow{
			Hash: hash,
			Path: hlp.Ptr(path),
		}))

		// Read row back
		row, err = store.GetRow(ctx, hash)
		require.NoError(t, err)
		require.Equal(
			t,
			&InventoryRow{
				Hash: hash,
				Path: hlp.Ptr(path),
			},
			row,
		)
	}

	t.Run("sqlite", func(t *testing.T) {
		t.Cleanup(func() {
			viper.Reset()
		})
		tmp := t.TempDir()
		viper.Set(StorageType, StorageKindSqlite)
		viper.Set(SqliteDBPath, tmp+"/agent.db")

		execute(t)
	})
}

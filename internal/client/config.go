package client

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

//go:generate go-enum -f $GOFILE -marshal -names

/*
ENUM(
sqlite
none
)
*/
type StorageKind string

const (
	StorageType  = "storage.type"
	SqliteDBPath = "sqlite.db_path"
)

var (
	DefaultStorageType = StorageKindSqlite.String()
)

func SetConfigDefaults() error {
	cachedir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("error getting user cache dir: %w", err)
	}

	viper.SetDefault(StorageType, DefaultStorageType)
	viper.SetDefault(SqliteDBPath, filepath.Join(cachedir, "plantr", "storage.db"))

	return nil
}

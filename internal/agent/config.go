package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicjohnson145/plantr/internal/logging"
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
	Port = "port"

	LoggingLevel  = "log.level"
	LoggingFormat = "log.format"
	LogRequests   = "log.requests"
	LogResponses  = "log.responses"

	ControllerAddress = "controller.address"
	PrivateKeyPath    = "private_key.path"
	NodeID            = "node.id"
	PollInterval      = "poll_interval"

	StorageType  = "storage.type"
	SqliteDBPath = "sqlite.db_path"
)

var (
	DefaultPort = "8080"

	DefaultLogLevel     = logging.LogLevelInfo.String()
	DefaultLogFormat    = logging.LogFormatJson.String()
	DefaultLogRequests  = false
	DefaultLogResponses = false

	DefaultStorageType  = StorageKindSqlite.String()

	DefaultPollInterval = "0s"
)

func InitConfig() error {
	cachedir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("error getting user cache dir: %w", err)
	}

	viper.SetDefault(Port, DefaultPort)

	viper.SetDefault(LoggingLevel, DefaultLogLevel)
	viper.SetDefault(LoggingFormat, DefaultLogFormat)
	viper.SetDefault(LogRequests, DefaultLogRequests)
	viper.SetDefault(LogResponses, DefaultLogResponses)

	viper.SetDefault(PollInterval, DefaultPollInterval)

	viper.SetDefault(StorageType, DefaultStorageType)
	viper.SetDefault(SqliteDBPath, filepath.Join(cachedir, "plantr", "storage.db"))

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return nil
}

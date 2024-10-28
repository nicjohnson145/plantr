package config

import (
	"strings"

	"github.com/spf13/viper"
)

//go:generate go-enum -f $GOFILE -marshal -names

/*
ENUM(
sqlite
)
*/
type StorageKind string

const (
	Port = "port"

	LoggingLevel  = "log.level"
	LoggingFormat = "log.format"

	StorageType = "storage.type"
)

var (
	DefaultPort = "8080"

	DefaultLogLevel  = LogLevelInfo.String()
	DefaultLogFormat = LogFormatJson.String()

	DefaultStorageType = StorageKindSqlite.String()
)

func InitConfig() {
	viper.SetDefault(Port, DefaultPort)

	viper.SetDefault(LoggingLevel, DefaultLogLevel)
	viper.SetDefault(LoggingFormat, DefaultLogFormat)

	viper.SetDefault(StorageType, DefaultStorageType)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

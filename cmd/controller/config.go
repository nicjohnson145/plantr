package main

import (
	"strings"

	"github.com/nicjohnson145/plantr/internal/logging"
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

	LogLevel  = "log.level"
	LogFormat = "log.format"

	StorageType = "storage.type"
)

var (
	DefaultPort = "8080"

	DefaultLogLevel  = logging.LogLevelInfo.String()
	DefaultLogFormat = logging.LogFormatJson.String()

	DefaultStorageType = StorageKindSqlite.String()
)

func InitConfig() {
	viper.SetDefault(Port, DefaultPort)

	viper.SetDefault(LogLevel, DefaultLogLevel)
	viper.SetDefault(LogFormat, DefaultLogFormat)

	viper.SetDefault(StorageType, DefaultStorageType)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

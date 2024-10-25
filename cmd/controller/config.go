package main

import (
	"strings"

	"github.com/nicjohnson145/plantr/internal/logging"
	"github.com/spf13/viper"
)

const (
	Port = "port"

	LogLevel  = "log.level"
	LogFormat = "log.format"

)

var (
	DefaultPort      = "8080"

	DefaultLogLevel  = logging.LogLevelInfo.String()
	DefaultLogFormat = logging.LogFormatJson.String()
)

func InitConfig() {
	viper.SetDefault(Port, DefaultPort)

	viper.SetDefault(LogLevel, DefaultLogLevel)
	viper.SetDefault(LogFormat, DefaultLogFormat)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

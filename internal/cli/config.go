package cli

import (
	"strings"

	"github.com/nicjohnson145/plantr/internal/logging"
	"github.com/spf13/viper"
)


const (
	LoggingLevel  = "log.level"
	LoggingFormat = "log.format"
)

var (
	DefaultLogLevel     = logging.LogLevelInfo.String()
	DefaultLogFormat    = logging.LogFormatHuman.String()
)

func InitConfig() {
	viper.SetDefault(LoggingLevel, DefaultLogLevel)
	viper.SetDefault(LoggingFormat, DefaultLogFormat)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

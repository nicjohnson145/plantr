package cli

import (
	"strings"

	"github.com/nicjohnson145/plantr/internal/logging"
	"github.com/spf13/viper"
)

const (
	LoggingLevel  = "log.level"
	LoggingFormat = "log.format"

	InitControllerAddress = "init.controller_address"
	InitNodeID            = "init.node_id"
	InitUserHome          = "init.user_home"
	InitPackageManager    = "init.package_manager"
	InitPublicKeyPath     = "init.public_key_path"
)

var (
	DefaultLogLevel  = logging.LogLevelInfo.String()
	DefaultLogFormat = logging.LogFormatHuman.String()
)

func InitConfig() {
	viper.SetDefault(LoggingLevel, DefaultLogLevel)
	viper.SetDefault(LoggingFormat, DefaultLogFormat)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

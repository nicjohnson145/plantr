package agent

import (
	"strings"

	"github.com/nicjohnson145/plantr/internal/client"
	"github.com/nicjohnson145/plantr/internal/logging"
	"github.com/spf13/viper"
)

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
)

var (
	DefaultPort = "8080"

	DefaultLogLevel     = logging.LogLevelInfo.String()
	DefaultLogFormat    = logging.LogFormatJson.String()
	DefaultLogRequests  = false
	DefaultLogResponses = false

	DefaultPollInterval = "0s"
)

func SetServiceDefaults() {
	viper.SetDefault(Port, DefaultPort)

	viper.SetDefault(LoggingLevel, DefaultLogLevel)
	viper.SetDefault(LoggingFormat, DefaultLogFormat)
	viper.SetDefault(LogRequests, DefaultLogRequests)
	viper.SetDefault(LogResponses, DefaultLogResponses)

	viper.SetDefault(PollInterval, DefaultPollInterval)
}

func InitConfig() error {
	SetServiceDefaults()
	if err := client.SetConfigDefaults(); err != nil {
		return err
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return nil
}

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicjohnson145/plantr/internal/agent"
	"github.com/nicjohnson145/plantr/internal/logging"
	"github.com/spf13/viper"
)

const (
	LoggingLevel  = "log.level"
	LoggingFormat = "log.format"
)

var (
	DefaultLogLevel  = logging.LogLevelInfo.String()
	DefaultLogFormat = logging.LogFormatHuman.String()
)

func InitConfig() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("error getting user config directory: %w", err)
	}

	viper.AddConfigPath(filepath.Join(configDir, "plantr"))
	viper.SetConfigType("yaml")
	viper.SetConfigName("cli")

	viper.SetDefault(LoggingLevel, DefaultLogLevel)
	viper.SetDefault(LoggingFormat, DefaultLogFormat)

	if err := agent.SetWorkerDefaults(); err != nil {
		return err
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))


	if err := viper.ReadInConfig(); err != nil && !errors.Is(err, &viper.ConfigFileNotFoundError{}) {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil
		}
		return fmt.Errorf("error reading config file: %s", err)
	}

	return nil
}

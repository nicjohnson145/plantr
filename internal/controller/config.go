package controller

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

/*
ENUM(
github
static
)
*/
type GitKind string

const (
	Port = "port"

	LoggingLevel  = "log.level"
	LoggingFormat = "log.format"
	LogRequests   = "log.requests"
	LogResponses  = "log.responses"

	StorageType = "storage.type"

	SqliteDBPath = "sqlite.db_path"

	GitType        = "git.type"
	GitAccessToken = "git.access_token"
	GitUrl         = "git.url"

	GitStaticCheckoutPath = "git.static.checkout_path"

	JWTSigningKey = "jwt.signing_key"
	JWTDuration   = "jwt.duration"

	VaultEnabled          = "vault.enabled"
	VaultHashicorpAddress = "vault.hashicorp.address"
)

var (
	DefaultPort = "8080"

	DefaultLogLevel     = logging.LogLevelInfo.String()
	DefaultLogFormat    = logging.LogFormatJson.String()
	DefaultLogRequests  = false
	DefaultLogResponses = false

	DefaultStorageType = StorageKindSqlite.String()

	DefaultSqliteDBPath = "/var/plantr/storage.db"

	DefaultGitType = GitKindGithub.String()

	DefaultJWTDuration = "240h" // 10 days

	DefaultAgentPollInterval = "60s"

	DefaultVaultEnabled = false
)

func InitConfig() {
	viper.SetDefault(Port, DefaultPort)

	viper.SetDefault(LoggingLevel, DefaultLogLevel)
	viper.SetDefault(LoggingFormat, DefaultLogFormat)
	viper.SetDefault(LogRequests, DefaultLogRequests)
	viper.SetDefault(LogResponses, DefaultLogResponses)

	viper.SetDefault(StorageType, DefaultStorageType)

	viper.SetDefault(SqliteDBPath, DefaultSqliteDBPath)

	viper.SetDefault(GitType, DefaultGitType)

	viper.SetDefault(JWTDuration, DefaultJWTDuration)

	viper.SetDefault(VaultEnabled, DefaultVaultEnabled)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

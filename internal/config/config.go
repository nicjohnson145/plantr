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

/*
ENUM(
github
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

	JWTSigningKey = "jwt.signing_key"
	JWTDuration   = "jwt.duration"
)

var (
	DefaultPort = "8080"

	DefaultLogLevel     = LogLevelInfo.String()
	DefaultLogFormat    = LogFormatJson.String()
	DefaultLogRequests  = false
	DefaultLogResponses = false

	DefaultStorageType = StorageKindSqlite.String()

	DefaultSqliteDBPath = "/var/plantr/storage.db"

	DefaultGitType = GitKindGithub.String()

	DefaultJWTDuration = "240h" // 10 days
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

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

package main

import (
	"net/http"
	"os"

	"github.com/nicjohnson145/plantr/gen/plantr/v1/plantrv1connect"
	"github.com/nicjohnson145/plantr/internal/controller"
	"github.com/nicjohnson145/plantr/internal/logging"
	"github.com/spf13/viper"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	InitConfig()

	logger := logging.Init(&logging.LoggingConfig{
		Level:  logging.LogLevel(viper.GetString(LogLevel)),
		Format: logging.LogFormat(viper.GetString(LogFormat)),
	})

	ctrl := controller.NewController(controller.ControllerConfig{
		Logger: logger,
	})

	mux := http.NewServeMux()
	mux.Handle(plantrv1connect.NewControllerHandler(ctrl))

	port := viper.GetString(Port)
	logger.Info().Msgf("starting server on port %v", port)
	if err := http.ListenAndServe(":"+port, h2c.NewHandler(mux, &http2.Server{})); err != nil {
		logger.Err(err).Msg("error serving")
		os.Exit(1)
	}
}

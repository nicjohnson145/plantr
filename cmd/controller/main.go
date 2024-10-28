package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nicjohnson145/plantr/gen/plantr/v1/plantrv1connect"
	"github.com/nicjohnson145/plantr/internal/config"
	"github.com/nicjohnson145/plantr/internal/controller"
	"github.com/nicjohnson145/plantr/internal/storage"
	"github.com/spf13/viper"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"connectrpc.com/grpcreflect"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	config.InitConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := config.Init(&config.LoggingConfig{
		Level:  config.LogLevel(viper.GetString(config.LoggingLevel)),
		Format: config.LogFormat(viper.GetString(config.LoggingFormat)),
	})

	storage, storageCleanup, err := storage.NewFromEnv(logger)
	defer storageCleanup()
	if err != nil {
		logger.Err(err).Msg("error initializing storage client")
		return err
	}

	_ = storage

	// Reflection
	reflector := grpcreflect.NewStaticReflector(
		plantrv1connect.ControllerName,
	)

	ctrl := controller.NewController(controller.ControllerConfig{
		Logger: logger,
	})

	mux := http.NewServeMux()
	mux.Handle(plantrv1connect.NewControllerHandler(ctrl))
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	port := viper.GetString(config.Port)
	lis, err := net.Listen("tcp4", ":" + port)
	if err != nil {
		logger.Err(err).Msg("error listening")
		return err
	}

	svr := http.Server{
		Addr:    ":" + port,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	// Setup signal handlers so we can gracefully shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		select {
		case s := <-sigChan:
			logger.Info().Msgf("got signal %v, attempting graceful shutdown", s)
			dieCtx, dieCancel := context.WithTimeout(ctx, 10*time.Second)
			defer dieCancel()
			svr.Shutdown(dieCtx)
			cancel()
			wg.Done()
		}
	}()

	logger.Info().Msgf("starting server on port %v", port)
	if err := svr.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Err(err).Msg("error serving")
		return err
	}

	return nil
}

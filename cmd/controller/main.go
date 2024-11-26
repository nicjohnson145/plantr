package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/nicjohnson145/hlp/set"
	"github.com/nicjohnson145/plantr/gen/plantr/controller/v1/controllerv1connect"
	"github.com/nicjohnson145/plantr/internal/config"
	"github.com/nicjohnson145/plantr/internal/controller"
	"github.com/nicjohnson145/plantr/internal/interceptors"
	"github.com/nicjohnson145/plantr/internal/vault"
	"github.com/spf13/viper"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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

	// JWT bits
	jwtKeyStr := viper.GetString(config.JWTSigningKey)
	if jwtKeyStr == "" {
		logger.Error().Msg("must provide JWT signing key")
		return fmt.Errorf("must provide JWT signing key")
	}

	storage, storageCleanup, err := controller.NewStorageClientFromEnv(config.Component(logger, "storage"))
	defer storageCleanup()
	if err != nil {
		logger.Err(err).Msg("error initializing storage client")
		return err
	}

	gitClient, err := controller.NewGitFromEnv(config.Component(logger, "git"))
	if err != nil {
		logger.Err(err).Msg("error initializing git client")
		return err
	}

	vaultClient, err := vault.NewFromEnv(config.Component(logger, "vault"))
	if err != nil {
		logger.Err(err).Msg("error initializing vault client")
		return err
	}

	// Reflection
	reflector := grpcreflect.NewStaticReflector(
		controllerv1connect.ControllerServiceName,
	)

	// Get the root configuration for the repo
	url := viper.GetString(config.GitUrl)
	if url == "" {
		logger.Error().Msg("must provide a repo url")
		return fmt.Errorf("must provide a repo url")
	}

	ctrl, err := controller.NewController(controller.ControllerConfig{
		Logger:        config.Component(logger, "service"),
		StorageClient: storage,
		GitClient:     gitClient,
		RepoURL:       url,
		JWTSigningKey: []byte(jwtKeyStr),
		JWTDuration:   viper.GetDuration(config.JWTDuration),
		VaultClient:   vaultClient,
	})
	if err != nil {
		logger.Err(err).Msg("error initializing controller")
		return err
	}

	mux := http.NewServeMux()
	mux.Handle(controllerv1connect.NewControllerServiceHandler(
		ctrl,
		connect.WithInterceptors(
			interceptors.NewLoggingInterceptor(
				logger,
				interceptors.LoggingInterceptorConfig{
					LogRequests:  viper.GetBool(config.LogRequests),
					LogResponses: viper.GetBool(config.LogResponses),
				},
			),
			interceptors.NewAuthInterceptor(
				logger,
				[]byte(jwtKeyStr),
				set.New(
					"/plantr.controller.v1.ControllerService/Login",
				),
			),
		),
	))
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	port := viper.GetString(config.Port)
	lis, err := net.Listen("tcp4", ":"+port)
	if err != nil {
		logger.Err(err).Msg("error listening")
		return err
	}

	svr := http.Server{
		Addr:              ":" + port,
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: 3 * time.Second,
	}

	// Setup signal handlers so we can gracefully shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		s := <-sigChan
		logger.Info().Msgf("got signal %v, attempting graceful shutdown", s)
		dieCtx, dieCancel := context.WithTimeout(ctx, 10*time.Second)
		defer dieCancel()
		_ = svr.Shutdown(dieCtx)
		cancel()
		wg.Done()
	}()

	logger.Info().Msgf("starting server on port %v", port)
	if err := svr.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Err(err).Msg("error serving")
		return err
	}

	return nil
}

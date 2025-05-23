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
	"github.com/nicjohnson145/plantr/internal/controller"
	"github.com/nicjohnson145/plantr/internal/interceptors"
	"github.com/nicjohnson145/plantr/internal/logging"
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
	controller.InitConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.Init(&logging.LoggingConfig{
		Level:  logging.LogLevel(viper.GetString(controller.LoggingLevel)),
		Format: logging.LogFormat(viper.GetString(controller.LoggingFormat)),
	})

	// JWT bits
	jwtKeyStr := viper.GetString(controller.JWTSigningKey)
	if jwtKeyStr == "" {
		logger.Error().Msg("must provide JWT signing key")
		return fmt.Errorf("must provide JWT signing key")
	}

	storage, storageCleanup, err := controller.NewStorageClientFromEnv(logging.Component(logger, "storage"))
	defer storageCleanup()
	if err != nil {
		logger.Err(err).Msg("error initializing storage client")
		return err
	}

	gitClient, err := controller.NewGitFromEnv(logging.Component(logger, "git"))
	if err != nil {
		logger.Err(err).Msg("error initializing git client")
		return err
	}

	vaultClient, err := controller.NewVaultFromEnv(logging.Component(logger, "vault"))
	if err != nil {
		logger.Err(err).Msg("error initializing vault client")
		return err
	}

	// Reflection
	reflector := grpcreflect.NewStaticReflector(
		controllerv1connect.ControllerServiceName,
	)

	// Get the root configuration for the repo
	url := viper.GetString(controller.GitUrl)
	if url == "" {
		logger.Error().Msg("must provide a repo url")
		return fmt.Errorf("must provide a repo url")
	}

	ctrl, err := controller.NewController(controller.ControllerConfig{
		Logger:              logging.Component(logger, "service"),
		StorageClient:       storage,
		GitClient:           gitClient,
		RepoURL:             url,
		JWTSigningKey:       []byte(jwtKeyStr),
		JWTDuration:         viper.GetDuration(controller.JWTDuration),
		VaultClient:         vaultClient,
		GithubReleaseToken:  viper.GetString(controller.GitAccessToken),
		GithubWebhookSecret: []byte(viper.GetString(controller.GithubWebhookSecret)),
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
					LogRequests:  viper.GetBool(controller.LogRequests),
					LogResponses: viper.GetBool(controller.LogResponses),
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
	mux.Handle(ctrl.NewGithubWebhookHandler())

	port := viper.GetString(controller.Port)
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

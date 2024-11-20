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

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	agentv1 "github.com/nicjohnson145/plantr/gen/plantr/agent/v1"
	"github.com/nicjohnson145/plantr/gen/plantr/agent/v1/agentv1connect"
	"github.com/nicjohnson145/plantr/internal/agent"
	"github.com/nicjohnson145/plantr/internal/config"
	"github.com/nicjohnson145/plantr/internal/interceptors"
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

	// Reflection
	reflector := grpcreflect.NewStaticReflector(
		agentv1connect.AgentServiceName,
	)

	// Actual sync worker
	keyPath := viper.GetString(config.PrivateKeyPath)
	if keyPath == "" {
		msg := "private key path must be set"
		logger.Error().Msg(msg)
		return errors.New(msg)
	}
	privateKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		logger.Err(err).Msg("error reading private key")
		return err
	}

	controllerAddress := viper.GetString(config.ControllerAddress)
	if controllerAddress == "" {
		msg := "controller address must be set"
		logger.Error().Msg(msg)
		return errors.New(msg)
	}

	nodeID := viper.GetString(config.NodeID)
	if nodeID == "" {
		msg := "node id must be set"
		logger.Error().Msg(msg)
		return errors.New(msg)
	}

	pollInterval := viper.GetDuration(config.AgentPollInterval)
	if pollInterval.Seconds() == 0 {
		msg := "poll interval must be set"
		logger.Error().Msg(msg)
		return errors.New(msg)
	}

	worker := agent.NewAgent(agent.AgentConfig{
		Logger:            logger.With().Str("component", "agent-worker").Logger(),
		NodeID:            nodeID,
		ControllerAddress: controllerAddress,
		PrivateKey:        string(privateKeyBytes),
	})

	srv := agent.NewService(agent.ServiceConfig{
		Logger: logger.With().Str("component", "service").Logger(),
		Agent:  worker,
	})

	mux := http.NewServeMux()
	mux.Handle(agentv1connect.NewAgentServiceHandler(
		srv,
		connect.WithInterceptors(
			interceptors.NewLoggingInterceptor(
				logger,
				interceptors.LoggingInterceptorConfig{
					LogRequests:  viper.GetBool(config.LogRequests),
					LogResponses: viper.GetBool(config.LogResponses),
				},
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
	wg.Add(2)

	go func() {
		logger.Info().Msgf("starting periodic sync loop with frequency of %v", pollInterval)
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_, err := worker.Sync(&agentv1.SyncRequest{})
				if err != nil {
					if errors.Is(err, agent.ErrSyncInProgressError) {
						logger.Info().Msg("periodic sync aborted, sync already in progress")
					} else {
						logger.Err(err).Msg("error during periodic sync")
					}
				}
			case <-ctx.Done():
				logger.Info().Msg("context cancelled, ending periodic sync loop")
				wg.Done()
				return
			}
		}
	}()

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

	wg.Wait()
	return nil
}

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
	agentv1 "github.com/nicjohnson145/plantr/gen/plantr/agent/v1"
	"github.com/nicjohnson145/plantr/gen/plantr/agent/v1/agentv1connect"
	"github.com/nicjohnson145/plantr/internal/agent"
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
	if err := agent.InitConfig(); err != nil {
		newErr := fmt.Errorf("error initializing config: %w", err)
		fmt.Println(newErr)
		return newErr
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.Init(&logging.LoggingConfig{
		Level:  logging.LogLevel(viper.GetString(agent.LoggingLevel)),
		Format: logging.LogFormat(viper.GetString(agent.LoggingFormat)),
	})

	// Reflection
	reflector := grpcreflect.NewStaticReflector(
		agentv1connect.AgentServiceName,
	)

	// Actual sync worker
	keyPath := viper.GetString(agent.PrivateKeyPath)
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

	controllerAddress := viper.GetString(agent.ControllerAddress)
	if controllerAddress == "" {
		msg := "controller address must be set"
		logger.Error().Msg(msg)
		return errors.New(msg)
	}

	nodeID := viper.GetString(agent.NodeID)
	if nodeID == "" {
		msg := "node id must be set"
		logger.Error().Msg(msg)
		return errors.New(msg)
	}

	worker := agent.NewAgent(agent.AgentConfig{
		Logger:            logging.Component(logger, "agent-worker"),
		NodeID:            nodeID,
		ControllerAddress: controllerAddress,
		PrivateKey:        string(privateKeyBytes),
	})

	srv := agent.NewService(agent.ServiceConfig{
		Logger: logging.Component(logger, "service"),
		Agent:  worker,
	})

	mux := http.NewServeMux()
	mux.Handle(agentv1connect.NewAgentServiceHandler(
		srv,
		connect.WithInterceptors(
			interceptors.NewLoggingInterceptor(
				logger,
				interceptors.LoggingInterceptorConfig{
					LogRequests:  viper.GetBool(agent.LogRequests),
					LogResponses: viper.GetBool(agent.LogResponses),
				},
			),
		),
	))
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	port := viper.GetString(agent.Port)
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

	pollInterval := viper.GetDuration(agent.PollInterval)
	if pollInterval.Seconds() == 0 {
		logger.Info().Msg("poll internal set to 0s, disabling background worker")
	} else {
		wg.Add(1)
		go func() {
			logger.Info().Msgf("starting periodic sync loop with frequency of %v", pollInterval)
			ticker := time.NewTicker(pollInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					_, err := worker.Sync(context.TODO(), &agentv1.SyncRequest{})
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
	}

	wg.Add(1)
	go func() {
		select {
		case s := <-sigChan:
			logger.Info().Msgf("got signal %v, attempting graceful shutdown", s)
			dieCtx, dieCancel := context.WithTimeout(ctx, 10*time.Second)
			defer dieCancel()
			_ = svr.Shutdown(dieCtx)
			cancel()
			wg.Done()
		case <-ctx.Done():
			wg.Done()
		}
	}()

	logger.Info().Msgf("starting server on port %v", port)
	if err := svr.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Err(err).Msg("error serving")
		return err
	}

	wg.Wait()
	return nil
}

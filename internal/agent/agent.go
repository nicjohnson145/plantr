package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/nicjohnson145/hlp"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/agent/v1"
	controllerv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/nicjohnson145/plantr/gen/plantr/controller/v1/controllerv1connect"
	"github.com/nicjohnson145/plantr/internal/encryption"
	"github.com/nicjohnson145/plantr/internal/interceptors"
	"github.com/rs/zerolog"
)

var (
	ErrSyncInProgressError = errors.New("sync already in progress")
)

type AgentConfig struct {
	Logger            zerolog.Logger
	NodeID            string
	PrivateKey        string
	ControllerAddress string
	NowFunc           func() time.Time
	HTTPClient        *http.Client
	Inventory         InventoryClient
}

func NewAgent(conf AgentConfig) *Agent {
	a := &Agent{
		log:               conf.Logger,
		nodeID:            conf.NodeID,
		privateKey:        conf.PrivateKey,
		controllerAddress: conf.ControllerAddress,
		mu:                &sync.Mutex{},
		nowFunc:           conf.NowFunc,
		httpClient:        conf.HTTPClient,
		inventory:         conf.Inventory,
	}

	if a.nowFunc == nil {
		a.nowFunc = func() time.Time {
			return time.Now().UTC()
		}
	}

	return a
}

type Agent struct {
	log               zerolog.Logger
	nodeID            string
	privateKey        string
	controllerAddress string
	mu                *sync.Mutex

	token           string
	tokenExpiration time.Time
	nowFunc         func() time.Time
	httpClient      *http.Client
	inventory       InventoryClient
}

func (a *Agent) logAndHandleError(err error, msg string) error {
	str := "an error occurred"
	if msg != "" {
		str = msg
	}

	a.log.Err(err).Msg(str)

	switch true {
	default:
		return err
	}
}

func (a *Agent) Sync(ctx context.Context, req *pbv1.SyncRequest) (*pbv1.SyncResponse, error) {
	if !a.mu.TryLock() {
		return nil, ErrSyncInProgressError
	}
	defer a.mu.Unlock()

	a.log.Info().Msg("beginning sync")

	a.log.Debug().Msg("fetching sync data")
	client := controllerv1connect.NewControllerServiceClient(http.DefaultClient, a.controllerAddress)
	token, err := a.getAccessToken(client)
	if err != nil {
		return nil, a.logAndHandleError(err, "error getting access token")
	}

	client = controllerv1connect.NewControllerServiceClient(
		http.DefaultClient,
		a.controllerAddress,
		connect.WithInterceptors(interceptors.NewClientAuthInterceptor(token)),
	)
	resp, err := client.GetSyncData(context.Background(), connect.NewRequest(&controllerv1.GetSyncDataRequest{}))
	if err != nil {
		return nil, a.logAndHandleError(err, "error getting sync data")
	}

	if err := a.executeSeeds(ctx, resp.Msg.Seeds); err != nil {
		return nil, a.logAndHandleError(err, "error executing seeds")
	}

	a.log.Info().Msg("sync completed successfully")
	return nil, nil
}

func (a *Agent) getAccessToken(client controllerv1connect.ControllerServiceClient) (string, error) {
	if a.token != "" && a.tokenExpiration.After(a.nowFunc().Add(5*time.Minute)) {
		a.log.Debug().Msg("token still valid, reusing")
		return a.token, nil
	}

	a.log.Debug().Msg("token missing or close/after expiration, attempting login")
	resp, err := client.Login(context.Background(), connect.NewRequest(&controllerv1.LoginRequest{
		NodeId: a.nodeID,
	}))
	if err != nil {
		return "", fmt.Errorf("error getting login challenge: %w", err)
	}

	challenge, err := encryption.DecryptValue(*resp.Msg.SealedChallenge, a.privateKey)
	if err != nil {
		return "", fmt.Errorf("error decrypting login challenge: %w", err)
	}

	resp, err = client.Login(context.Background(), connect.NewRequest(&controllerv1.LoginRequest{
		NodeId:         a.nodeID,
		ChallengeId:    resp.Msg.ChallengeId,
		ChallengeValue: &challenge,
	}))
	if err != nil {
		return "", fmt.Errorf("error getting login token: %w", err)
	}

	a.token = *resp.Msg.Token
	// TODO: actually unpack token and get its expiration
	a.tokenExpiration = a.nowFunc().Add(24 * time.Hour)

	return a.token, nil
}

func (a *Agent) executeSeeds(ctx context.Context, seeds []*controllerv1.Seed) error {
	var errs []error

	for _, seed := range seeds {
		switch concrete := seed.Element.(type) {
		case *controllerv1.Seed_ConfigFile:
			if err := a.executeSeed_configFile(ctx, concrete.ConfigFile, seed.Metadata); err != nil {
				errs = append(errs, fmt.Errorf("error executing config file: %w", err))
			}
		case *controllerv1.Seed_GithubRelease:
			if err := a.executeSeed_githubRelease(ctx, concrete.GithubRelease, seed.Metadata); err != nil {
				errs = append(errs, fmt.Errorf("error executing github release: %w", err))
			}
		case *controllerv1.Seed_SystemPackage:
			if err := a.executeSeed_systemPackage(ctx, concrete.SystemPackage, seed.Metadata); err != nil {
				errs = append(errs, fmt.Errorf("error executing system package: %w", err))
			}
		case *controllerv1.Seed_GitRepo:
			if err := a.executeSeed_gitRepo(ctx, concrete.GitRepo, seed.Metadata); err != nil {
				errs = append(errs, fmt.Errorf("error executing git repo: %w", err))
			}
		default:
			a.log.Warn().Msgf("dropping unknown seed type %T", concrete)
		}
	}

	return errors.Join(errs...)
}

func (a *Agent) executeSeed_configFile(ctx context.Context, seed *controllerv1.ConfigFile, metadata *controllerv1.Seed_Metadata) error {
	row, err := a.inventory.GetRow(ctx, metadata.Hash)
	if err != nil {
		return fmt.Errorf("error checking inventory: %w", err)
	}
	if row != nil {
		a.log.Debug().Msg("config file found in inventory, skipping")
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(seed.Destination), 0775); err != nil {
		return fmt.Errorf("error creating containing dir: %w", err)
	}
	// TODO: configurable permissions
	if err := os.WriteFile(seed.Destination, []byte(seed.Content), 0644); err != nil { //nolint:gosec // ignore until configurable permissions
		return fmt.Errorf("error creating file: %w", err)
	}

	err = a.inventory.WriteRow(ctx, InventoryRow{
		Hash: metadata.Hash,
		Path: hlp.Ptr(seed.Destination),
	})
	if err != nil {
		return fmt.Errorf("error writing inventory record: %w", err)
	}

	return nil
}

func (a *Agent) executeSeed_systemPackage(ctx context.Context, seed *controllerv1.SystemPackage, metadata *controllerv1.Seed_Metadata) error {
	row, err := a.inventory.GetRow(ctx, metadata.Hash)
	if err != nil {
		return fmt.Errorf("error checking inventory: %w", err)
	}
	if row != nil {
		a.log.Debug().Msg("system package found in inventory, skipping")
		return nil
	}

	// Individual functions are responsible for writing
	switch concrete := seed.Pkg.(type) {
	case *controllerv1.SystemPackage_Apt:
		return a.executeSeed_systemPackage_apt(ctx, concrete.Apt, metadata)
	default:
		return fmt.Errorf("unhandled system package type of %T", concrete)
	}
}

func (a *Agent) executeSeed_systemPackage_apt(ctx context.Context, pkg *controllerv1.SystemPackage_AptPkg, metadata *controllerv1.Seed_Metadata) error {
	// TODO: proper version support & `apt update` cached for the whole run
	_, stderr, err := ExecuteOSCommand("/bin/sh", "-c", fmt.Sprintf("sudo DEBIAN_FRONTEND=noninteractive apt install -y %v", pkg.Name))
	if err != nil {
		return fmt.Errorf("error during installation: %v\n%v", err, stderr)
	}

	err = a.inventory.WriteRow(ctx, InventoryRow{
		Hash:    metadata.Hash,
		Package: hlp.Ptr(pkg.Name),
	})
	if err != nil {
		return fmt.Errorf("error writing to inventory: %w", err)
	}

	return nil
}

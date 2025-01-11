package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/carlmjohnson/requests"
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


	noopSkip := func(_ *controllerv1.Seed) bool {
		return false
	}

	for _, seed := range seeds {
		var executeFunc func(context.Context, *controllerv1.Seed) (*InventoryRow, error)
		var skipInventoryFunc func(*controllerv1.Seed) bool
		// TODO: generate this server-side & pack it into metadata
		var msg string

		switch concrete := seed.Element.(type) {
		case *controllerv1.Seed_ConfigFile:
			msg = fmt.Sprintf("rendering config file %v", concrete.ConfigFile.Destination)
			executeFunc = a.executeSeed_configFile
			skipInventoryFunc = noopSkip
		case *controllerv1.Seed_GithubRelease:
			// TODO: pass through the repo name for logging purposes
			msg = fmt.Sprintf("downloading github_release %v", concrete.GithubRelease.DownloadUrl)
			executeFunc = a.executeSeed_githubRelease
			skipInventoryFunc = noopSkip
		case *controllerv1.Seed_SystemPackage:
			// TODO: clever way to get the name?
			msg = "installing system_package"
			executeFunc = a.executeSeed_systemPackage
			skipInventoryFunc = noopSkip
		case *controllerv1.Seed_GitRepo:
			msg = fmt.Sprintf("cloning git_repo %v", concrete.GitRepo.Url)
			executeFunc = a.executeSeed_gitRepo
			skipInventoryFunc = noopSkip
		case *controllerv1.Seed_Golang:
			msg = fmt.Sprintf("install go@%v", concrete.Golang.Version)
			executeFunc = a.executeSeed_golang
			skipInventoryFunc = noopSkip
		case *controllerv1.Seed_GoInstall:
			msg = fmt.Sprintf("installing go binary %v", concrete.GoInstall.Package)
			executeFunc = a.executeSeed_goInstall
			// If we're not specifying a version, that means "latest", so dont check inventory to guarantee that we try
			// it again
			skipInventoryFunc = func(s *controllerv1.Seed) bool {
				return s.GetGoInstall().Version == nil
			}
		default:
			a.log.Warn().Msgf("dropping unknown seed type %T", concrete)
			continue
		}

		a.log.Info().Msg(msg)

		if !skipInventoryFunc(seed) {
			row, err := a.inventory.GetRow(ctx, seed.Metadata.Hash)
			if err != nil {
				return fmt.Errorf("error reading inventory: %w", err)
			}
			if row != nil {
				a.log.Debug().Msg("already exists in inventory, skipping")
				continue
			}
		}
		row, err := executeFunc(ctx, seed)
		if err != nil {
			errs = append(errs, fmt.Errorf("error executing: %w", err))
			continue
		}


		if row != nil {
			row.Hash = seed.Metadata.Hash
			if err := a.inventory.WriteRow(ctx, *row); err != nil {
				errs = append(errs, fmt.Errorf("error writing to inventory: %w", err))
				continue
			}
		}
	}

	return errors.Join(errs...)
}

func (a *Agent) executeSeed_configFile(ctx context.Context, pbseed *controllerv1.Seed) (*InventoryRow, error) {
	seed := pbseed.Element.(*controllerv1.Seed_ConfigFile).ConfigFile

	if err := os.MkdirAll(filepath.Dir(seed.Destination), 0775); err != nil {
		return nil, fmt.Errorf("error creating containing dir: %w", err)
	}
	// TODO: configurable permissions
	if err := os.WriteFile(seed.Destination, []byte(seed.Content), 0644); err != nil { //nolint:gosec // ignore until configurable permissions
		return nil, fmt.Errorf("error creating file: %w", err)
	}

	return &InventoryRow{
		Path: hlp.Ptr(seed.Destination),
	}, nil
}

func (a *Agent) executeSeed_systemPackage(ctx context.Context, pbseed *controllerv1.Seed) (*InventoryRow, error) {
	seed := pbseed.Element.(*controllerv1.Seed_SystemPackage).SystemPackage

	switch concrete := seed.Pkg.(type) {
	case *controllerv1.SystemPackage_Apt:
		return a.executeSeed_systemPackage_apt(ctx, concrete.Apt)
	default:
		return nil, fmt.Errorf("unhandled system package type of %T", concrete)
	}
}

func (a *Agent) executeSeed_systemPackage_apt(_ context.Context, pkg *controllerv1.SystemPackage_AptPkg) (*InventoryRow, error) {
	// TODO: proper version support & `apt update` cached for the whole run
	_, stderr, err := ExecuteOSCommand("/bin/sh", "-c", fmt.Sprintf("sudo DEBIAN_FRONTEND=noninteractive apt install -y %v", pkg.Name))
	if err != nil {
		return nil, fmt.Errorf("error during installation: %v\n%v", err, stderr)
	}

	return &InventoryRow{
		Package: hlp.Ptr(pkg.Name),
	}, nil
}

func (a *Agent) executeSeed_golang(_ context.Context, seed *controllerv1.Seed) (*InventoryRow, error) {
	golang := seed.Element.(*controllerv1.Seed_Golang).Golang

	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("golang install only available for linux OS")
	}

	a.log.Trace().Msg("removing existing installation")
	// make sure to clean out the old version first per the golang docs. Run this command through the shell so we can
	// elivate privileges
	_, _, err := ExecuteOSCommand("/bin/sh", "-c", "sudo rm -rf /usr/local/go")
	if err != nil {
		return nil, fmt.Errorf("error removing old golang installation: %w", err)
	}

	a.log.Trace().Msg("downloading release tarball")
	dir, err := os.MkdirTemp("", "plantr-golang-")
	if err != nil {
		return nil, fmt.Errorf("unable to make temp directory")
	}
	defer os.RemoveAll(dir)

	tarball := fmt.Sprintf("go%v.linux-%v.tar.gz", golang.Version, runtime.GOARCH)
	filepath := filepath.Join(dir, tarball)
	err = requests.
		URL(fmt.Sprintf("https://go.dev/dl/%v", tarball)).
		ToFile(filepath).
		Fetch(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error downloading tarball: %w", err)
	}

	a.log.Trace().Msg("extracting tarball")
	// Execute this through the shell so we can elevate privileges with sudo
	_, _, err = ExecuteOSCommand("/bin/sh", "-c", fmt.Sprintf("sudo tar -C /usr/local -xzf %v", filepath))
	if err != nil {
		return nil, fmt.Errorf("error unpacking tarball: %w", err)
	}

	return &InventoryRow{
		Path: hlp.Ptr("/usr/local/go"),
	}, nil
}

func (a *Agent) executeSeed_goInstall(ctx context.Context, seed *controllerv1.Seed) (*InventoryRow, error) {
	install := seed.Element.(*controllerv1.Seed_GoInstall).GoInstall

	gopath, err := exec.LookPath("go")
	if err != nil {
		return nil, fmt.Errorf("go not found in $PATH")
	}

	// Only check inventory if we're not installing "latest", otherwise just re-install it anyways
	version := "latest"
	if install.Version != nil {
		version = *install.Version
	}

	_, _, err = ExecuteOSCommand(gopath, "install", install.Package+"@"+version)
	if err != nil {
		return nil, fmt.Errorf("error installing package: %w", err)
	}

	return &InventoryRow{
		Package: hlp.Ptr(install.Package),
	}, nil
}

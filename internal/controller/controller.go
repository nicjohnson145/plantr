package controller

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"

	"sync"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt"
	"github.com/nicjohnson145/hlp/hashset"
	hsqlx "github.com/nicjohnson145/hlp/sqlx"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/nicjohnson145/plantr/internal/encryption"
	"github.com/nicjohnson145/plantr/internal/interceptors"
	"github.com/nicjohnson145/plantr/internal/parsingv2"
	"github.com/nicjohnson145/plantr/internal/token"
	"github.com/oklog/ulid/v2"
	"github.com/qdm12/reprint"
	"github.com/rs/zerolog"
)

var (
	ErrNoNodeIDError                = errors.New("node_id is required")
	ErrNoChallengeIDError           = errors.New("challenge_id required")
	ErrNoChallengeValueError        = errors.New("challenge_value required")
	ErrUnknownNodeIDError           = errors.New("unknown node_id")
	ErrUnknownChallengeIDError      = errors.New("unknown challenge_id")
	ErrIncorrectChallengeValueError = errors.New("incorrect challenge_value")
)

type ControllerConfig struct {
	Logger             zerolog.Logger
	GitClient          GitClient
	StorageClient      StorageClient
	RepoURL            string
	JWTSigningKey      []byte
	JWTDuration        time.Duration
	VaultClient        VaultClient
	HttpClient         *http.Client
	GithubReleaseToken string

	GithubWebhookSecret []byte

	NowFunc  func() time.Time             // for unit tests
	HashFunc func(*parsingv2.Seed) string // for unit tests
}

func NewController(conf ControllerConfig) (*Controller, error) {
	repoUrl := conf.RepoURL
	if !strings.HasSuffix(repoUrl, ".git") {
		repoUrl = repoUrl + ".git"
	}
	ctrl := &Controller{
		log:                 conf.Logger,
		git:                 conf.GitClient,
		store:               conf.StorageClient,
		repoUrl:             repoUrl,
		jwtSigningKey:       conf.JWTSigningKey,
		jwtDuration:         conf.JWTDuration,
		nowFunc:             conf.NowFunc,
		vault:               conf.VaultClient,
		httpClient:          conf.HttpClient,
		githubReleaseToken:  conf.GithubReleaseToken,
		configMu:            &sync.RWMutex{},
		vaultMu:             &sync.RWMutex{},
		hashFunc:            conf.HashFunc,
		githubWebhookSecret: conf.GithubWebhookSecret,
	}

	if ctrl.nowFunc == nil {
		ctrl.nowFunc = func() time.Time {
			return time.Now().UTC()
		}
	}
	if ctrl.hashFunc == nil {
		ctrl.hashFunc = seedHash
	}

	return ctrl, nil
}

type Controller struct {
	log                zerolog.Logger
	git                GitClient
	store              StorageClient
	repoUrl            string
	jwtSigningKey      []byte
	jwtDuration        time.Duration
	vault              VaultClient
	httpClient         *http.Client
	githubReleaseToken string

	githubWebhookSecret []byte

	configMu *sync.RWMutex
	config   *parsingv2.Config

	vaultMu   *sync.RWMutex
	vaultData *vaultData

	nowFunc  func() time.Time             // for unit tests
	hashFunc func(*parsingv2.Seed) string // for unit tests
}

func (c *Controller) now() time.Time {
	return c.nowFunc()
}

func (c *Controller) logAndHandleError(err error, msg string) error {
	str := "an error occurred"
	if msg != "" {
		str = msg
	}

	c.log.Err(err).Msg(str)

	switch true {
	case errors.Is(err, ErrNoNodeIDError):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, ErrNoChallengeIDError):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, ErrNoChallengeValueError):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, ErrUnknownNodeIDError):
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("permission denied"))
	case errors.Is(err, ErrUnknownChallengeIDError):
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("permission denied"))
	case errors.Is(err, ErrIncorrectChallengeValueError):
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("permission denied"))
	default:
		return err
	}
}

func (c *Controller) ensureConfig() error {
	if c.config != nil {
		c.log.Debug().Msg("config already present, nothing to do")
		return nil
	}

	c.log.Debug().Msg("no config loaded, loading now")

	c.log.Trace().Msgf("fetching latest commit for %v", c.repoUrl)
	latest, err := c.git.GetLatestCommit(c.repoUrl)
	if err != nil {
		return fmt.Errorf("error getting latest commit: %w", err)
	}

	c.log.Trace().Msg("cloning repo")
	repoFS, err := c.git.CloneAtCommit(c.repoUrl, latest)
	if err != nil {
		return fmt.Errorf("error cloning repo: %w", err)
	}

	c.log.Trace().Msg("parsing config from cloned repo")
	config, err := parsingv2.ParseFS(repoFS)
	if err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}

	c.configMu.Lock()
	c.config = config
	c.configMu.Unlock()

	return nil
}

func (c *Controller) cloneConfig() (*parsingv2.Config, error) {
	c.configMu.RLock()
	defer c.configMu.RUnlock()

	out := &parsingv2.Config{}
	if err := reprint.FromTo(c.config, out); err != nil {
		return nil, fmt.Errorf("error cloning config: %w", err)
	}
	return out, nil
}

func (c *Controller) ensureVault() error {
	c.vaultMu.Lock()
	defer c.vaultMu.Unlock()

	c.log.Trace().Msg("getting latest secret version")
	latest, err := c.vault.GetSecretVersion()
	if err != nil {
		return fmt.Errorf("error getting latest secret version: %w", err)
	}

	if c.vaultData != nil && c.vaultData.Version == latest {
		c.log.Debug().Msg("vautl data already at latest version, nothing to do")
		return nil
	}

	c.log.Trace().Msg("fetching secret data")
	data, err := c.vault.ReadSecretData()
	if err != nil {
		return fmt.Errorf("error reading secret data: %w", err)
	}

	if c.vaultData == nil {
		c.vaultData = &vaultData{}
	}
	c.vaultData.Version = latest
	c.vaultData.Data = data

	return nil
}

func (c *Controller) cloneVaultData() (map[string]any, error) {
	c.vaultMu.RLock()
	defer c.vaultMu.RUnlock()

	out := map[string]any{}
	if err := reprint.FromTo(c.vaultData.Data, &out); err != nil {
		return nil, fmt.Errorf("error cloning vault data: %w", err)
	}

	return out, nil
}

func (c *Controller) NewGithubWebhookHandler() (string, http.Handler) {
	return "/webhooks/github", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := c.handleGithubWebhook(r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}

type githubPushBody struct {
	Ref   string `json:"ref"`
	After string `json:"after"`
}

func (c *Controller) handleGithubWebhook(req *http.Request) error {
	body, err := c.validateGithubRequest(req)
	if err != nil {
		return c.logAndHandleError(err, "error validating webhook payload")
	}

	var pushBody githubPushBody
	if err := json.Unmarshal(body, &pushBody); err != nil {
		return c.logAndHandleError(err, "error unmarshalling body")
	}

	// TODO: configurable
	if pushBody.Ref != "refs/heads/main" {
		return nil
	}

	c.log.Info().Msg("recieved github webhook event for main, refreshing repo")
	repoFS, err := c.git.CloneAtCommit(c.repoUrl, pushBody.After)
	if err != nil {
		return c.logAndHandleError(err, "error cloning repo")
	}
	config, err := parsingv2.ParseFS(repoFS)
	if err != nil {
		return c.logAndHandleError(err, "error parsing config")
	}

	c.configMu.Lock()
	c.config = config
	c.configMu.Unlock()

	return nil
}

func (c *Controller) validateGithubRequest(req *http.Request) ([]byte, error) {
	if len(c.githubWebhookSecret) == 0 {
		return nil, fmt.Errorf("github webhook secret not set")
	}

	digest := req.Header.Get("X-Hub-Signature-256")
	if digest == "" {
		return nil, fmt.Errorf("missing X-Hub-Signature-256 header")
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %w", err)
	}

	mac := hmac.New(sha256.New, c.githubWebhookSecret)
	_, err = mac.Write(body)
	if err != nil {
		return nil, fmt.Errorf("error computing hash: %w", err)
	}
	computed := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(strings.TrimPrefix(digest, "sha256=")), []byte(computed)) {
		return nil, fmt.Errorf("computed digest does not match header")
	}

	return body, nil
}

func (c *Controller) Login(ctx context.Context, req *connect.Request[pbv1.LoginRequest]) (*connect.Response[pbv1.LoginResponse], error) {
	if err := c.validateLogin(req.Msg); err != nil {
		return nil, c.logAndHandleError(err, "error validating")
	}

	// Get the node that we're supposed to be logging in
	if err := c.ensureConfig(); err != nil {
		return nil, c.logAndHandleError(err, "error ensuring config")
	}
	conf, err := c.cloneConfig()
	if err != nil {
		return nil, c.logAndHandleError(err, "error cloning config")
	}

	var node *parsingv2.Node
	for _, n := range conf.Nodes {
		if n.ID == req.Msg.NodeId {
			node = n
		}
	}
	if node == nil {
		return nil, c.logAndHandleError(ErrUnknownNodeIDError, "unable to find node matching ID")
	}

	// If we've only got a node id, create a challenge and send it back
	if req.Msg.ChallengeId == nil {
		challengeID := ulid.Make().String()
		challengeValue := ulid.Make().String()

		encryptedValue, err := encryption.EncryptValue(challengeValue, node.PublicKey)
		if err != nil {
			return nil, c.logAndHandleError(err, "error encrypting challenge")
		}

		challenge := &Challenge{
			ID:    challengeID,
			Value: challengeValue,
		}
		if err := c.store.WriteChallenge(ctx, challenge); err != nil {
			return nil, c.logAndHandleError(err, "error inserting challenge")
		}

		return connect.NewResponse(&pbv1.LoginResponse{
			ChallengeId:     &challengeID,
			SealedChallenge: &encryptedValue,
		}), nil
	}

	// Otherwise, lets validate their challenge response and issue them a token
	storedChallenge, err := c.store.ReadChallenge(ctx, *req.Msg.ChallengeId)
	if err != nil {
		if errors.Is(err, hsqlx.ErrNotFoundError) {
			return nil, c.logAndHandleError(ErrUnknownChallengeIDError, "challenge id not found")
		}
		return nil, c.logAndHandleError(err, "error reading challenge")
	}
	if storedChallenge.Value != *req.Msg.ChallengeValue {
		return nil, c.logAndHandleError(ErrIncorrectChallengeValueError, "incorrect challenge value")
	}

	token, err := token.GenerateJWT(c.jwtSigningKey, token.Token{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: c.now().Add(c.jwtDuration).Unix(),
		},
		NodeID: req.Msg.NodeId,
	})
	if err != nil {
		return nil, c.logAndHandleError(err, "error generating JWT")
	}

	return connect.NewResponse(&pbv1.LoginResponse{
		Token: &token,
	}), nil
}

func (c *Controller) validateLogin(req *pbv1.LoginRequest) error {
	if req.NodeId == "" {
		return ErrNoNodeIDError
	}

	if req.ChallengeId != nil && req.ChallengeValue == nil {
		return ErrNoChallengeValueError
	}

	if req.ChallengeValue != nil && req.ChallengeId == nil {
		return ErrNoChallengeIDError
	}

	return nil
}

func (c *Controller) GetSyncData(ctx context.Context, req *connect.Request[pbv1.GetSyncDataRequest]) (*connect.Response[pbv1.GetSyncDataResponse], error) {
	token, err := interceptors.ClaimsFromCtx(ctx)
	if err != nil {
		return nil, c.logAndHandleError(err, "error getting token claims")
	}

	seeds, node, err := c.collectSeeds(token.NodeID)
	if err != nil {
		return nil, c.logAndHandleError(err, "error collecting seeds")
	}

	pbSeeds, err := c.renderSeeds(ctx, node, seeds)
	if err != nil {
		return nil, c.logAndHandleError(err, "error rendering seeds")
	}

	return connect.NewResponse(&pbv1.GetSyncDataResponse{
		Seeds: pbSeeds,
	}), nil
}

func (c *Controller) collectSeeds(nodeID string) ([]*parsingv2.Seed, *parsingv2.Node, error) {
	if err := c.ensureConfig(); err != nil {
		return nil, nil, fmt.Errorf("error ensuring config: %w", err)
	}
	conf, err := c.cloneConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("error cloning config: %w", err)
	}

	c.log.Trace().Msg("finding node from config")
	var node *parsingv2.Node
	for _, n := range conf.Nodes {
		if n.ID == nodeID {
			node = n
			break
		}
	}
	if node == nil {
		return nil, nil, ErrUnknownNodeIDError
	}

	c.log.Trace().Msg("collecting seeds from defined roles")
	seedSet := hashset.New(c.hashFunc)
	for _, roleName := range node.Roles {
		c.log.Trace().Msgf("collecting from role %v", roleName)
		seeds, ok := conf.Roles[roleName]
		if !ok {
			return nil, nil, fmt.Errorf("node %v references unknown role %v", nodeID, roleName)
		}

		for _, seed := range seeds {
			seed.Hash = c.hashFunc(seed)
			seedSet.Add(seed)
		}
	}

	return seedSet.AsSlice(), node, nil
}

func (c *Controller) renderSeeds(ctx context.Context, node *parsingv2.Node, seeds []*parsingv2.Seed) ([]*pbv1.Seed, error) {
	// Do this once per render instead of once per config file
	if err := c.ensureVault(); err != nil {
		return nil, fmt.Errorf("error ensuring vault data: %w", err)
	}
	vaultData, err := c.cloneVaultData()
	if err != nil {
		return nil, fmt.Errorf("error cloning vault data: %w", err)
	}

	outSeeds := make([]*pbv1.Seed, len(seeds))
	for i, seed := range seeds {
		switch concrete := seed.Element.(type) {
		case *parsingv2.ConfigFile:
			s, err := c.renderSeed_configFile(concrete, node, vaultData)
			if err != nil {
				return nil, fmt.Errorf("error converting config file: %w", err)
			}
			outSeeds[i] = s
		case *parsingv2.GithubRelease:
			s, err := c.renderSeed_githubRelease(ctx, concrete, node)
			if err != nil {
				return nil, fmt.Errorf("error converting github release: %w", err)
			}
			outSeeds[i] = s 
		case *parsingv2.SystemPackage:
			s, err := c.renderSeed_systemPackage(concrete, node)
			if err != nil {
				return nil, fmt.Errorf("error converting system package: %w", err)
			}
			outSeeds[i] = s
		case *parsingv2.GitRepo:
			s, err := c.renderSeed_gitRepo(concrete, node)
			if err != nil {
				return nil, fmt.Errorf("error converting git repo: %w", err)
			}
			outSeeds[i] = s
		case *parsingv2.Golang:
			outSeeds[i] = c.renderSeed_golang(concrete)
		case *parsingv2.GoInstall:
			outSeeds[i] = c.renderSeed_goInstall(concrete)
		case *parsingv2.UrlDownload:
			s, err := c.renderSeed_urlDownload(concrete, node)
			if err != nil {
				return nil, fmt.Errorf("error converting url_download: %w", err)
			}
			outSeeds[i] = s
		default:
			return nil, fmt.Errorf("unhandled seed type of %T", concrete)
		}

		// Set hash here so we dont have to pass it through all over
		if outSeeds[i].Metadata == nil {
			outSeeds[i].Metadata = &pbv1.Seed_Metadata{}
		}
		outSeeds[i].Metadata.Hash = seed.Hash
	}

	return outSeeds, nil
}

func (c *Controller) renderSeed_configFile(file *parsingv2.ConfigFile, node *parsingv2.Node, vaultData map[string]any) (*pbv1.Seed, error) {
	t, err := template.New("").Parse(file.TemplateContent)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	data := map[string]any{
		"Vault": vaultData,
	}

	buf := &bytes.Buffer{}

	if err := t.Execute(buf, data); err != nil {
		return nil, fmt.Errorf("error rendering template: %w", err)
	}

	dest := strings.ReplaceAll(file.Destination, "~", node.UserHome)
	return &pbv1.Seed{
		Metadata: &pbv1.Seed_Metadata{
			DisplayName: dest,
		},
		Element: &pbv1.Seed_ConfigFile{
			ConfigFile: &pbv1.ConfigFile{
				Content:     buf.String(),
				Destination: dest,
			},
		},
	}, nil
}

func (c *Controller) renderSeed_systemPackage(pkg *parsingv2.SystemPackage, node *parsingv2.Node) (*pbv1.Seed, error) {
	outPkg := &pbv1.SystemPackage{}

	var display string
	switch node.PackageManager {
	case "apt":
		if pkg.Apt == nil {
			return nil, fmt.Errorf("node has configured package manager 'apt', but no apt apt package configured")
		}
		outPkg.Pkg = &pbv1.SystemPackage_Apt{
			Apt: &pbv1.SystemPackage_AptPkg{
				Name: pkg.Apt.Name,
			},
		}
		display = pkg.Apt.Name
	default:
		return nil, fmt.Errorf("unhandled package manager %v", node.PackageManager)
	}

	return &pbv1.Seed{
		Metadata: &pbv1.Seed_Metadata{
			DisplayName: display,
		},
		Element: &pbv1.Seed_SystemPackage{
			SystemPackage: outPkg,
		},
	}, nil
}

func (c *Controller) renderSeed_gitRepo(repo *parsingv2.GitRepo, node *parsingv2.Node) (*pbv1.Seed, error) {
	outRepo := &pbv1.GitRepo{
		Url:      repo.URL,
		Location: strings.ReplaceAll(repo.Location, "~", node.UserHome),
	}

	switch true {
	case repo.Tag != nil:
		outRepo.Ref = &pbv1.GitRepo_Tag{Tag: *repo.Tag}
	case repo.Commit != nil:
		outRepo.Ref = &pbv1.GitRepo_Commit{Commit: *repo.Commit}
	default:
		return nil, fmt.Errorf("unable to determine output ref")
	}

	return &pbv1.Seed{
		Metadata: &pbv1.Seed_Metadata{
			DisplayName: repo.URL,
		},
		Element: &pbv1.Seed_GitRepo{
			GitRepo: outRepo,
		},
	}, nil
}

func (c *Controller) renderSeed_golang(golang *parsingv2.Golang) *pbv1.Seed {
	return &pbv1.Seed{
		Metadata: &pbv1.Seed_Metadata{
			DisplayName: fmt.Sprintf("go@%v", golang.Version),
		},
		Element: &pbv1.Seed_Golang{
			Golang: &pbv1.Golang{
				Version: golang.Version,
			},
		},
	}
}

func (c *Controller) renderSeed_goInstall(goinstall *parsingv2.GoInstall) *pbv1.Seed {
	return &pbv1.Seed{
		Metadata: &pbv1.Seed_Metadata{
			DisplayName: goinstall.Package,
		},
		Element: &pbv1.Seed_GoInstall{
			GoInstall: &pbv1.GoInstall{
				Package: goinstall.Package,
				Version: goinstall.Version,
			},
		},
	}
}

func (c *Controller) renderSeed_urlDownload(urlDownload *parsingv2.UrlDownload, node *parsingv2.Node) (*pbv1.Seed, error) {
	missingErr := errors.New(fmt.Sprintf("no url configured for %v/%v", node.OS, node.Arch))

	archMap, ok := urlDownload.Urls[node.OS]
	if !ok {
		return nil, missingErr
	}

	url, ok := archMap[node.Arch]
	if !ok {
		return nil, missingErr
	}

	return &pbv1.Seed{
		Metadata: &pbv1.Seed_Metadata{
			DisplayName: url,
		},
		Element: &pbv1.Seed_UrlDownload{
			UrlDownload: &pbv1.UrlDownload{
				NameOverride:         urlDownload.NameOverride,
				DownloadUrl:          url,
				DestinationDirectory: node.BinDir,
			},
		},
	}, nil
}

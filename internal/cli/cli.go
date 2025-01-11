package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"buf.build/go/protoyaml"
	"github.com/manifoldco/promptui"
	configv1 "github.com/nicjohnson145/plantr/gen/plantr/config/v1"
	"github.com/nicjohnson145/plantr/internal/encryption"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
)

type CLIConfig struct {
	Logger zerolog.Logger
}

func NewCLI(conf CLIConfig) *CLI {
	return &CLI{
		log: conf.Logger,
	}
}

type CLI struct {
	log zerolog.Logger
}

func (c *CLI) logErr(err error, ctx string) error {
	c.log.Err(err).Msg(ctx)
	return fmt.Errorf("%v: %w", ctx, err)
}

func (c *CLI) GenerateKeyPair() error {
	public, private, err := encryption.GenerateKeyPair(&encryption.KeyOpts{})
	if err != nil {
		c.log.Err(err).Msg("error generating keypair")
		return err
	}

	if err := os.WriteFile("key", []byte(private), 0664); err != nil { //nolint: gosec
		c.log.Err(err).Msg("error writing private key file")
		return err
	}

	if err := os.WriteFile("key.pub", []byte(public), 0664); err != nil { //nolint: gosec
		c.log.Err(err).Msg("error writing public key file")
		return err
	}

	return nil
}

type InitOpts struct {
	ControllerAddress string
	ID                string
	PublicKeyPath     string
	UserHome          string
	PackageManager    string
}

func (c *CLI) Init(opts InitOpts) error {
	hostname, err := os.Hostname()
	if err != nil {
		return c.logErr(err, "error getting hostname")
	}

	node := &configv1.Node{
		Os:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// Set node id
	if opts.ID == "" {
		p := promptui.Prompt{
			Label:     "Node ID (empty will generate a ulid)",
			Default:   hostname,
			AllowEdit: true,
		}
		res, err := p.Run()
		if err != nil {
			return c.logErr(err, "error getting node id")
		}

		var id string
		if res == "" {
			id = ulid.Make().String()
		} else {
			id = res
		}

		node.Id = id
	} else {
		node.Id = opts.ID
	}

	if opts.UserHome == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return c.logErr(err, "error getting user home dir")
		}

		p := promptui.Prompt{
			Label:     "User Home Directory",
			Default:   userHome,
			AllowEdit: true,
		}

		res, err := p.Run()
		if err != nil {
			return c.logErr(err, "error getting user home directory")
		}
		node.UserHome = res
	} else {
		node.UserHome = opts.UserHome
	}

	if opts.PackageManager == "" {
		mgr, err := c.tryDetectPackageManager()
		if err != nil {
			return c.logErr(err, "error attempting to auto detect package manager")
		}

		p := promptui.Prompt{
			Label:     "Package Manager",
			Default:   mgr,
			AllowEdit: true,
		}

		res, err := p.Run()
		if err != nil {
			return c.logErr(err, "error getting package manager")
		}

		node.PackageManager = res
	} else {
		node.PackageManager = opts.PackageManager
	}

	marshaller := protoyaml.MarshalOptions{
		UseProtoNames: true,
	}

	nodeYaml, err := marshaller.Marshal(node)
	if err != nil {
		return c.logErr(err, "error marshalling node to yaml")
	}

	fmt.Println("plantr.yaml node definition")
	fmt.Println(string(nodeYaml))

	return nil
}

func (c *CLI) tryDetectPackageManager() (string, error) {
	if _, err := exec.LookPath("apt"); err != nil {
		if !errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("error checking $PATH for `apt`: %w", err)
		}
	} else {
		return "apt", nil
	}

	return "", nil
}

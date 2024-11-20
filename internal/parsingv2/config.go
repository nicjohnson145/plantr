package parsingv2

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"buf.build/go/protoyaml"
	"github.com/nicjohnson145/hlp"
	configv1 "github.com/nicjohnson145/plantr/gen/plantr/config/v1"
)

var (
	ErrParseError = errors.New("parse error")

	ErrConfigFileNoPathError        = errors.New("path is required")
	ErrConfigFileNoDestinationError = errors.New("destination is required")

	ErrNodeNoIDError            = errors.New("id is required")
	ErrNodeNoPulblicKeyError    = errors.New("public_key_b64 is required")
	ErrNodeNoUserHomeError      = errors.New("user_home is required")
	ErrNodePublicKeyDecodeError = errors.New("error decoding public key")
)

func ParseFS(fsys fs.FS) (*Config, error) {
	conf, err := parseFS(fsys)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseError, err)
	}
	return conf, nil
}

func parseFS(fsys fs.FS) (*Config, error) {
	root, err := fs.ReadFile(fsys, "plantr.yaml")
	if err != nil {
		return nil, fmt.Errorf("error reading platnr.yaml: %w", err)
	}

	pbConfig := &configv1.Config{}
	if err := protoyaml.Unmarshal(root, pbConfig); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	config := &Config{
		Roles: make(map[string][]*Seed),
	}
	for i, node := range pbConfig.Nodes {
		outNode, err := parseNode(node)
		if err != nil {
			return nil, fmt.Errorf("error parsing node %v: %w", i, err)
		}
		config.Nodes = append(config.Nodes, outNode)
	}

	for rolename, role := range pbConfig.Roles {
		seeds, err := parseRole(fsys, role.Seeds)
		if err != nil {
			return nil, fmt.Errorf("error parsing role %v: %w", rolename, err)
		}
		config.Roles[rolename] = seeds
	}

	return config, nil
}

func parseNode(node *configv1.Node) (*Node, error) {
	if node.Id == "" {
		return nil, ErrNodeNoIDError
	}
	if node.PublicKeyB64 == "" {
		return nil, ErrNodeNoPulblicKeyError
	}
	if node.UserHome == "" {
		return nil, ErrNodeNoUserHomeError
	}

	pubKeyBytes, err := base64.StdEncoding.DecodeString(node.PublicKeyB64)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNodePublicKeyDecodeError, err)
	}

	return &Node{
		ID:        node.Id,
		Hostname:  node.Hostname,
		PublicKey: string(pubKeyBytes),
		Roles:     node.Roles,
		UserHome:  node.UserHome,
		BinDir:    hlp.Ternary(node.BinDir == "", filepath.Join(node.UserHome, "bin"), node.BinDir),
	}, nil
}

func parseRole(fsys fs.FS, seeds []*configv1.Seed) ([]*Seed, error) {
	outSeeds := make([]*Seed, len(seeds))
	for i, s := range seeds {
		switch concrete := s.Element.(type) {
		case *configv1.Seed_ConfigFile:
			seed, err := parseSeed_configFile(fsys, concrete.ConfigFile)
			if err != nil {
				return nil, fmt.Errorf("error parsing item %v: %w", i, err)
			}
			outSeeds[i] = seed
		}
	}

	return outSeeds, nil
}

func parseSeed_configFile(fsys fs.FS, file *configv1.ConfigFile) (*Seed, error) {
	if file.Path == "" {
		return nil, ErrConfigFileNoPathError
	}
	if file.Destination == "" {
		return nil, ErrConfigFileNoDestinationError
	}

	tmplBytes, err := fs.ReadFile(fsys, file.Path)
	if err != nil {
		return nil, fmt.Errorf("error reading template content: %w", err)
	}

	return &Seed{
		Element: &ConfigFile{
			TemplateContent: string(tmplBytes),
			Destination:     file.Destination,
		},
	}, nil
}

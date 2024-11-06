package parsing

import (
	"buf.build/go/protoyaml"
	"errors"
	"fmt"
	"io/fs"
	"encoding/base64"

	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/v1"
)

var (
	ErrParseError          = errors.New("parse error")
	ErrNoTemplatePathError = errors.New("template_path is required")
	ErrNoDestinationError  = errors.New("destination is required")
)

func ParseRepoFS(repoFS fs.FS) (*pbv1.Config, error) {
	root, err := fs.ReadFile(repoFS, "plantr.yaml")
	if err != nil {
		return nil, fmt.Errorf("error reading root config: %w", err)
	}

	conf := &pbv1.Config{}
	if err := protoyaml.Unmarshal(root, conf); err != nil {
		return nil, fmt.Errorf("error unmarshalling: %w", err)
	}

	for roleName, role := range conf.Roles {
		for i, seed := range role.Seeds {
			if err := parseSeed(repoFS, seed); err != nil {
				return nil, fmt.Errorf("error parsing role/%v/seed/%v: %w", roleName, i, err)
			}
		}
	}

	for _, node := range conf.Nodes {
		parseNode(node)
	}

	return conf, nil
}

func parseNode(node *pbv1.Node) error {
	// Decode the public key from b64
	outBytes, err := base64.StdEncoding.DecodeString(node.PublicKey)
	if err != nil {
		return fmt.Errorf("error base64 decoding public key: %w", err)
	}
	node.PublicKey = string(outBytes)

	return nil
}

func parseSeed(repoFS fs.FS, seed *pbv1.Seed) error {
	parse := func() error {
		switch concrete := seed.Element.(type) {
		case *pbv1.Seed_ConfigFile:
			if err := parseSeed_configFile(repoFS, concrete.ConfigFile); err != nil {
				return fmt.Errorf("error parsing ConfigFile: %w", err)
			}
			return nil
		default:
			return nil
		}
	}

	if err := parse(); err != nil {
		return fmt.Errorf("%w: %w", ErrParseError, err)
	}
	return nil
}

func parseSeed_configFile(repoFS fs.FS, configFile *pbv1.ConfigFile) error {
	if configFile.TemplatePath == "" {
		return ErrNoTemplatePathError
	}
	if configFile.Destination == "" {
		return ErrNoDestinationError
	}

	content, err := fs.ReadFile(repoFS, configFile.TemplatePath)
	if err != nil {
		return fmt.Errorf("error reading template path: %w", err)
	}

	configFile.TemplateContent = string(content)

	return nil
}

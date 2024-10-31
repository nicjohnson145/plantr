package parsing

import (
	"buf.build/go/protoyaml"
	"errors"
	"fmt"
	"io/fs"

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

	return conf, nil
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

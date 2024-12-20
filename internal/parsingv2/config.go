package parsingv2

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"

	"buf.build/go/protoyaml"
	"github.com/bufbuild/protovalidate-go"
	"github.com/nicjohnson145/hlp"
	configv1 "github.com/nicjohnson145/plantr/gen/plantr/config/v1"
)

var (
	ErrParseError                     = errors.New("parse error")
	ErrNodePublicKeyDecodeError       = errors.New("error decoding public key")
	ErrGithubReleaseInvalidRegexError = errors.New("invalid regex")
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
	if err := protovalidate.Validate(node); err != nil {
		return nil, fmt.Errorf("error validating: %w", err)
	}

	pubKeyBytes, err := base64.StdEncoding.DecodeString(node.PublicKeyB64)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNodePublicKeyDecodeError, err)
	}

	return &Node{
		ID:             node.Id,
		Hostname:       node.Hostname,
		PublicKey:      string(pubKeyBytes),
		Roles:          node.Roles,
		UserHome:       node.UserHome,
		BinDir:         hlp.Ternary(node.BinDir == "", filepath.Join(node.UserHome, "bin"), node.BinDir),
		OS:             node.Os,
		Arch:           node.Arch,
		PackageManager: node.PackageManager,
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
		case *configv1.Seed_GithubRelease:
			seed, err := parseSeed_githubRelease(concrete.GithubRelease)
			if err != nil {
				return nil, fmt.Errorf("error parsing item %v: %w", i, err)
			}
			outSeeds[i] = seed
		case *configv1.Seed_SystemPackage:
			seed, err := parseSeed_systemPackage(concrete.SystemPackage)
			if err != nil {
				return nil, fmt.Errorf("error parsing item %v: %w", i, err)
			}
			outSeeds[i] = seed
		default:
			return nil, fmt.Errorf("unhandled seed type %T", concrete)
		}
	}

	return outSeeds, nil
}

func parseSeed_configFile(fsys fs.FS, file *configv1.ConfigFile) (*Seed, error) {
	if err := protovalidate.Validate(file); err != nil {
		return nil, fmt.Errorf("error validating: %w", err)
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

func parseSeed_githubRelease(release *configv1.GithubRelease) (*Seed, error) {
	if err := protovalidate.Validate(release); err != nil {
		return nil, fmt.Errorf("error validating: %w", err)
	}

	assetPatterns := map[string]map[string]*regexp.Regexp{}
	if release.AssetPatterns != nil {
		if release.AssetPatterns.Linux != nil {
			assetPatterns["linux"] = map[string]*regexp.Regexp{}
			if release.AssetPatterns.Linux.Amd64 != "" {
				pat, err := regexp.Compile(release.AssetPatterns.Linux.Amd64)
				if err != nil {
					return nil, fmt.Errorf("%w: error parsing regex for %v/%v: %w", ErrGithubReleaseInvalidRegexError, "linux", "amd64", err)
				}
				assetPatterns["linux"]["amd64"] = pat
			}
			if release.AssetPatterns.Linux.Arm64 != "" {
				pat, err := regexp.Compile(release.AssetPatterns.Linux.Arm64)
				if err != nil {
					return nil, fmt.Errorf("%w: error parsing regex for %v/%v: %w", ErrGithubReleaseInvalidRegexError, "linux", "arm64", err)
				}
				assetPatterns["linux"]["arm64"] = pat
			}
		}
		if release.AssetPatterns.Mac != nil {
			assetPatterns["mac"] = map[string]*regexp.Regexp{}
			if release.AssetPatterns.Mac.Amd64 != "" {
				pat, err := regexp.Compile(release.AssetPatterns.Mac.Amd64)
				if err != nil {
					return nil, fmt.Errorf("%w: error parsing regex for %v/%v: %w", ErrGithubReleaseInvalidRegexError, "mac", "amd64", err)
				}
				assetPatterns["mac"]["amd64"] = pat
			}
			if release.AssetPatterns.Mac.Arm64 != "" {
				pat, err := regexp.Compile(release.AssetPatterns.Mac.Arm64)
				if err != nil {
					return nil, fmt.Errorf("%w: error parsing regex for %v/%v: %w", ErrGithubReleaseInvalidRegexError, "mac", "arm64", err)
				}
				assetPatterns["mac"]["arm64"] = pat
			}
		}
	}

	return &Seed{
		Element: &GithubRelease{
			Repo:               release.Repo,
			AssetPatterns:      assetPatterns,
			Tag:                release.Tag,
			BinaryNameOverride: release.BinaryNameOverride,
		},
	}, nil
}

func parseSeed_systemPackage(pkg *configv1.SystemPackage) (*Seed, error) {
	if err := protovalidate.Validate(pkg); err != nil {
		return nil, fmt.Errorf("error validating: %w", err)
	}

	outPkg := &SystemPackage{}

	if pkg.Apt != nil {
		outPkg.Apt = &SystemPackageApt{
			Name: pkg.Apt.Name,
		}
	}

	return &Seed{
		Element: outPkg,
	}, nil
}

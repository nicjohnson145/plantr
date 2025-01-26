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
		seeds, err := parseRole(pbConfig, fsys, role.Seeds)
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

func parseRole(rootConfig *configv1.Config, fsys fs.FS, seeds []*configv1.Seed) ([]*Seed, error) {
	outSeeds := []*Seed{}
	for i, s := range seeds {
		var seed *Seed
		var err error
		var seeds []*Seed
		switch concrete := s.Element.(type) {
		case *configv1.Seed_ConfigFile:
			seed, err = parseSeed_configFile(fsys, concrete.ConfigFile)
		case *configv1.Seed_GithubRelease:
			seed, err = parseSeed_githubRelease(concrete.GithubRelease)
		case *configv1.Seed_SystemPackage:
			seed, err = parseSeed_systemPackage(concrete.SystemPackage)
		case *configv1.Seed_GitRepo:
			seed, err = parseSeed_gitRepo(concrete.GitRepo)
		case *configv1.Seed_Golang:
			seed, err = parseSeed_golang(concrete.Golang)
		case *configv1.Seed_GoInstall:
			seed, err = parseSeed_goInstall(concrete.GoInstall)
		case *configv1.Seed_UrlDownload:
			seed, err = parseSeed_urlDownload(concrete.UrlDownload)
		case *configv1.Seed_RoleGroup:
			seeds, err = parseSeed_roleGroup(rootConfig, fsys, concrete.RoleGroup)
		default:
			return nil, fmt.Errorf("unhandled seed type %T", concrete)
		}
		if err != nil {
			return nil, fmt.Errorf("error parsing item %v: %w", i, err)
		}

		if seed != nil {
			var meta *SeedMetadata
			if s.Meta != nil {
				meta = &SeedMetadata{
					Name: s.Meta.Name,
				}
			}
			seed.Metadata = meta

			outSeeds = append(outSeeds, seed)
		}
		if seeds != nil {
			outSeeds = append(outSeeds, seeds...)
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

	if release.BinaryRegex != nil {
		_, err := regexp.Compile(*release.BinaryRegex)
		if err != nil {
			return nil, fmt.Errorf("%w: error parsing binary regex: %w", ErrGithubReleaseInvalidRegexError, err)
		}
	}

	return &Seed{
		Element: &GithubRelease{
			Repo:           release.Repo,
			AssetPatterns:  assetPatterns,
			Tag:            release.Tag,
			NameOverride:   release.NameOverride,
			ArchiveRelease: release.ArchiveRelease,
			BinaryRegex:    release.BinaryRegex,
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
	if pkg.Brew != nil {
		outPkg.Brew = &SystemPackageBrew{
			Name: pkg.Brew.Name,
		}
	}

	return &Seed{
		Element: outPkg,
	}, nil
}

func parseSeed_gitRepo(repo *configv1.GitRepo) (*Seed, error) {
	if err := protovalidate.Validate(repo); err != nil {
		return nil, fmt.Errorf("error validating: %w", err)
	}

	outRepo := &GitRepo{
		URL:      repo.Url,
		Location: repo.Location,
	}
	switch concrete := repo.Ref.(type) {
	case *configv1.GitRepo_Tag:
		outRepo.Tag = hlp.Ptr(concrete.Tag)
	case *configv1.GitRepo_Commit:
		outRepo.Commit = hlp.Ptr(concrete.Commit)
	default:
		return nil, fmt.Errorf("unhandled ref type of %T", concrete)
	}

	return &Seed{
		Element: outRepo,
	}, nil
}

func parseSeed_golang(golang *configv1.Golang) (*Seed, error) {
	if err := protovalidate.Validate(golang); err != nil {
		return nil, fmt.Errorf("error validating: %w", err)
	}

	return &Seed{
		Element: &Golang{
			Version: golang.Version,
		},
	}, nil
}

func parseSeed_goInstall(goinstall *configv1.GoInstall) (*Seed, error) {
	if err := protovalidate.Validate(goinstall); err != nil {
		return nil, fmt.Errorf("error validating: %w", err)
	}

	return &Seed{
		Element: &GoInstall{
			Package: goinstall.Package,
			Version: goinstall.Version,
		},
	}, nil
}

func parseSeed_urlDownload(urlDownload *configv1.UrlDownload) (*Seed, error) {
	if err := protovalidate.Validate(urlDownload); err != nil {
		return nil, fmt.Errorf("error validating: %w", err)
	}

	element := &UrlDownload{
		NameOverride: urlDownload.NameOverride,
		Urls:         map[string]map[string]string{},
	}

	setArchUrls := func(archGroup *configv1.UrlDownload_OsGroup_ArchGroup) map[string]string {
		out := map[string]string{}
		if archGroup.Amd64 != nil {
			out["amd64"] = *archGroup.Amd64
		}
		if archGroup.Arm64 != nil {
			out["arm64"] = *archGroup.Arm64
		}

		return out
	}

	if urlDownload.Urls.Linux != nil {
		element.Urls["linux"] = setArchUrls(urlDownload.Urls.Linux)
	}
	if urlDownload.Urls.Mac != nil {
		element.Urls["darwin"] = setArchUrls(urlDownload.Urls.Mac)
	}

	urlCount := 0
	for _, archMap := range element.Urls {
		urlCount += len(archMap)
	}
	if urlCount == 0 {
		return nil, fmt.Errorf("must specify at least one OS/Arch url")
	}

	return &Seed{
		Element: element,
	}, nil
}

func parseSeed_roleGroup(rootConfig *configv1.Config, fsys fs.FS, roleGroup *configv1.RoleGroup) ([]*Seed, error) {
	if err := protovalidate.Validate(roleGroup); err != nil {
		return nil, fmt.Errorf("error validating: %w", err)
	}

	seeds := []*Seed{}

	for _, roleName := range roleGroup.Roles {
		roleObj, ok := rootConfig.Roles[roleName]
		if !ok {
			return nil, fmt.Errorf("referenced role %v not found", roleName)
		}

		roleSeeds, err := parseRole(rootConfig, fsys, roleObj.Seeds)
		if err != nil {
			return nil, fmt.Errorf("error parsing sub role %v: %w", roleName, err)
		}

		seeds = append(seeds, roleSeeds...)
	}

	return seeds, nil
}

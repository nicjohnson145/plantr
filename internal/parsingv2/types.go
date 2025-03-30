package parsingv2

import (
	"crypto/md5" //nolint:gosec // its for fingerprinting, it doesnt have to be cryptographically secure
	"fmt"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"regexp"
	"strings"
)

func hash(parts []string) string {
	return fmt.Sprint(md5.Sum([]byte(strings.Join(parts, "")))) //nolint: gosec // its a hash, it doesnt have to be cryptographically secure
}

type ISeed interface {
	DisplayName(n *Node) (string, error)
	ComputeHash(n *Node) (string, error)
}

type Node struct {
	ID             string
	Hostname       string
	PublicKey      string
	Roles          []string
	UserHome       string
	BinDir         string
	OS             string
	Arch           string
	PackageManager string
}

type SeedMetadata struct {
	Name *string
}

var _ ISeed = (*Seed)(nil)

type Seed struct {
	Metadata *SeedMetadata
	Element  ISeed
}

func (s *Seed) DisplayName(n *Node) (string, error) {
	if s.Metadata != nil && s.Metadata.Name != nil {
		return *s.Metadata.Name, nil
	}
	return s.Element.DisplayName(n)
}

func (s *Seed) ComputeHash(n *Node) (string, error) {
	return s.Element.ComputeHash(n)
}

type Config struct {
	Roles map[string][]*Seed
	Nodes []*Node
}

var _ ISeed = (*ConfigFile)(nil)

type ConfigFile struct {
	TemplateContent string
	Destination     string
	Mode            string
}

func (c *ConfigFile) DisplayName(_ *Node) (string, error) {
	return c.Destination, nil
}

func (c *ConfigFile) ComputeHash(_ *Node) (string, error) {
	return hash([]string{
		"ConfigFile",
		c.TemplateContent,
		c.Destination,
		c.Mode,
	}), nil
}

var _ ISeed = (*GithubRelease)(nil)

type GithubRelease struct {
	Repo           string
	AssetPatterns  map[string]map[string]*regexp.Regexp
	Tag            string
	NameOverride   *string
	ArchiveRelease bool
	BinaryRegex    *string
}

func (g *GithubRelease) DisplayName(_ *Node) (string, error) {
	return g.Repo + "@" + g.Tag, nil
}

func (g *GithubRelease) ComputeHash(_ *Node) (string, error) {
	return hash([]string{
		"GithubRelease",
		g.Repo,
		g.Tag,
	}), nil
}

func (g *GithubRelease) GetAssetPattern(os string, arch string) *regexp.Regexp {
	archMap, ok := g.AssetPatterns[os]
	if !ok {
		return nil
	}
	return archMap[arch]
}

type SystemPackageApt struct {
	Name string
}

type SystemPackageBrew struct {
	Name string
}

var _ ISeed = (*SystemPackage)(nil)

type SystemPackage struct {
	Apt  *SystemPackageApt
	Brew *SystemPackageBrew
}

func (s *SystemPackage) DisplayName(node *Node) (string, error) {
	name, err := s.GetPackageName(node)
	if err != nil {
		return "", err
	}
	return "PKG:" + name, nil
}

func (s *SystemPackage) ComputeHash(node *Node) (string, error) {
	name, err := s.GetPackageName(node)
	if err != nil {
		return "", err
	}

	return hash([]string{
		"SystemPackage",
		name,
	}), nil
}

func (s *SystemPackage) GetPackageName(node *Node) (string, error) {
	name, _, err := s.getNameObject(node)
	if err != nil {
		return "", err
	}
	return name, nil
}

func (s *SystemPackage) GetPackageObject(node *Node) (*pbv1.Seed_SystemPackage, error) {
	_, obj, err := s.getNameObject(node)
	if err != nil {
		return nil, err
	}
	return &pbv1.Seed_SystemPackage{
		SystemPackage: obj,
	}, nil
}

func (s *SystemPackage) getNameObject(node *Node) (string, *pbv1.SystemPackage, error) {
	switch node.PackageManager {
	case "apt":
		if s.Apt == nil {
			return "", nil, fmt.Errorf("node has apt package manager but no apt package configured")
		}
		return s.Apt.Name, &pbv1.SystemPackage{Pkg: &pbv1.SystemPackage_Apt{Apt: &pbv1.SystemPackage_AptPkg{Name: s.Apt.Name}}}, nil
	case "brew":
		if s.Brew == nil {
			return "", nil, fmt.Errorf("node has brew package manager but no brew package configured")
		}
		return s.Brew.Name, &pbv1.SystemPackage{Pkg: &pbv1.SystemPackage_Brew{Brew: &pbv1.SystemPackage_BrewPkg{Name: s.Brew.Name}}}, nil
	default:
		return "", nil, fmt.Errorf("unhandled package manager %v", node.PackageManager)
	}
}

var _ ISeed = (*GitRepo)(nil)

type GitRepo struct {
	URL      string
	Location string
	Tag      *string
	Commit   *string
}

func (g *GitRepo) DisplayName(_ *Node) (string, error) {
	return g.URL + "@" + g.getRef(), nil
}

func (g *GitRepo) ComputeHash(_ *Node) (string, error) {
	return hash([]string{
		"GitRepo",
		g.URL,
		g.getRef(),
		g.Location,
	}), nil
}

func (g *GitRepo) getRef() string {
	if g.Tag != nil {
		return *g.Tag
	} else {
		return *g.Commit
	}
}

var _ ISeed = (*Golang)(nil)

type Golang struct {
	Version string
}

func (g *Golang) DisplayName(_ *Node) (string, error) {
	return "go@" + g.Version, nil
}

func (g *Golang) ComputeHash(_ *Node) (string, error) {
	return hash([]string{
		"Golang",
		g.Version,
	}), nil
}

var _ ISeed = (*GoInstall)(nil)

type GoInstall struct {
	Package string
	Version *string
}

func (g *GoInstall) DisplayName(_ *Node) (string, error) {
	return g.Package + "@" + g.getVersion(), nil
}

func (g *GoInstall) ComputeHash(_ *Node) (string, error) {
	return hash([]string{
		"GoInstall",
		g.Package,
		g.getVersion(),
	}), nil
}

func (g *GoInstall) getVersion() string {
	if g.Version != nil {
		return *g.Version
	}
	return "latest"
}

var _ ISeed = (*UrlDownload)(nil)

type UrlDownload struct {
	NameOverride   *string
	Urls           map[string]map[string]string
	ArchiveRelease bool
}

func (u *UrlDownload) GetUrl(node *Node) (string, error) {
	missingErr := fmt.Errorf("no url configured for %v/%v", node.OS, node.Arch)

	archMap, ok := u.Urls[node.OS]
	if !ok {
		return "", missingErr
	}

	url, ok := archMap[node.Arch]
	if !ok {
		return "", missingErr
	}

	return url, nil
}

func (u *UrlDownload) DisplayName(node *Node) (string, error) {
	url, err := u.GetUrl(node)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (u *UrlDownload) ComputeHash(node *Node) (string, error) {
	url, err := u.GetUrl(node)
	if err != nil {
		return "", err
	}

	return hash([]string{
		"UrlDownload",
		url,
	}), nil
}

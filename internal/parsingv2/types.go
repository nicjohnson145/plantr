package parsingv2

import "regexp"

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

type Seed struct {
	Element any
}

type Config struct {
	Roles map[string][]*Seed
	Nodes []*Node
}

type ConfigFile struct {
	TemplateContent string
	Destination     string
}

type GithubRelease struct {
	Repo               string
	AssetPatterns      map[string]map[string]*regexp.Regexp
	Tag                string
	BinaryNameOverride *string
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

type SystemPackage struct {
	Apt *SystemPackageApt
}

type GitRepo struct {
	URL      string
	Location string
	Tag      *string
	Commit   *string
}

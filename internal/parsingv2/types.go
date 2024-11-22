package parsingv2

type Node struct {
	ID        string
	Hostname  string
	PublicKey string
	Roles     []string
	UserHome  string
	BinDir    string
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
	Repo          string
	AssetPatterns map[string]map[string]string
}

func (g *GithubRelease) GetAssetPattern(os string, arch string) string {
	archMap, ok := g.AssetPatterns[os]
	if !ok {
		return ""
	}
	return archMap[arch]
}

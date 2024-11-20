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

type ConfigFile struct {
	TemplateContent string
	Destination     string
}

type Config struct {
	Roles map[string][]*Seed
	Nodes []*Node
}

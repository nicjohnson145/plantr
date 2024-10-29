package parsing

import (
	"fmt"
	"io/fs"
	"buf.build/go/protoyaml"

	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/v1"
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

	return conf, nil
}

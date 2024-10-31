package parsing

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
)

func protoMustEqual(t *testing.T, want any, got any, opts ...cmp.Option) {
	defaultOpts := []cmp.Option{protocmp.Transform()}
	defaultOpts = append(defaultOpts, opts...)

	if diff := cmp.Diff(want, got, defaultOpts...); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestParseRepoFS(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		got, err := ParseRepoFS(os.DirFS("./testdata/basic-repo"))
		require.NoError(t, err)

		want := &pbv1.Config{
			Roles: map[string]*pbv1.Role{
				"first_role": {
					Seeds: []*pbv1.Seed{
						{
							Element: &pbv1.Seed_GithubReleaseBinary{
								GithubReleaseBinary: &pbv1.GithubReleaseBinary{
									RepoUrl: "https://github.com/foo/bar",
								},
							},
						},
						{
							Element: &pbv1.Seed_ConfigFile{
								ConfigFile: &pbv1.ConfigFile{
									TemplatePath: "templates/some-config",
									Destination: "~/.some-config",
									TemplateContent: "im a config file on {{ .Host }}\n",
								},
							},
						},
					},
				},
			},
			Nodes: []*pbv1.Node{
				{
					Id:        "node-one-id",
					Hostname:  "node_one",
					PublicKey: "node-one-public-key",
					Roles: []string{
						"first_role",
					},
				},
			},
		}
		protoMustEqual(t, want, got)
	})
}

package controller

import (
	"context"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/nicjohnson145/hlp"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/nicjohnson145/plantr/internal/parsingv2"
	"github.com/stretchr/testify/require"
)

func TestRenderSeedConfigFile(t *testing.T) {
	t.Run("has seed", func(t *testing.T) {
		ctrl, err := NewController(ControllerConfig{
			VaultClient: &NoopVault{},
		})
		require.NoError(t, err)

		node := &parsingv2.Node{
			UserHome: "/tmp/someuser",
		}
		seeds := []*parsingv2.Seed{
			{
				Metadata: &parsingv2.SeedMetadata{
					Name: hlp.Ptr("seed-one"),
				},
				Element: &parsingv2.ConfigFile{
					TemplateContent: "seed-one-content",
					Destination: "~/seed-one",
				},
			},
			{
				Element: &parsingv2.ConfigFile{
					TemplateContent: dedent.Dedent(`
						{{- if HasSeed "seed-one" }}
						seed-one-installed
						{{- else}}
						seed-one-not-installed
						{{- end}}
					`),
					Destination: "~/seed-two",
				},
			},
		}

		pbSeeds, err := ctrl.renderSeeds(context.Background(), node, seeds)
		require.NoError(t, err)

		wantPb := []*pbv1.Seed{
			{
				Metadata: &pbv1.Seed_Metadata{
					DisplayName: "/tmp/someuser/seed-one",
				},
				Element: &pbv1.Seed_ConfigFile{
					ConfigFile: &pbv1.ConfigFile{
						Content: "seed-one-content",
						Destination: "/tmp/someuser/seed-one",
					},
				},
			},
			{
				Metadata: &pbv1.Seed_Metadata{
					DisplayName: "/tmp/someuser/seed-two",
				},
				Element: &pbv1.Seed_ConfigFile{
					ConfigFile: &pbv1.ConfigFile{
						Content: "\nseed-one-installed\n",
						Destination: "/tmp/someuser/seed-two",
					},
				},
			},
		}
		pbEqual(t, wantPb, pbSeeds)
	})
}

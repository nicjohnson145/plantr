package controller

import (
	"context"
	"fmt"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/nicjohnson145/hlp"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/nicjohnson145/plantr/internal/parsingv2"
	"github.com/stretchr/testify/mock"
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
					Destination:     "~/seed-one",
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
					DisplayName: "seed-one",
				},
				Element: &pbv1.Seed_ConfigFile{
					ConfigFile: &pbv1.ConfigFile{
						Content:     "seed-one-content",
						Destination: "/tmp/someuser/seed-one",
					},
				},
			},
			{
				Metadata: &pbv1.Seed_Metadata{
					DisplayName: "~/seed-two",
				},
				Element: &pbv1.Seed_ConfigFile{
					ConfigFile: &pbv1.ConfigFile{
						Content:     "\nseed-one-installed\n",
						Destination: "/tmp/someuser/seed-two",
					},
				},
			},
		}
		pbEqual(t, wantPb, pbSeeds)
	})

	t.Run("secret references change hash", func(t *testing.T) {
		// Make an vault mock that returns different "secret data" on each call to simulate it being changed
		count := 0
		vault := NewMockVaultClient(t)
		vault.
			EXPECT().
			ReadSecretData(mock.Anything).
			RunAndReturn(func(ctx context.Context) (map[string]any, error) {
				count += 1
				return map[string]any{
					"foo": fmt.Sprintf("bar-%v", count),
				}, nil
			})

		ctrl, err := NewController(ControllerConfig{
			VaultClient: vault,
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
					TemplateContent: "hello im {{ .Vault.foo }}",
					Destination:     "~/seed-one",
				},
			},
		}

		firstPbSeeds, err := ctrl.renderSeeds(context.Background(), node, seeds)
		require.NoError(t, err)
		fmt.Println(firstPbSeeds[0].GetConfigFile().Content)

		secondPbSeeds, err := ctrl.renderSeeds(context.Background(), node, seeds)
		require.NoError(t, err)
		fmt.Println(secondPbSeeds[0].GetConfigFile().Content)

		// ensure that the hash for the config file is different between runs
		require.NotEqual(t, firstPbSeeds[0].Metadata.Hash, secondPbSeeds[0].Metadata.Hash)
	})
}

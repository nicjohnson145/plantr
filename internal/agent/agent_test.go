package agent

import (
	"context"
	"testing"

	controllerv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/stretchr/testify/require"
)

func TestSystemPackageUpdateRunsOnce(t *testing.T) {
	// Replace our "update function" with something that just tracks calls
	count := 0
	unitTestSystemUpdateFunc = func() error {
		count += 1
		return nil
	}
	t.Cleanup(func() {
		unitTestSystemUpdateFunc = nil
	})

	// Replace our "execute function" with just a noop
	unitTestExecuteFunc = func(s1 string, s2 ...string) (string, string, error) {
		return "", "", nil
	}
	t.Cleanup(func() {
		unitTestExecuteFunc = nil
	})

	a := NewAgent(AgentConfig{
		Inventory: NewNoopInventory(NoopInventoryConfig{}),
	})

	require.NoError(t, a.executeSeeds(context.Background(), []*controllerv1.Seed{
		{
			Metadata: &controllerv1.Seed_Metadata{
				DisplayName: "pkg-one",
			},
			Element: &controllerv1.Seed_SystemPackage{
				SystemPackage: &controllerv1.SystemPackage{
					Pkg: &controllerv1.SystemPackage_Apt{
						Apt: &controllerv1.SystemPackage_AptPkg{
							Name: "pkg-one",
						},
					},
				},
			},
		},
		{
			Metadata: &controllerv1.Seed_Metadata{
				DisplayName: "pkg-two",
			},
			Element: &controllerv1.Seed_SystemPackage{
				SystemPackage: &controllerv1.SystemPackage{
					Pkg: &controllerv1.SystemPackage_Apt{
						Apt: &controllerv1.SystemPackage_AptPkg{
							Name: "pkg-two",
						},
					},
				},
			},
		},
	}))

	require.Equal(t, 1, count)
}

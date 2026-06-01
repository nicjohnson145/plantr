package client

import (
	"context"
	"fmt"

	"github.com/nicjohnson145/hlp"
	controllerv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/rs/zerolog"
)

func getSystemPackageUpdateFunc(ctx context.Context, seeds []*controllerv1.Seed) (func() error, error) {
	// Only for unit testing purposes
	if unitTestSystemUpdateFunc != nil {
		return unitTestSystemUpdateFunc, nil
	}

	log := zerolog.Ctx(ctx)

	idx := hlp.First(seeds, func(x *controllerv1.Seed) bool {
		return x.GetSystemPackage() != nil
	})
	// If there are not system packages, then it doesnt matter, return noop function
	if idx == -1 {
		return func() error { return nil }, nil
	}

	// Otherwise, all the system packages should be the same, so figure out which package manager we are and return the
	// appropriate "update" command for that manager
	switch concrete := seeds[idx].GetSystemPackage().Pkg.(type) {
	case *controllerv1.SystemPackage_Apt:
		return func() error {
			log.Debug().Msg("executing `sudo apt update`")
			_, stderr, err := ExecuteOSCommand("/bin/sh", "-c", "sudo DEBIAN_FRONTEND=noninteractive apt update")
			if err != nil {
				return fmt.Errorf("error during update: %v\n%v", err, stderr)
			}
			return nil
		}, nil
	case *controllerv1.SystemPackage_Brew:
		return func() error {
			log.Warn().Msg("brew pre-update function not implemented yet")
			return nil
		}, nil
	case *controllerv1.SystemPackage_Pacman:
		return func() error {
			log.Debug().Msg("pacman system updates could be breaking or undesired, intentionally not updating remotes")
			return nil
		}, nil
	default:
		return nil, fmt.Errorf("unhandled package type %T", concrete)
	}
}

func executeSeed_systemPackage(ctx context.Context, pbseed *controllerv1.Seed) (*InventoryRow, error) {
	seed := pbseed.Element.(*controllerv1.Seed_SystemPackage).SystemPackage

	switch concrete := seed.Pkg.(type) {
	case *controllerv1.SystemPackage_Apt:
		return executeSeed_systemPackage_apt(ctx, concrete.Apt)
	case *controllerv1.SystemPackage_Brew:
		return executeSeed_systemPackage_brew(ctx, concrete.Brew)
	case *controllerv1.SystemPackage_Pacman:
		return executeSeed_systemPackage_pacman(ctx, concrete.Pacman)
	default:
		return nil, fmt.Errorf("unhandled system package type of %T", concrete)
	}
}

func executeSeed_systemPackage_apt(_ context.Context, pkg *controllerv1.SystemPackage_AptPkg) (*InventoryRow, error) {
	// TODO: proper version support
	_, stderr, err := ExecuteOSCommand("/bin/sh", "-c", fmt.Sprintf("sudo DEBIAN_FRONTEND=noninteractive apt install -y %v", pkg.Name))
	if err != nil {
		return nil, fmt.Errorf("error during installation: %v\n%v", err, stderr)
	}

	return &InventoryRow{
		Package: hlp.Ptr(pkg.Name),
	}, nil
}

func executeSeed_systemPackage_brew(_ context.Context, pkg *controllerv1.SystemPackage_BrewPkg) (*InventoryRow, error) {
	// TODO: proper version support & `brew update` cached for the whole run
	_, stderr, err := ExecuteOSCommand("brew", "install", pkg.Name)
	if err != nil {
		return nil, fmt.Errorf("error during installation: %v\n%v", err, stderr)
	}

	return &InventoryRow{
		Package: hlp.Ptr(pkg.Name),
	}, nil
}

func executeSeed_systemPackage_pacman(_ context.Context, pkg *controllerv1.SystemPackage_PacmanPkg) (*InventoryRow, error) {
	_, stderr, err := ExecuteOSCommand("/bin/sh", "-c", fmt.Sprintf("sudo pacman -S --noconfirm %v", pkg.Name))
	if err != nil {
		return nil, fmt.Errorf("error running pacman: %w\n%v", err, stderr)
	}
	return &InventoryRow{
		Package: hlp.Ptr(pkg.Name),
	}, nil
}

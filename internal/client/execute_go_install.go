package client

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/nicjohnson145/hlp"
	controllerv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
)

func executeSeed_goInstall(ctx context.Context, seed *controllerv1.Seed) (*InventoryRow, error) {
	install := seed.Element.(*controllerv1.Seed_GoInstall).GoInstall

	gopath, err := exec.LookPath("go")
	if err != nil {
		return nil, fmt.Errorf("go not found in $PATH")
	}

	version := "latest"
	if install.Version != nil {
		version = *install.Version
	}

	_, _, err = ExecuteOSCommand(gopath, "install", install.Package+"@"+version)
	if err != nil {
		return nil, fmt.Errorf("error installing package: %w", err)
	}

	return &InventoryRow{
		Package: hlp.Ptr(install.Package),
	}, nil
}

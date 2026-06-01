package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/carlmjohnson/requests"
	"github.com/nicjohnson145/hlp"
	controllerv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/rs/zerolog"
)

func executeSeed_golang(ctx context.Context, seed *controllerv1.Seed) (*InventoryRow, error) {
	golang := seed.Element.(*controllerv1.Seed_Golang).Golang

	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("golang install only available for linux OS")
	}

	log := zerolog.Ctx(ctx)

	log.Trace().Msg("removing existing installation")
	// make sure to clean out the old version first per the golang docs. Run this command through the shell so we can
	// elivate privileges
	_, _, err := ExecuteOSCommand("/bin/sh", "-c", "sudo rm -rf /usr/local/go")
	if err != nil {
		return nil, fmt.Errorf("error removing old golang installation: %w", err)
	}

	log.Trace().Msg("downloading release tarball")
	dir, err := os.MkdirTemp("", "plantr-golang-")
	if err != nil {
		return nil, fmt.Errorf("unable to make temp directory")
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	tarball := fmt.Sprintf("go%v.linux-%v.tar.gz", golang.Version, runtime.GOARCH)
	filepath := filepath.Join(dir, tarball)
	err = requests.
		URL(fmt.Sprintf("https://go.dev/dl/%v", tarball)).
		ToFile(filepath).
		Fetch(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error downloading tarball: %w", err)
	}

	log.Trace().Msg("extracting tarball")
	// Execute this through the shell so we can elevate privileges with sudo
	_, _, err = ExecuteOSCommand("/bin/sh", "-c", fmt.Sprintf("sudo tar -C /usr/local -xzf %v", filepath))
	if err != nil {
		return nil, fmt.Errorf("error unpacking tarball: %w", err)
	}

	return &InventoryRow{
		Path: hlp.Ptr("/usr/local/go"),
	}, nil
}

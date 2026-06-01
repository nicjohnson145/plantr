package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nicjohnson145/hlp"
	controllerv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
)

func executeSeed_configFile(ctx context.Context, pbseed *controllerv1.Seed) (*InventoryRow, error) {
	seed := pbseed.Element.(*controllerv1.Seed_ConfigFile).ConfigFile

	if err := os.MkdirAll(filepath.Dir(seed.Destination), 0755); err != nil {
		return nil, fmt.Errorf("error creating containing dir: %w", err)
	}

	modeVal, err := strconv.ParseUint("0"+seed.Mode, 8, 32)
	if err != nil {
		return nil, fmt.Errorf("error parsing filemode as base8 u32: %w", err)
	}

	if err := os.WriteFile(seed.Destination, []byte(seed.Content), os.FileMode(modeVal)); err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}

	return &InventoryRow{
		Path: hlp.Ptr(seed.Destination),
	}, nil
}

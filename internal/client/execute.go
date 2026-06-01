package client

import (
	"context"
	"errors"
	"fmt"
	"sync"

	controllerv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/rs/zerolog"
)

var (
	// escape hatch for unit tests to handle the "update" portion of system update
	unitTestSystemUpdateFunc func() error
)

type ExecuteSeedsRequest struct {
	Inventory InventoryClient
	Seeds     []*controllerv1.Seed
}

func ExecuteSeeds(ctx context.Context, req *ExecuteSeedsRequest) error {
	var errs []error

	noopSkip := func(_ *controllerv1.Seed) bool {
		return false
	}
	noopPreExecute := func() error {
		return nil
	}
	sysUpdateFunc, err := getSystemPackageUpdateFunc(ctx, req.Seeds)
	if err != nil {
		return fmt.Errorf("error getting system_package update function: %w", err)
	}
	preSystemUpdate := sync.OnceValue(func() error {
		return sysUpdateFunc()
	})

	log := zerolog.Ctx(ctx)

	for _, seed := range req.Seeds {
		namedError := func(err error, ctx string) error {
			return fmt.Errorf("%v: %v, %w", seed.Metadata.DisplayName, ctx, err)
		}

		var executeFunc func(context.Context, *controllerv1.Seed) (*InventoryRow, error)
		var skipInventoryFunc func(*controllerv1.Seed) bool
		var preExecuteFunc func() error
		var msg string

		switch concrete := seed.Element.(type) {
		case *controllerv1.Seed_ConfigFile:
			msg = fmt.Sprintf("rendering config file %v", seed.Metadata.DisplayName)
			executeFunc = executeSeed_configFile
			skipInventoryFunc = noopSkip
			preExecuteFunc = noopPreExecute
		case *controllerv1.Seed_GithubRelease:
			msg = fmt.Sprintf("downloading github_release %v", seed.Metadata.DisplayName)
			executeFunc = executeSeed_githubRelease
			skipInventoryFunc = noopSkip
			preExecuteFunc = noopPreExecute
		case *controllerv1.Seed_SystemPackage:
			msg = fmt.Sprintf("installing system_package %v", seed.Metadata.DisplayName)
			executeFunc = executeSeed_systemPackage
			skipInventoryFunc = noopSkip
			preExecuteFunc = preSystemUpdate
		case *controllerv1.Seed_GitRepo:
			msg = fmt.Sprintf("cloning git_repo %v", seed.Metadata.DisplayName)
			executeFunc = executeSeed_gitRepo
			skipInventoryFunc = noopSkip
			preExecuteFunc = noopPreExecute
		case *controllerv1.Seed_Golang:
			msg = fmt.Sprintf("installing %v", seed.Metadata.DisplayName)
			executeFunc = executeSeed_golang
			skipInventoryFunc = noopSkip
			preExecuteFunc = noopPreExecute
		case *controllerv1.Seed_GoInstall:
			msg = fmt.Sprintf("installing go binary %v", seed.Metadata.DisplayName)
			executeFunc = executeSeed_goInstall
			// If we're not specifying a version, that means "latest", so dont check inventory to guarantee that we try
			// it again
			skipInventoryFunc = func(s *controllerv1.Seed) bool {
				return s.GetGoInstall().Version == nil
			}
			preExecuteFunc = noopPreExecute
		case *controllerv1.Seed_UrlDownload:
			msg = fmt.Sprintf("downloading %v", seed.Metadata.DisplayName)
			executeFunc = executeSeed_urlDownload
			skipInventoryFunc = noopSkip
			preExecuteFunc = noopPreExecute
		default:
			log.Warn().Msgf("dropping unknown seed type %T", concrete)
			continue
		}

		log.Info().Msg(msg)

		if !skipInventoryFunc(seed) {
			row, err := req.Inventory.GetRow(ctx, seed.Metadata.Hash)
			if err != nil {
				return namedError(err, "error reading inventory")
			}
			if row != nil {
				log.Debug().Msg("already exists in inventory, skipping")
				continue
			}
		}

		if err := preExecuteFunc(); err != nil {
			return namedError(err, "error executing pre-execute function")
		}

		row, err := executeFunc(ctx, seed)
		if err != nil {
			errs = append(errs, namedError(err, "error executing"))
			continue
		}

		if row != nil {
			row.Hash = seed.Metadata.Hash
			if err := req.Inventory.WriteRow(ctx, *row); err != nil {
				errs = append(errs, namedError(err, "error writing to inventory"))
				continue
			}
		}
	}

	return errors.Join(errs...)
}

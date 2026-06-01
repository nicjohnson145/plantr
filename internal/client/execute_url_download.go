package client

import (
	"context"

	"github.com/nicjohnson145/hlp"
	controllerv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/rs/zerolog"
)

func executeSeed_urlDownload(ctx context.Context, pbseed *controllerv1.Seed) (*InventoryRow, error) {
	seed := pbseed.Element.(*controllerv1.Seed_UrlDownload).UrlDownload

	log := zerolog.Ctx(ctx)

	resp, err := DownloadFromUrl(ctx, &DownloadRequest{
		Logger:               *log,
		URL:                  seed.DownloadUrl,
		DestinationDirectory: seed.DestinationDirectory,
		NameOverride:         seed.NameOverride,
		PreserveArchive:      seed.ArchiveRelease,
	})
	if err != nil {
		return nil, err
	}

	return &InventoryRow{
		Path: hlp.Ptr(resp.DownloadPath),
	}, nil
}

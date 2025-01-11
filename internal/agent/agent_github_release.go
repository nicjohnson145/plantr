package agent

import (
	"context"

	"github.com/carlmjohnson/requests"
	"github.com/nicjohnson145/hlp"
	controllerv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
)

func (a *Agent) executeSeed_githubRelease(ctx context.Context, pbseed *controllerv1.Seed) (*InventoryRow, error) {
	seed := pbseed.Element.(*controllerv1.Seed_GithubRelease).GithubRelease

	resp, err := DownloadFromUrl(ctx, &DownloadRequest{
		Logger: a.log,
		Client: a.httpClient,
		URL:    seed.DownloadUrl,
		RequestModFunc: func(builder *requests.Builder) *requests.Builder {
			if seed.Authentication != nil && seed.Authentication.BearerAuth != "" {
				builder = builder.Header("Authorization", seed.Authentication.BearerAuth)
			}
			return builder
		},
		DestinationDirectory: seed.DestinationDirectory,
		PreserveArchive:      seed.ArchiveRelease,
		NameOverride:         seed.NameOverride,
	})
	if err != nil {
		return nil, err
	}

	return &InventoryRow{
		Path: hlp.Ptr(resp.DownloadPath),
	}, nil

}

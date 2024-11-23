package controller

import (
	"context"
	"fmt"
	"regexp"

	"github.com/carlmjohnson/requests"
	"github.com/nicjohnson145/hlp"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/nicjohnson145/plantr/internal/parsingv2"
)

type githubAsset struct {
	Name        string `json:"name"`
	DownloadUrl string `json:"browser_download_url"`
	Url         string `json:"url"`
}

func (c *Controller) renderSeed_githubRelease(release *parsingv2.GithubRelease, node *parsingv2.Node) (*pbv1.GithubRelease, error) {
	type response struct {
		Assets []githubAsset `json:"assets"`
	}

	var resp response
	builder := requests.
		URL("https://api.github.com").
		Pathf("repos/%v/releases/tags/%v", release.Repo, release.Tag).
		Client(c.httpClient).
		ToJSON(&resp)
	if c.githubReleaseToken == "" {
		c.log.Warn().Msg("making un-authenticated request to github API, this will likely result in being very quickly rate limited")
	} else {
		builder = builder.Header("Authorization", c.githubReleaseToken)
	}

	if err := builder.Fetch(context.Background()); err != nil {
		return nil, fmt.Errorf("error getting release assets: %w", err)
	}

	asset, err := c.getAssetForOSArch(release, node, resp.Assets)
	if err != nil {
		return nil, fmt.Errorf("error filtering release assets: %w", err)
	}

	return &pbv1.GithubRelease{
		DownloadUrl: asset.DownloadUrl,
	}, nil
}

var (
	regexMusl     = regexp.MustCompile("(?i)musl")
	regexLinuxPkg = regexp.MustCompile(`(?i)(\.deb|\.rpm|\.apk)$`)

	osRegexMap = map[string]*regexp.Regexp{
		"linux":   regexp.MustCompile("(?i)linux"),
		"darwin":  regexp.MustCompile(`(?i)(darwin|mac(os)?|apple|osx)`),
	}

	archRegexMap = map[string]*regexp.Regexp{
		"amd64": regexp.MustCompile(`(?i)(x86_64|amd64|x64)`),
		"arm64": regexp.MustCompile(`(?i)(arm64|aarch64)`),
	}
)

func (c *Controller) getAssetForOSArch(release *parsingv2.GithubRelease, node *parsingv2.Node, assets []githubAsset) (*githubAsset, error) {
	userPattern := release.GetAssetPattern(node.OS, node.Arch)
	if userPattern != nil {
		c.log.Debug().Msg("using user defined asset pattern")
		assets := c.filterAssets(assets, userPattern, true)
		if len(assets) != 1 {
			return nil, fmt.Errorf("expected 1 matching asset for user pattern, got %v", len(assets))
		}
	}

	c.log.Debug().Msg("no pattern given, attempting to auto-detect")

	return nil, nil
}

func (c *Controller) filterAssets(assets []githubAsset, pat *regexp.Regexp, match bool) ([]githubAsset) {
	return hlp.Filter(assets, func(x githubAsset, _ int) bool {
		return pat.MatchString(x.Name) == match
	})
}

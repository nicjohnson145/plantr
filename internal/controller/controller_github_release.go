package controller

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/carlmjohnson/requests"
	"github.com/nicjohnson145/hlp"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/nicjohnson145/plantr/internal/parsingv2"
)

var (
	ErrUnableToAutoDetectAssetError = errors.New("unable to auto-detect asset")
)

type githubTagResponse struct {
	Assets []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name        string `json:"name"`
	DownloadUrl string `json:"browser_download_url"`
	Url         string `json:"url"`
}

func (c *Controller) renderSeed_githubRelease(release *parsingv2.GithubRelease, node *parsingv2.Node) (*pbv1.GithubRelease, error) {
	var resp githubTagResponse
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
	regexMusl     = regexp.MustCompile(`(?i)musl`)
	regexChecksum = regexp.MustCompile(`(?i)\b(.sha256|.sha256sum)$`)
	regexLinuxPkg = regexp.MustCompile(`(?i)\b(\.deb|\.rpm|\.apk)$`)

	osRegexMap = map[string]*regexp.Regexp{
		"linux":  regexp.MustCompile(`(?i)\blinux`),
		"darwin": regexp.MustCompile(`(?i)\b(darwin|mac(os)?|apple|osx)`),
	}

	archRegexMap = map[string]*regexp.Regexp{
		"amd64": regexp.MustCompile(`(?i)\b(x86_64|amd64|x64)`),
		"arm64": regexp.MustCompile(`(?i)\b(arm64|aarch64)`),
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

	type filterStep struct {
		function    func() (*regexp.Regexp, error)
		shouldMatch bool
		msg         string
	}

	steps := []filterStep{
		{
			function: func() (*regexp.Regexp, error) {
				osPatt, ok := osRegexMap[node.OS]
				if !ok {
					return nil, fmt.Errorf("no pre-made patterns for OS %v", node.OS)
				}
				return osPatt, nil
			},
			shouldMatch: true,
			msg:         "attempting to filter assets by OS",
		},
		{
			function: func() (*regexp.Regexp, error) {
				archPat, ok := archRegexMap[node.Arch]
				if !ok {
					return nil, fmt.Errorf("no pre-made patterns for ARCH %v", node.Arch)
				}
				return archPat, nil
			},
			shouldMatch: true,
			msg:         "attempting to filter assets by OS",
		},
		{
			function: func() (*regexp.Regexp, error) {
				return regexLinuxPkg, nil
			},
			shouldMatch: false,
			msg:         "attempting to filter off 'package' assets",
		},
		{
			function: func() (*regexp.Regexp, error) {
				return regexChecksum, nil
			},
			shouldMatch: false,
			msg:         "attempting to filter off checksum assets",
		},
		{
			function: func() (*regexp.Regexp, error) {
				if node.OS == "linux" {
					return regexMusl, nil
				}
				return nil, nil
			},
			shouldMatch: true,
			msg:         "linux OS detected, attempting to select musl variant",
		},
	}

	filteredAssets := assets
	for _, step := range steps {
		patt, err := step.function()
		if err != nil {
			return nil, err
		}
		if patt == nil {
			continue
		}
		c.log.Trace().Msg(step.msg)
		filteredAssets = c.filterAssets(filteredAssets, patt, step.shouldMatch)
		if len(filteredAssets) == 1 {
			return &filteredAssets[0], nil
		}
		if len(filteredAssets) == 0 {
			return nil, ErrUnableToAutoDetectAssetError
		}
	}

	c.log.Trace().Msg("filtering exhausted, giving up")
	return nil, ErrUnableToAutoDetectAssetError
}

func (c *Controller) filterAssets(assets []githubAsset, pat *regexp.Regexp, match bool) []githubAsset {
	return hlp.Filter(assets, func(x githubAsset, _ int) bool {
		return pat.MatchString(x.Name) == match
	})
}

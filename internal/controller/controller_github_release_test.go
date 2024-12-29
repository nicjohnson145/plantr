package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/nicjohnson145/plantr/internal/parsingv2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGithubRelease_GetAssetForOSArch(t *testing.T) {
	testData := []struct {
		file string
		os   string
		arch string
		want string
	}{
		{
			file: "ripgrep-14.1.1",
			os:   "linux",
			arch: "amd64",
			want: "ripgrep-14.1.1-x86_64-unknown-linux-musl.tar.gz",
		},
		{
			file: "ripgrep-14.1.1",
			os:   "darwin",
			arch: "arm64",
			want: "ripgrep-14.1.1-aarch64-apple-darwin.tar.gz",
		},
		{
			file: "jq-jq-1.7.1",
			os:   "linux",
			arch: "amd64",
			want: "jq-linux-amd64",
		},
		{
			file: "jq-jq-1.7.1",
			os:   "darwin",
			arch: "arm64",
			want: "jq-macos-arm64",
		},
		{
			file: "zellij-v0.41.2",
			os:   "linux",
			arch: "amd64",
			want: "zellij-x86_64-unknown-linux-musl.tar.gz",
		},
		{
			file: "zellij-v0.41.2",
			os:   "darwin",
			arch: "arm64",
			want: "zellij-aarch64-apple-darwin.tar.gz",
		},
	}

	ctrl := &Controller{
		log: zerolog.New(os.Stdout).Level(zerolog.TraceLevel),
	}

	for _, tc := range testData {
		t.Run(tc.file, func(t *testing.T) {
			content, err := os.ReadFile("./testdata/get-assets-for-os-arch/" + tc.file + ".json")
			require.NoError(t, err)

			var resp githubTagResponse
			require.NoError(t, json.Unmarshal(content, &resp))

			got, err := ctrl.getAssetForOSArch(
				&parsingv2.GithubRelease{},
				&parsingv2.Node{
					OS:   tc.os,
					Arch: tc.arch,
				},
				resp.Assets,
			)
			require.NoError(t, err)
			require.Equal(t, tc.want, got.Name)
		})
	}
}

func TestGithubRelease_AssetUrlCaching(t *testing.T) {
	t.Parallel()

	const (
		hash        = "some-md5-string"
		nodeOS      = "linux"
		nodeArch    = "amd64"
		releaseRepo = "release-repo"
		releaseTag  = "release-tag"
		downloadURL = "some-download-url"
	)

	t.Run("cache hit", func(t *testing.T) {
		t.Parallel()

		// Intentionally dont set any resolvers here, we shouldnt call the GitHub API on a cache hit
		mockTransport := httpmock.NewMockTransport()

		mockStore := NewMockStorageClient(t)
		mockStore.
			EXPECT().
			ReadGithubReleaseAsset(mock.Anything, &DBGithubRelease{
				Hash: hash,
				OS:   nodeOS,
				Arch: nodeArch,
			}).
			Return(downloadURL, nil)

		ctrl, err := NewController(ControllerConfig{
			HttpClient: &http.Client{
				Transport: mockTransport,
			},
			StorageClient: mockStore,
			HashFunc: func(s *parsingv2.Seed) string {
				return hash
			},
		})
		require.NoError(t, err)

		release, err := ctrl.renderSeed_githubRelease(context.Background(), &parsingv2.GithubRelease{}, &parsingv2.Node{
			OS:   nodeOS,
			Arch: nodeArch,
		})
		require.NoError(t, err)

		require.Equal(t, downloadURL, release.DownloadUrl)
	})

	t.Run("cache miss", func(t *testing.T) {
		t.Parallel()

		mockTransport := httpmock.NewMockTransport()
		mockTransport.RegisterResponder(
			http.MethodGet,
			fmt.Sprintf("https://api.github.com/repos/%v/releases/tags/%v", releaseRepo, releaseTag),
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				map[string]any{
					"assets": []map[string]any{
						{
							"name":                 "some-binary-linux-amd64",
							"browser_download_url": downloadURL,
						},
					},
				},
			),
		)

		mockStore := NewMockStorageClient(t)
		mockStore.
			EXPECT().
			ReadGithubReleaseAsset(mock.Anything, &DBGithubRelease{
				Hash: hash,
				OS:   nodeOS,
				Arch: nodeArch,
			}).
			Return("", nil)

		mockStore.
			EXPECT().
			WriteGithubReleaseAsset(mock.Anything, &DBGithubRelease{
				Hash:        hash,
				OS:          nodeOS,
				Arch:        nodeArch,
				DownloadURL: downloadURL,
			}).
			Return(nil)

		ctrl, err := NewController(ControllerConfig{
			HttpClient: &http.Client{
				Transport: mockTransport,
			},
			StorageClient: mockStore,
			HashFunc: func(s *parsingv2.Seed) string {
				return hash
			},
		})
		require.NoError(t, err)

		release, err := ctrl.renderSeed_githubRelease(
			context.Background(),
			&parsingv2.GithubRelease{
				Repo: releaseRepo,
				Tag:  releaseTag,
			},
			&parsingv2.Node{
				OS:   nodeOS,
				Arch: nodeArch,
			},
		)
		require.NoError(t, err)

		require.Equal(t, downloadURL, release.DownloadUrl)
	})
}

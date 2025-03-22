package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/jarcoal/httpmock"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/nicjohnson145/plantr/internal/parsingv2"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestGithubRelease_GetAssetForOSArch(t *testing.T) {
	testData := []struct {
		file          string
		os            string
		arch          string
		want          string
		assetPatterns map[string]map[string]*regexp.Regexp
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
		{
			file: "neovim-v0.10.3",
			os:   "linux",
			arch: "amd64",
			want: "nvim-linux64.tar.gz",
		},
		{
			file: "curlie-v1.7.2",
			os:   "linux",
			arch: "amd64",
			want: "curlie_1.7.2_linux_amd64.tar.gz",
		},
		{
			file: "minikube-v1.35.0",
			os:   "linux",
			arch: "amd64",
			want: "minikube-linux-amd64",
			assetPatterns: map[string]map[string]*regexp.Regexp{
				"linux": {
					"amd64": regexp.MustCompile(`^minikube-linux-amd64$`),
				},
			},
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
				&parsingv2.GithubRelease{
					AssetPatterns: tc.assetPatterns,
				},
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

	t.Run("smokes", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Cleanup(viper.Reset)
		viper.Set(SqliteDBPath, filepath.Join(tmpDir, "some-db.sqlite"))
		viper.Set(StorageType, StorageKindSqlite.String())

		storageClient, storageCleanup, err := NewStorageClientFromEnv(zerolog.New(os.Stderr))
		require.NoError(t, err)
		t.Cleanup(storageCleanup)

		url := fmt.Sprintf("https://api.github.com/repos/%v/releases/tags/%v", releaseRepo, releaseTag)
		mockTransport := httpmock.NewMockTransport()
		mockTransport.RegisterResponder(
			http.MethodGet,
			url,
			httpmock.NewJsonResponderOrPanic(http.StatusOK, map[string]any{
				"assets": []map[string]any{
					{
						"name":                 "some-bin-linux-amd64",
						"browser_download_url": downloadURL,
					},
				},
			}),
		)

		ctrl, err := NewController(ControllerConfig{
			StorageClient: storageClient,
			HttpClient: &http.Client{
				Transport: mockTransport,
			},
		})
		require.NoError(t, err)

		// Execute it once to set the cache
		got, err := ctrl.renderSeed_githubRelease(
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
		require.Equal(
			t,
			&pbv1.Seed{
				Element: &pbv1.Seed_GithubRelease{
					GithubRelease: &pbv1.GithubRelease{
						DownloadUrl: downloadURL,
					},
				},
			},
			got,
		)

		// Execute it again to ensure the cache is read from
		got, err = ctrl.renderSeed_githubRelease(
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
		require.Equal(
			t,
			&pbv1.Seed{
				Element: &pbv1.Seed_GithubRelease{
					GithubRelease: &pbv1.GithubRelease{
						DownloadUrl: downloadURL,
					},
				},
			},
			got,
		)

		// Make sure we only hit the GH url once
		require.Equal(
			t,
			map[string]int{
				"GET " + url: 1,
			},
			mockTransport.GetCallCountInfo(),
		)
	})
}

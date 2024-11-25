package controller

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/nicjohnson145/plantr/internal/parsingv2"
	"github.com/rs/zerolog"
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

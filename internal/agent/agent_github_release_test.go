package agent

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/nicjohnson145/hlp"
	controllerv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/stretchr/testify/require"
)

func TestExecuteGithubRelease(t *testing.T) {
	t.Run("archive", func(t *testing.T) {
		const (
			downloadURL = "http://fake-place.example.com/downloads/bat-v0.24.0-x86_64-unknown-linux-musl.tar.gz"
		)

		tarBytes, err := os.ReadFile("./testdata/github-release/bat-v0.24.0-x86_64-unknown-linux-musl.tar.gz")
		require.NoError(t, err)

		tmpDir := t.TempDir()
		destDir := filepath.Join(tmpDir, "bin")

		mockTransport := httpmock.NewMockTransport()
		mockTransport.RegisterResponder(
			http.MethodGet,
			downloadURL,
			httpmock.NewBytesResponder(http.StatusOK, tarBytes),
		)

		a := NewAgent(AgentConfig{
			HTTPClient: &http.Client{
				Transport: mockTransport,
			},
			Inventory: NewNoopInventory(NoopInventoryConfig{}),
		})

		require.NoError(t, a.executeSeed_githubRelease(
			context.Background(),
			&controllerv1.GithubRelease{
				DownloadUrl:          downloadURL,
				DestinationDirectory: destDir,
			},
			&controllerv1.Seed_Metadata{},
		))

		// <tmp>/bin/bat should now exist
		_, err = os.Stat(filepath.Join(destDir, "bat"))
		require.NoError(t, err)
	})

	t.Run("archive - name override", func(t *testing.T) {
		const (
			downloadURL = "http://fake-place.example.com/downloads/bat-v0.24.0-x86_64-unknown-linux-musl.tar.gz"
		)

		tarBytes, err := os.ReadFile("./testdata/github-release/bat-v0.24.0-x86_64-unknown-linux-musl.tar.gz")
		require.NoError(t, err)

		tmpDir := t.TempDir()
		destDir := filepath.Join(tmpDir, "bin")

		mockTransport := httpmock.NewMockTransport()
		mockTransport.RegisterResponder(
			http.MethodGet,
			downloadURL,
			httpmock.NewBytesResponder(http.StatusOK, tarBytes),
		)

		a := NewAgent(AgentConfig{
			HTTPClient: &http.Client{
				Transport: mockTransport,
			},
			Inventory: NewNoopInventory(NoopInventoryConfig{}),
		})

		require.NoError(t, a.executeSeed_githubRelease(
			context.Background(),
			&controllerv1.GithubRelease{
				DownloadUrl:          downloadURL,
				DestinationDirectory: destDir,
				NameOverride:         hlp.Ptr("bat2"),
			},
			&controllerv1.Seed_Metadata{},
		))

		// <tmp>/bin/bat should now exist
		_, err = os.Stat(filepath.Join(destDir, "bat2"))
		require.NoError(t, err)
	})

	t.Run("archive release", func(t *testing.T) {
		const (
			downloadURL = "http://fake-place.example.com/downloads/nvim-linux64.tar.gz"
		)

		tarBytes, err := os.ReadFile("./testdata/github-release/nvim-linux64.tar.gz")
		require.NoError(t, err)

		tmpDir := t.TempDir()
		destDir := filepath.Join(tmpDir, "bin")

		mockTransport := httpmock.NewMockTransport()
		mockTransport.RegisterResponder(
			http.MethodGet,
			downloadURL,
			httpmock.NewBytesResponder(http.StatusOK, tarBytes),
		)

		a := NewAgent(AgentConfig{
			HTTPClient: &http.Client{
				Transport: mockTransport,
			},
			Inventory: NewNoopInventory(NoopInventoryConfig{}),
		})

		require.NoError(t, a.executeSeed_githubRelease(
			context.Background(),
			&controllerv1.GithubRelease{
				DownloadUrl:          downloadURL,
				DestinationDirectory: destDir,
				NameOverride:         hlp.Ptr("neovim"),
				ArchiveRelease:       true,
			},
			&controllerv1.Seed_Metadata{},
		))

		// <tmp>/bin/neovim should exist and be a directory
		info, err := os.Stat(filepath.Join(destDir, "neovim"))
		require.NoError(t, err)
		require.True(t, info.IsDir())
		// Inside it should be the binary structure
		_, err = os.Stat(filepath.Join(destDir, "neovim", "bin", "nvim"))
		require.NoError(t, err)
	})
}

func TestFullTrimSuffix(t *testing.T) {
	testData := []struct {
		name     string
		expected string
	}{
		{
			name:     "some-dir.tar.gz",
			expected: "some-dir",
		},
		{
			name:     "some-dir.bz2",
			expected: "some-dir",
		},
		{
			name:     "some-dir",
			expected: "some-dir",
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, fullTrimSuffix(tc.name))
		})
	}
}

package agent

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
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
		})

		require.NoError(t, a.executeSeed_githubRelease(&controllerv1.GithubRelease{
			DownloadUrl:          downloadURL,
			DestinationDirectory: destDir,
		}))

		// <tmp>/bin/bat should now exist
		_, err = os.Stat(filepath.Join(destDir, "bat"))
		require.NoError(t, err)
	})
}

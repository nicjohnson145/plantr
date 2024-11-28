package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/carlmjohnson/requests"
	"github.com/nicjohnson145/hlp/set"
	controllerv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
)

var (
	archiveExtensions = set.New(
		".gz",
		".zip",
	)
)

func (a *Agent) executeSeed_githubRelease(seed *controllerv1.GithubRelease) error {
	// TOOD: inventory tracking
	if err := os.MkdirAll(seed.DestinationDirectory, 0775); err != nil {
		return fmt.Errorf("error creating binary directory: %w", err)
	}

	dir, err := os.MkdirTemp("", "plantr-agent")
	if err != nil {
		return fmt.Errorf("error creating scratch directory: %w", err)
	}
	defer os.RemoveAll(dir)

	tmpPath := filepath.Join(dir, filepath.Base(seed.DownloadUrl))
	builder := requests.URL(seed.DownloadUrl).ToFile(tmpPath)
	if seed.Authentication != nil && seed.Authentication.BearerAuth != "" {
		builder = builder.Header("Authorization", seed.Authentication.BearerAuth)
	}
	if err := builder.Fetch(context.Background()); err != nil {
		return fmt.Errorf("error executing download: %w", err)
	}

	var binaryContent []byte
	var destName string
	if archiveExtensions.Contains(filepath.Ext(tmpPath)) {
		return fmt.Errorf("archives not implemented yet")
	} else {
		content, err := os.ReadFile(tmpPath)
		if err != nil {
			return fmt.Errorf("error reading file contents: %w", err)
		}
		binaryContent = content
		destName = filepath.Base(tmpPath)
	}

	if seed.BinaryNameOverride != nil {
		destName = *seed.BinaryNameOverride
	}

	outPath := filepath.Join(seed.DestinationDirectory, destName)
	if err := os.WriteFile(outPath, binaryContent, 0755); err != nil {
		return fmt.Errorf("error writing final output path: %w", err)
	}

	return nil
}

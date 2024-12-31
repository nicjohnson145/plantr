package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/carlmjohnson/requests"
	"github.com/mholt/archives"
	"github.com/nicjohnson145/hlp"
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

	filename := filepath.Base(seed.DownloadUrl)
	tmpPath := filepath.Join(dir, filename)
	builder := requests.
		URL(seed.DownloadUrl).
		ToFile(tmpPath).
		Client(a.httpClient)

	if seed.Authentication != nil && seed.Authentication.BearerAuth != "" {
		builder = builder.Header("Authorization", seed.Authentication.BearerAuth)
	}
	if err := builder.Fetch(context.Background()); err != nil {
		return fmt.Errorf("error executing download: %w", err)
	}

	var binaryContent []byte
	var destName string
	if archiveExtensions.Contains(filepath.Ext(tmpPath)) {
		fl, err := os.Open(tmpPath)
		if err != nil {
			return fmt.Errorf("error opening file for reading: %w", err)
		}

		archive, stream, err := archives.Identify(context.Background(), filename, fl)
		if err != nil {
			return fmt.Errorf("error detecting archive type: %w", err)
		}

		extractor, ok := archive.(archives.Extractor)
		if !ok {
			return fmt.Errorf("does not implement Extractor, cannot procede")
		}

		executableFiles := map[string][]byte{}

		err = extractor.Extract(context.Background(), stream, func(ctx context.Context, info archives.FileInfo) error {
			if info.Mode().IsDir() {
				return nil
			}

			ownerExecutable := info.Mode()&0100 != 0

			if !ownerExecutable {
				return nil
			}

			fl, err := info.Open()
			if err != nil {
				return fmt.Errorf("error opening file: %w", err)
			}
			defer fl.Close()

			flBytes, err := io.ReadAll(fl)
			if err != nil {
				return fmt.Errorf("error reading file contents: %w", err)
			}

			executableFiles[info.NameInArchive] = flBytes
			return nil
		})
		if err != nil {
			return fmt.Errorf("error during extraction: %w", err)
		}

		if len(executableFiles) != 1 {
			return fmt.Errorf("expected to find 1 executable file, instead found %v", len(executableFiles))
		}

		name := hlp.Keys(executableFiles)[0]
		binaryContent = executableFiles[name]
		destName = filepath.Base(name)
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
	if err := os.WriteFile(outPath, binaryContent, 0755); err != nil { //nolint: gosec // it has to be executable, its an executable binary
		return fmt.Errorf("error writing final output path: %w", err)
	}

	return nil
}

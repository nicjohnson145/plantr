package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

func (a *Agent) executeSeed_githubRelease(ctx context.Context, seed *controllerv1.GithubRelease, metadata *controllerv1.Seed_Metadata) error {
	row, err := a.inventory.GetRow(ctx, metadata.Hash)
	if err != nil {
		return fmt.Errorf("error checking inventory: %w", err)
	}

	if row != nil {
		a.log.Debug().Msg("release present in inventory, skipping")
		return nil
	}

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

	var extractor archives.Extractor
	var stream io.Reader
	var binaryContent []byte
	var destName string

	isArchive := archiveExtensions.Contains(filepath.Ext(tmpPath))

	if isArchive {
		fl, err := os.Open(tmpPath)
		if err != nil {
			return fmt.Errorf("error opening file for reading: %w", err)
		}
		defer fl.Close()

		a, s, err := archives.Identify(ctx, filename, fl)
		if err != nil {
			return fmt.Errorf("error detecting archive type: %w", err)
		}

		ex, ok := a.(archives.Extractor)
		if !ok {
			return fmt.Errorf("does not implement Extractor, cannot procde")
		}

		extractor = ex
		stream = s
	}

	// TOOD: refactor for complexity here, this is kinda gross
	if isArchive && seed.ArchiveRelease { // we should extract the archive, maintaining its structure
		targetName := fullTrimSuffix(filepath.Base(tmpPath))
		targetDir := targetName
		if seed.NameOverride != nil {
			targetDir = *seed.NameOverride
		}
		targetPath := filepath.Join(seed.DestinationDirectory, targetDir)

		if err := os.MkdirAll(targetPath, 0775); err != nil {
			return fmt.Errorf("error making target extraction directory: %w", err)
		}

		err = extractor.Extract(ctx, stream, func(ctx context.Context, info archives.FileInfo) error {
			infoPath := strings.TrimPrefix(info.NameInArchive, targetName + "/")
			// i.e its the top level directory
			if infoPath == "" {
				return nil
			}

			dstPath := filepath.Join(targetPath, infoPath)
			if info.IsDir() {
				if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
					return fmt.Errorf("4rror replicating directory from archive: %w", err)
				}
				return nil
			}
			fl, err := info.Open()
			if err != nil {
				return fmt.Errorf("error opening file in archive: %w", err)
			}
			defer fl.Close()

			dstFile, err := os.Create(dstPath)
			if err != nil {
				return fmt.Errorf("error creating destination file: %w", err)
			}
			defer dstFile.Close()

			if _, err := io.Copy(dstFile, fl); err != nil {
				return fmt.Errorf("error copying file: %w", err)
			}
			if err := os.Chmod(dstPath, info.Mode()); err != nil {
				return fmt.Errorf("error copying permissions: %w", err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("error extracting archive: %w", err)
		}

		return nil
	} else if isArchive { // we should extract a single binary from the archive
		executableFiles := map[string][]byte{}
		err := extractor.Extract(ctx, stream, func(ctx context.Context, info archives.FileInfo) error {
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
			return fmt.Errorf("error extracting archive: %w", err)
		}
		if len(executableFiles) != 1 {
			return fmt.Errorf("expected to find 1 executable file, instead found %v", len(executableFiles))
		}
		name := hlp.Keys(executableFiles)[0]
		binaryContent = executableFiles[name]
		destName = filepath.Base(name)
	} else { // the asset is already only a single binary
		content, err := os.ReadFile(tmpPath)
		if err != nil {
			return fmt.Errorf("error reading file contents: %w", err)
		}
		binaryContent = content
		destName = filepath.Base(tmpPath)
	}

	if seed.NameOverride != nil {
		destName = *seed.NameOverride
	}

	outPath := filepath.Join(seed.DestinationDirectory, destName)
	if err := os.WriteFile(outPath, binaryContent, 0755); err != nil { //nolint: gosec // it has to be executable, its an executable binary
		return fmt.Errorf("error writing final output path: %w", err)
	}

	err = a.inventory.WriteRow(ctx, InventoryRow{
		Hash: metadata.Hash,
		Path: hlp.Ptr(outPath),
	})
	if err != nil {
		return fmt.Errorf("error writing release to inventory: %w", err)
	}

	return nil
}

func fullTrimSuffix(name string) string {
	ext := "starter"
	base := name
	for ext != "" {
		newExt := filepath.Ext(base)
		base = base[:len(base)-len(newExt)]
		ext = newExt
	}
	return base
}

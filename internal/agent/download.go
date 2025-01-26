package agent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/mholt/archives"
	"github.com/nicjohnson145/hlp"
	"github.com/nicjohnson145/hlp/set"
	"github.com/rs/zerolog"
)

var (
	archiveExtensions = set.New(
		".gz",
		".zip",
		".xz",
	)
)

type DownloadRequest struct {
	Logger               zerolog.Logger
	Client               *http.Client
	URL                  string
	RequestModFunc       func(builder *requests.Builder) *requests.Builder
	DestinationDirectory string
	PreserveArchive      bool
	NameOverride         *string
	BinaryRegex          *string
}

type DownloadResponse struct {
	DownloadPath string
}

func DownloadFromUrl(ctx context.Context, req *DownloadRequest) (*DownloadResponse, error) {
	req.Logger.Trace().Msg("ensuring destination directory")
	if err := os.MkdirAll(req.DestinationDirectory, 0775); err != nil {
		return nil, fmt.Errorf("error creating destination directory: %w", err)
	}

	req.Logger.Trace().Msg("creating temp directory to land download")
	tmpDir, err := os.MkdirTemp("", "plantr-agent")
	if err != nil {
		return nil, fmt.Errorf("error creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Base(req.URL)
	tmpPath := filepath.Join(tmpDir, filename)
	builder := requests.
		URL(req.URL).
		ToFile(tmpPath).
		Client(req.Client)

	if req.RequestModFunc != nil {
		builder = req.RequestModFunc(builder)
	}

	req.Logger.Trace().Msg("executing request")
	if err := builder.Fetch(ctx); err != nil {
		return nil, fmt.Errorf("error executing download request: %w", err)
	}

	var extractor archives.Extractor
	var stream io.Reader
	var binaryContent []byte
	var destName string

	isArchive := archiveExtensions.Contains(filepath.Ext(filename))

	if isArchive {
		req.Logger.Trace().Msg("archive file detected, identifying and opening")
		fl, err := os.Open(tmpPath)
		if err != nil {
			return nil, fmt.Errorf("error opening file for reading: %w", err)
		}
		defer fl.Close()

		a, s, err := archives.Identify(ctx, filename, fl)
		if err != nil {
			return nil, fmt.Errorf("error detecting archive type: %w", err)
		}

		ex, ok := a.(archives.Extractor)
		if !ok {
			return nil, fmt.Errorf("does not implement Extractor, cannot procede")
		}

		extractor = ex
		stream = s
	}

	// TOOD: refactor for complexity here, this is kinda gross
	if isArchive && req.PreserveArchive {
		targetName := fullTrimSuffix(filename)
		targetDir := targetName
		if req.NameOverride != nil {
			targetDir = *req.NameOverride
		}
		targetPath := filepath.Join(req.DestinationDirectory, targetDir)

		if err := os.MkdirAll(targetPath, 0775); err != nil {
			return nil, fmt.Errorf("error making target extraction directory: %w", err)
		}

		err = extractor.Extract(ctx, stream, func(ctx context.Context, info archives.FileInfo) error {
			infoPath := strings.TrimPrefix(info.NameInArchive, targetName+"/")
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
			return nil, fmt.Errorf("error extracting archive: %w", err)
		}

		return &DownloadResponse{
			DownloadPath: targetPath,
		}, nil
	} else if isArchive { // we should extract a single binary from the archive
		binaryRegexFunc := func(path string) bool {
			return true
		}
		if req.BinaryRegex != nil {
			userRegex, err := regexp.Compile(*req.BinaryRegex)
			if err != nil {
				return nil, fmt.Errorf("error compiling user supplied regex: %w", err)
			}

			binaryRegexFunc = func(path string) bool {
				return userRegex.MatchString(path)
			}
		}

		executableFiles := map[string][]byte{}
		err := extractor.Extract(ctx, stream, func(ctx context.Context, info archives.FileInfo) error {
			if info.Mode().IsDir() {
				return nil
			}

			ownerExecutable := info.Mode()&0100 != 0

			if !ownerExecutable {
				return nil
			}

			if !binaryRegexFunc(info.NameInArchive) {
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
			return nil, fmt.Errorf("error extracting archive: %w", err)
		}
		if len(executableFiles) != 1 {
			return nil, fmt.Errorf("expected to find 1 executable file, instead found %v", len(executableFiles))
		}
		name := hlp.Keys(executableFiles)[0]
		binaryContent = executableFiles[name]
		destName = filepath.Base(name)
	} else { // the asset is already only a single binary
		content, err := os.ReadFile(tmpPath)
		if err != nil {
			return nil, fmt.Errorf("error reading file contents: %w", err)
		}
		binaryContent = content
		destName = filepath.Base(tmpPath)
	}

	if req.NameOverride != nil {
		destName = *req.NameOverride
	}

	outPath := filepath.Join(req.DestinationDirectory, destName)
	if err := os.WriteFile(outPath, binaryContent, 0755); err != nil { //nolint: gosec // it has to be executable, its an executable binary
		return nil, fmt.Errorf("error writing final output path: %w", err)
	}

	return &DownloadResponse{
		DownloadPath: outPath,
	}, nil
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

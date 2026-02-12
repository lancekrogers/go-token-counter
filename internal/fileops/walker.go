package fileops

import (
	"context"
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"

	"github.com/Obedience-Corp/go-token-counter/internal/errors"
)

// WalkResult contains information about walked files.
type WalkResult struct {
	Files         []string
	TotalFiles    int
	SkippedBinary int
	SkippedIgnore int
}

// WalkDirectory recursively walks a directory, respecting .gitignore files
// and filtering out binary files.
func WalkDirectory(ctx context.Context, rootPath string) (*WalkResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &WalkResult{
		Files: []string{},
	}

	gitignoreFile := filepath.Join(rootPath, ".gitignore")
	var gi *gitignore.GitIgnore
	if _, err := os.Stat(gitignoreFile); err == nil {
		gi, err = gitignore.CompileIgnoreFile(gitignoreFile)
		if err != nil {
			return nil, errors.Parse("parsing .gitignore", err).WithField("path", gitignoreFile)
		}
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		result.TotalFiles++

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		if gi != nil && gi.MatchesPath(relPath) {
			result.SkippedIgnore++
			return nil
		}

		isBinary, err := IsBinaryFile(path)
		if err != nil {
			result.SkippedBinary++
			return nil
		}
		if isBinary {
			result.SkippedBinary++
			return nil
		}

		result.Files = append(result.Files, path)
		return nil
	})

	if err != nil {
		return nil, errors.IO("walking directory", err).WithField("path", rootPath)
	}

	return result, nil
}

// AggregateFileContents reads all files and returns combined content.
func AggregateFileContents(ctx context.Context, files []string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var totalContent []byte

	for _, file := range files {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return nil, errors.IO("reading file", err).WithField("path", file)
		}
		totalContent = append(totalContent, content...)
	}

	return totalContent, nil
}

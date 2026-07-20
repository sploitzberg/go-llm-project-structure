// Copyright (c) 2026 sploitzberg
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func gitTopLevel(ctx context.Context, directory string) (string, error) {
	output, err := gitOutput(ctx, directory, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("find Git repository root: %w", err)
	}
	return filepath.Clean(strings.TrimSpace(string(output))), nil
}

func goModulePath(ctx context.Context, root string) (string, error) {
	output, err := goOutput(ctx, root, "list", "-m", "-f={{.Path}}")
	if err != nil {
		return "", fmt.Errorf("derive current module with go list -m: %w", err)
	}
	module := strings.TrimSpace(string(output))
	if module == "" {
		return "", fmt.Errorf("derive current module with go list -m: empty module path")
	}
	return module, nil
}

func ensureTrackedWorktreeClean(ctx context.Context, root string) error {
	output, err := gitOutput(ctx, root, "status", "--porcelain=v1", "--untracked-files=no")
	if err != nil {
		return fmt.Errorf("inspect tracked worktree state: %w", err)
	}
	if len(output) == 0 {
		return nil
	}
	return fmt.Errorf("tracked worktree is dirty; commit or stash tracked changes first:\n%s", strings.TrimSpace(string(output)))
}

func candidateFiles(ctx context.Context, root string, patterns []string) ([]string, error) {
	const fixedArgumentCount = 6

	arguments := make([]string, 0, fixedArgumentCount+2*len(patterns))
	arguments = append(arguments, "grep", "-I", "-l", "-z", "-F")
	for _, pattern := range patterns {
		arguments = append(arguments, "-e", pattern)
	}
	arguments = append(arguments, "--")

	output, err := gitOutput(ctx, root, arguments...)
	if err != nil {
		if exitError, ok := errors.AsType[*exec.ExitError](err); ok && exitError.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("find tracked text files: %w", err)
	}

	var files []string
	for rawPath := range bytes.SplitSeq(output, []byte{0}) {
		if len(rawPath) > 0 {
			files = append(files, filepath.FromSlash(string(rawPath)))
		}
	}
	slices.Sort(files)
	return files, nil
}

func trackedFilesUnder(ctx context.Context, root, relativePath string) ([]string, error) {
	output, err := gitOutput(ctx, root, "ls-files", "-z", "--", filepath.ToSlash(relativePath))
	if err != nil {
		return nil, fmt.Errorf("list tracked files under %s: %w", relativePath, err)
	}

	var files []string
	for rawPath := range bytes.SplitSeq(output, []byte{0}) {
		if len(rawPath) > 0 {
			files = append(files, filepath.FromSlash(string(rawPath)))
		}
	}
	return files, nil
}

func isTracked(ctx context.Context, root, relativePath string) (bool, error) {
	_, err := gitOutput(ctx, root, "ls-files", "--error-unmatch", "--", filepath.ToSlash(relativePath))
	if err == nil {
		return true, nil
	}
	if exitError, ok := errors.AsType[*exec.ExitError](err); ok && exitError.ExitCode() == 1 {
		return false, nil
	}
	return false, fmt.Errorf("check whether %s is tracked: %w", relativePath, err)
}

func gitOutput(ctx context.Context, directory string, arguments ...string) ([]byte, error) {
	return commandOutput(ctx, directory, "git", arguments...)
}

func goOutput(ctx context.Context, directory string, arguments ...string) ([]byte, error) {
	return commandOutput(ctx, directory, "go", arguments...)
}

func commandOutput(ctx context.Context, directory, name string, arguments ...string) ([]byte, error) {
	// #nosec G204 -- name is provided only by the fixed gitOutput and goOutput wrappers.
	command := exec.CommandContext(ctx, name, arguments...)
	command.Dir = directory
	output, err := command.Output()
	if err == nil {
		return output, nil
	}
	if exitError, ok := errors.AsType[*exec.ExitError](err); ok {
		message := strings.TrimSpace(string(exitError.Stderr))
		if message != "" {
			return nil, fmt.Errorf("%s: %w", message, err)
		}
	}
	return nil, fmt.Errorf("run %s: %w", name, err)
}

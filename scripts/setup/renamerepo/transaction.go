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
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

type fileSnapshot struct {
	path        string
	contents    []byte
	permissions fs.FileMode
	existed     bool
}

type transaction struct {
	root         string
	snapshots    []fileSnapshot
	path         pathChange
	shouldRename bool
	renamed      bool
}

func executePlan(ctx context.Context, root string, plan *renamePlan, streams streams) error {
	snapshots, err := captureSnapshots(ctx, root, plan.files)
	if err != nil {
		return err
	}
	tx := transaction{
		root:         root,
		snapshots:    snapshots,
		path:         plan.commandDir,
		shouldRename: plan.renameCommand,
		renamed:      false,
	}

	if err := tx.applyFiles(plan.files); err != nil {
		return tx.fail(err)
	}
	if err := tx.renameCommand(); err != nil {
		return tx.fail(err)
	}
	writef(streams.output, "\nRunning go mod tidy...\n")
	if err := runGoModTidy(ctx, root, streams.output, streams.errorOutput); err != nil {
		return tx.fail(err)
	}
	if err := ensureNoNewGoSum(root, snapshots); err != nil {
		return tx.fail(err)
	}
	return nil
}

func captureSnapshots(ctx context.Context, root string, changes []fileChange) ([]fileSnapshot, error) {
	paths := make(map[string]struct{}, len(changes)+2)
	for _, change := range changes {
		paths[change.path] = struct{}{}
	}
	paths["go.mod"] = struct{}{}
	paths["go.sum"] = struct{}{}

	snapshots := make([]fileSnapshot, 0, len(paths))
	for path := range paths {
		snapshot, err := captureSnapshot(ctx, root, path)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}

func captureSnapshot(ctx context.Context, root, relativePath string) (fileSnapshot, error) {
	filename := filepath.Join(root, relativePath)
	info, err := os.Lstat(filename)
	if os.IsNotExist(err) {
		return fileSnapshot{
			path:        relativePath,
			contents:    nil,
			permissions: 0,
			existed:     false,
		}, nil
	}
	if err != nil {
		return fileSnapshot{}, fmt.Errorf("inspect transaction file %s: %w", relativePath, err)
	}
	tracked, err := isTracked(ctx, root, relativePath)
	if err != nil {
		return fileSnapshot{}, err
	}
	if !tracked {
		return fileSnapshot{}, fmt.Errorf("refuse to let go mod tidy modify untracked %s", relativePath)
	}
	if !info.Mode().IsRegular() {
		return fileSnapshot{}, fmt.Errorf("refuse to modify non-regular tracked path %s", relativePath)
	}
	// #nosec G304 -- relativePath is either a Git-tracked candidate or go.mod/go.sum.
	contents, err := os.ReadFile(filename)
	if err != nil {
		return fileSnapshot{}, fmt.Errorf("read transaction file %s: %w", relativePath, err)
	}
	if bytes.IndexByte(contents, 0) >= 0 {
		return fileSnapshot{}, fmt.Errorf("refuse to modify binary tracked file %s", relativePath)
	}
	return fileSnapshot{
		path:        relativePath,
		contents:    contents,
		permissions: info.Mode().Perm(),
		existed:     true,
	}, nil
}

func (tx *transaction) applyFiles(changes []fileChange) error {
	for _, change := range changes {
		if err := os.WriteFile(filepath.Join(tx.root, change.path), change.after, change.permissions); err != nil {
			return fmt.Errorf("update tracked file %s: %w", change.path, err)
		}
	}
	return nil
}

func (tx *transaction) renameCommand() error {
	if !tx.shouldRename {
		return nil
	}
	if err := os.Rename(filepath.Join(tx.root, tx.path.oldPath), filepath.Join(tx.root, tx.path.newPath)); err != nil {
		return fmt.Errorf("rename %s to %s: %w", tx.path.oldPath, tx.path.newPath, err)
	}
	tx.renamed = true
	return nil
}

func runGoModTidy(ctx context.Context, root string, stdout, stderr io.Writer) error {
	command := exec.CommandContext(ctx, "go", "mod", "tidy")
	command.Dir = root
	command.Env = append(os.Environ(), "GOWORK=off")
	command.Stdout = stdout
	command.Stderr = stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}
	return nil
}

func ensureNoNewGoSum(root string, snapshots []fileSnapshot) error {
	for _, snapshot := range snapshots {
		if snapshot.path != "go.sum" || snapshot.existed {
			continue
		}
		if _, err := os.Lstat(filepath.Join(root, snapshot.path)); err == nil {
			return fmt.Errorf("go mod tidy created untracked go.sum")
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect go.sum after go mod tidy: %w", err)
		}
	}
	return nil
}

func (tx *transaction) fail(operationErr error) error {
	rollbackErr := tx.rollback()
	if rollbackErr != nil {
		return fmt.Errorf("repository rename and rollback failed: %w", errors.Join(operationErr, rollbackErr))
	}
	return fmt.Errorf("%w; all repository changes were rolled back", operationErr)
}

func (tx *transaction) rollback() error {
	var rollbackErrors []error
	if tx.renamed {
		if err := os.Rename(filepath.Join(tx.root, tx.path.newPath), filepath.Join(tx.root, tx.path.oldPath)); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("restore command directory: %w", err))
		}
	}
	for _, snapshot := range tx.snapshots {
		if err := restoreSnapshot(tx.root, snapshot); err != nil {
			rollbackErrors = append(rollbackErrors, err)
		}
	}
	return errors.Join(rollbackErrors...)
}

func restoreSnapshot(root string, snapshot fileSnapshot) error {
	filename := filepath.Join(root, snapshot.path)
	if !snapshot.existed {
		if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove newly created %s: %w", snapshot.path, err)
		}
		return nil
	}
	if err := os.WriteFile(filename, snapshot.contents, snapshot.permissions); err != nil {
		return fmt.Errorf("restore %s: %w", snapshot.path, err)
	}
	return nil
}

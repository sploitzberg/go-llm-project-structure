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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type fileChange struct {
	path        string
	before      []byte
	after       []byte
	permissions fs.FileMode
}

type pathChange struct {
	oldPath string
	newPath string
}

type renamePlan struct {
	current       repositoryID
	target        repositoryID
	files         []fileChange
	commandDir    pathChange
	renameCommand bool
}

func buildPlan(ctx context.Context, root string, current, target repositoryID) (renamePlan, error) {
	replacer := newIdentifierReplacer(current, target)
	patterns := []string{current.module, current.owner + "/" + current.name, current.name, current.owner}
	files, err := candidateFiles(ctx, root, patterns)
	if err != nil {
		return renamePlan{}, err
	}

	changes, err := buildFileChanges(root, files, replacer)
	if err != nil {
		return renamePlan{}, err
	}
	commandDir, renameCommand, err := planCommandRename(ctx, root, current.name, target.name)
	if err != nil {
		return renamePlan{}, err
	}
	return renamePlan{
		current:       current,
		target:        target,
		files:         changes,
		commandDir:    commandDir,
		renameCommand: renameCommand,
	}, nil
}

func buildFileChanges(root string, files []string, replacer identifierReplacer) ([]fileChange, error) {
	changes := make([]fileChange, 0, len(files))
	for _, relativePath := range files {
		filename := filepath.Join(root, relativePath)
		info, err := os.Lstat(filename)
		if err != nil {
			return nil, fmt.Errorf("inspect tracked file %s: %w", relativePath, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("refuse to modify non-regular tracked path %s", relativePath)
		}
		// #nosec G304 -- relativePath comes exclusively from git grep's tracked-file output.
		contents, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("read tracked file %s: %w", relativePath, err)
		}
		updated := replacer.replace(contents)
		if bytes.Equal(contents, updated) {
			continue
		}
		changes = append(changes, fileChange{
			path:        relativePath,
			before:      contents,
			after:       updated,
			permissions: info.Mode().Perm(),
		})
	}
	return changes, nil
}

func planCommandRename(ctx context.Context, root, currentName, targetName string) (pathChange, bool, error) {
	noChange := pathChange{oldPath: "", newPath: ""}
	if currentName == targetName {
		return noChange, false, nil
	}
	oldPath := filepath.Join("cmd", currentName)
	newPath := filepath.Join("cmd", targetName)
	oldInfo, err := os.Stat(filepath.Join(root, oldPath))
	if os.IsNotExist(err) {
		return noChange, false, nil
	}
	if err != nil {
		return pathChange{}, false, fmt.Errorf("inspect %s: %w", oldPath, err)
	}
	if !oldInfo.IsDir() {
		return pathChange{}, false, fmt.Errorf("expected %s to be a directory", oldPath)
	}
	if _, err := os.Stat(filepath.Join(root, newPath)); err == nil {
		return pathChange{}, false, fmt.Errorf("refuse to overwrite existing %s", newPath)
	} else if !os.IsNotExist(err) {
		return pathChange{}, false, fmt.Errorf("inspect %s: %w", newPath, err)
	}
	if err := validateTrackedTextTree(ctx, root, oldPath); err != nil {
		return pathChange{}, false, err
	}
	return pathChange{oldPath: oldPath, newPath: newPath}, true, nil
}

func validateTrackedTextTree(ctx context.Context, root, relativeRoot string) error {
	trackedFiles, err := trackedFilesUnder(ctx, root, relativeRoot)
	if err != nil {
		return err
	}
	if len(trackedFiles) == 0 {
		return fmt.Errorf("refuse to rename untracked directory %s", relativeRoot)
	}

	tracked, directories := trackedTreeSets(relativeRoot, trackedFiles)
	tree := trackedTree{root: root, files: tracked, directories: directories}
	if err := filepath.WalkDir(filepath.Join(root, relativeRoot), tree.validateEntry); err != nil {
		return fmt.Errorf("validate tracked command directory: %w", err)
	}
	return nil
}

type trackedTree struct {
	root        string
	files       map[string]struct{}
	directories map[string]struct{}
}

func trackedTreeSets(relativeRoot string, files []string) (tracked, directories map[string]struct{}) {
	tracked = make(map[string]struct{}, len(files))
	directories = map[string]struct{}{filepath.Clean(relativeRoot): {}}
	for _, filename := range files {
		cleaned := filepath.Clean(filename)
		tracked[cleaned] = struct{}{}
		for directory := filepath.Dir(cleaned); directory != "."; directory = filepath.Dir(directory) {
			directories[directory] = struct{}{}
			if directory == filepath.Clean(relativeRoot) {
				break
			}
		}
	}
	return tracked, directories
}

func (tree trackedTree) validateEntry(path string, entry fs.DirEntry, walkErr error) error {
	if walkErr != nil {
		return fmt.Errorf("inspect command directory: %w", walkErr)
	}
	relativePath, err := filepath.Rel(tree.root, path)
	if err != nil {
		return fmt.Errorf("resolve command path: %w", err)
	}
	if entry.IsDir() {
		if _, ok := tree.directories[relativePath]; !ok {
			return fmt.Errorf("refuse to move untracked or ignored directory %s", relativePath)
		}
		return nil
	}
	if _, ok := tree.files[relativePath]; !ok {
		return fmt.Errorf("refuse to move untracked or ignored file %s", relativePath)
	}
	info, err := entry.Info()
	if err != nil {
		return fmt.Errorf("inspect tracked path %s: %w", relativePath, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("refuse to move non-regular tracked path %s", relativePath)
	}
	// #nosec G304 -- path is constrained to the validated repository command tree.
	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read tracked path %s: %w", relativePath, err)
	}
	if bytes.IndexByte(contents, 0) >= 0 {
		return fmt.Errorf("refuse to move binary tracked file %s", relativePath)
	}
	return nil
}

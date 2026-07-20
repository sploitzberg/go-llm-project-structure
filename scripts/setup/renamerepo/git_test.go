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
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func TestCandidateFilesIncludesOnlyTrackedText(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	root := t.TempDir()
	runTestGit(t, root, "init", "--quiet")
	writeTestFile(t, root, ".gitignore", "ignored.txt\n")
	writeTestFile(t, root, "tracked.txt", "github.com/old/repo\n")
	writeTestFile(t, root, "untracked.txt", "github.com/old/repo\n")
	writeTestFile(t, root, "ignored.txt", "github.com/old/repo\n")
	if err := os.WriteFile(filepath.Join(root, "binary.bin"), []byte("github.com/old/repo\x00binary"), 0o600); err != nil {
		t.Fatalf("write binary fixture: %v", err)
	}
	runTestGit(t, root, "add", ".gitignore", "tracked.txt", "binary.bin")

	got, err := candidateFiles(t.Context(), root, []string{"github.com/old/repo"})
	if err != nil {
		t.Fatalf("candidateFiles() error = %v", err)
	}
	want := []string{"tracked.txt"}
	if !slices.Equal(got, want) {
		t.Fatalf("candidateFiles() = %q, want %q", got, want)
	}
}

func runTestGit(t *testing.T, root string, arguments ...string) {
	t.Helper()
	// #nosec G204 -- arguments are fixed test fixture commands, not external input.
	command := exec.CommandContext(t.Context(), "git", arguments...)
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", arguments, err, output)
	}
}

func writeTestFile(t *testing.T, root, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

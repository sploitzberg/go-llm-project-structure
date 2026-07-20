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
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecutePlanRenamesTrackedFilesAndCommandPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	runTestGit(t, root, "init", "--quiet")
	writeTestFile(t, root, "go.mod", "module github.com/old/repo\n\ngo 1.26.5\n")
	writeTestFile(t, root, "README.md", "github.com/old/repo\n")
	writeTestFile(t, root, "untracked.txt", "github.com/old/repo\n")
	if err := os.MkdirAll(filepath.Join(root, "cmd", "repo"), 0o750); err != nil {
		t.Fatalf("create command directory: %v", err)
	}
	writeTestFile(t, root, filepath.Join("cmd", "repo", "main.go"), "package main\n\nfunc main() {}\n")
	runTestGit(t, root, "add", "go.mod", "README.md", "cmd/repo/main.go")

	current := repositoryID{owner: "old", name: "repo", module: "github.com/old/repo"}
	target := repositoryID{owner: "new", name: "service", module: "github.com/new/service"}
	plan, err := buildPlan(t.Context(), root, current, target)
	if err != nil {
		t.Fatalf("buildPlan() error = %v", err)
	}
	quiet := streams{input: strings.NewReader(""), output: io.Discard, errorOutput: io.Discard}
	if err := executePlan(t.Context(), root, &plan, quiet); err != nil {
		t.Fatalf("executePlan() error = %v", err)
	}

	assertTestFile(t, root, "go.mod", "module github.com/new/service\n\ngo 1.26.5\n")
	assertTestFile(t, root, "README.md", "github.com/new/service\n")
	assertTestFile(t, root, "untracked.txt", "github.com/old/repo\n")
	if _, err := os.Stat(filepath.Join(root, "cmd", "service", "main.go")); err != nil {
		t.Fatalf("renamed command path is missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "cmd", "repo")); !os.IsNotExist(err) {
		t.Fatalf("old command path still exists: %v", err)
	}
}

func TestExecutePlanRollsBackFilesAndPathWhenTidyFails(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	runTestGit(t, root, "init", "--quiet")
	goMod := "module github.com/old/repo\n\ngo 1.26.5\n\nreplace example.com/missing => ./does-not-exist\n"
	readme := "github.com/old/repo\n"
	writeTestFile(t, root, "go.mod", goMod)
	writeTestFile(t, root, "README.md", readme)
	if err := os.MkdirAll(filepath.Join(root, "cmd", "repo"), 0o750); err != nil {
		t.Fatalf("create command directory: %v", err)
	}
	writeTestFile(t, root, filepath.Join("cmd", "repo", "main.go"), "package main\n\nimport _ \"example.com/missing\"\n")
	runTestGit(t, root, "add", "go.mod", "README.md", "cmd/repo/main.go")

	current := repositoryID{owner: "old", name: "repo", module: "github.com/old/repo"}
	target := repositoryID{owner: "new", name: "service", module: "github.com/new/service"}
	plan, err := buildPlan(t.Context(), root, current, target)
	if err != nil {
		t.Fatalf("buildPlan() error = %v", err)
	}
	quiet := streams{input: strings.NewReader(""), output: io.Discard, errorOutput: io.Discard}
	err = executePlan(t.Context(), root, &plan, quiet)
	if err == nil || !strings.Contains(err.Error(), "rolled back") {
		t.Fatalf("executePlan() error = %v, want rollback error", err)
	}

	assertTestFile(t, root, "go.mod", goMod)
	assertTestFile(t, root, "README.md", readme)
	if _, err := os.Stat(filepath.Join(root, "cmd", "repo", "main.go")); err != nil {
		t.Fatalf("original command path was not restored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "cmd", "service")); !os.IsNotExist(err) {
		t.Fatalf("new command path still exists after rollback: %v", err)
	}
}

func assertTestFile(t *testing.T, root, name, want string) {
	t.Helper()
	// #nosec G304 -- root and name identify only test-owned temporary fixtures.
	contents, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	if string(contents) != want {
		t.Fatalf("%s = %q, want %q", name, contents, want)
	}
}

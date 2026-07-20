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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testModule = "example.com/project"

//nolint:funlen // Keeping the complete allowed matrix in one fixture makes omissions visible.
func TestRunAcceptsDocumentedDependencyMatrix(t *testing.T) {
	t.Parallel()

	root := writeFixture(t, map[string]string{
		"internal/core/domain/entity.go": `package domain

import "errors"

var ErrInvalid = errors.New("invalid")
`,
		"internal/core/ports/primary/usecase.go": `package primary

import "example.com/project/internal/core/domain"

type UseCase interface { Run(domain.Entity) error }
`,
		"internal/core/ports/secondary/repository.go": `package secondary

import "example.com/project/internal/core/domain"

type Repository interface { Save(domain.Entity) error }
`,
		"internal/core/services/service.go": `package services

import (
	"example.com/project/internal/core/domain"
	"example.com/project/internal/core/ports/primary"
	"example.com/project/internal/core/ports/secondary"
)

var _ primary.UseCase
var _ secondary.Repository
var _ domain.Entity
`,
		"internal/config/config.go": `package config

import "time"

type Config struct { Timeout time.Duration }
`,
		"internal/adapter/primary/http.go": `package primaryadapter

import (
	"github.com/example/router"
	"example.com/project/internal/config"
	"example.com/project/internal/core/domain"
	primaryport "example.com/project/internal/core/ports/primary"
)

var _ router.Router
var _ config.Config
var _ domain.Entity
var _ primaryport.UseCase
`,
		"internal/adapter/secondary/database.go": `package secondaryadapter

import (
	"github.com/example/database"
	"example.com/project/internal/core/domain"
	secondaryport "example.com/project/internal/core/ports/secondary"
)

type Adapter struct{}

var _ database.Client
var _ domain.Entity
var _ secondaryport.Repository = (*Adapter)(nil)
`,
		"internal/core/domain/entity_test.go": `package domain_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	"example.com/project/internal/core/domain"
)

func TestEntity(t *testing.T) { require.NotNil(t, domain.Entity{}) }
`,
	})

	stdout, stderr, exitCode := runFixture(root)
	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0\nstderr:\n%s", exitCode, stderr)
	}
	if !strings.Contains(stdout, "guardrail passed") {
		t.Fatalf("stdout = %q, want success message", stdout)
	}
}

//nolint:funlen // The table intentionally enumerates every forbidden cross-layer edge.
func TestRunRejectsForbiddenDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		filename   string
		source     string
		dependency string
		want       string
	}{
		{
			name:       "domain to config",
			filename:   "internal/core/domain/entity.go",
			source:     "package domain",
			dependency: "internal/config/config.go",
			want:       `core/domain must not import "example.com/project/internal/config"`,
		},
		{
			name:       "primary port to service",
			filename:   "internal/core/ports/primary/usecase.go",
			source:     "package primary",
			dependency: "internal/core/services/service.go",
			want:       `core/ports/primary must not import "example.com/project/internal/core/services"`,
		},
		{
			name:       "secondary port to config",
			filename:   "internal/core/ports/secondary/repository.go",
			source:     "package secondary",
			dependency: "internal/config/config.go",
			want:       `core/ports/secondary must not import "example.com/project/internal/config"`,
		},
		{
			name:       "service to config",
			filename:   "internal/core/services/service.go",
			source:     "package services",
			dependency: "internal/config/config.go",
			want:       `core/services must not import "example.com/project/internal/config"`,
		},
		{
			name:       "primary adapter to secondary port",
			filename:   "internal/adapter/primary/http.go",
			source:     "package primaryadapter",
			dependency: "internal/core/ports/secondary/repository.go",
			want:       `adapter/primary must not import "example.com/project/internal/core/ports/secondary"`,
		},
		{
			name:       "primary adapter to secondary adapter",
			filename:   "internal/adapter/primary/http.go",
			source:     "package primaryadapter",
			dependency: "internal/adapter/secondary/database.go",
			want:       `adapter/primary must not import "example.com/project/internal/adapter/secondary"`,
		},
		{
			name:       "secondary adapter to service",
			filename:   "internal/adapter/secondary/database.go",
			source:     "package secondaryadapter",
			dependency: "internal/core/services/service.go",
			want:       `adapter/secondary must not import "example.com/project/internal/core/services"`,
		},
		{
			name:       "secondary adapter to primary port",
			filename:   "internal/adapter/secondary/database.go",
			source:     "package secondaryadapter",
			dependency: "internal/core/ports/primary/usecase.go",
			want:       `adapter/secondary must not import "example.com/project/internal/core/ports/primary"`,
		},
		{
			name:       "config to domain",
			filename:   "internal/config/config.go",
			source:     "package config",
			dependency: "internal/core/domain/entity.go",
			want:       `config must not import "example.com/project/internal/core/domain"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dependencyPackage := filepath.Base(filepath.Dir(test.dependency))
			root := writeFixture(t, map[string]string{
				test.filename: test.source + `

import _ "` + testModule + `/` + filepath.ToSlash(filepath.Dir(test.dependency)) + `"
`,
				test.dependency: "package " + dependencyPackage + "\n",
			})

			_, stderr, exitCode := runFixture(root)
			if exitCode == 0 {
				t.Fatalf("run() exit code = 0, want failure\nstderr:\n%s", stderr)
			}
			if !strings.Contains(stderr, test.want) {
				t.Fatalf("stderr:\n%s\nwant substring: %s", stderr, test.want)
			}
		})
	}
}

func TestRunRejectsThirdPartyDependenciesInCoreProductionCode(t *testing.T) {
	t.Parallel()

	root := writeFixture(t, map[string]string{
		"internal/core/domain/entity.go": `package domain

import _ "github.com/example/framework"
`,
	})

	_, stderr, exitCode := runFixture(root)
	if exitCode == 0 {
		t.Fatal("run() exit code = 0, want failure")
	}
	if !strings.Contains(stderr, `core/domain must not import third-party package "github.com/example/framework"`) {
		t.Fatalf("stderr:\n%s", stderr)
	}
}

func TestRunRejectsStandardLibraryInfrastructureInCore(t *testing.T) {
	t.Parallel()

	for _, importPath := range []string{"database/sql", "net/http", "net/rpc", "net/smtp"} {
		t.Run(importPath, func(t *testing.T) {
			t.Parallel()
			root := writeFixture(t, map[string]string{
				"internal/core/services/service.go": "package services\n\nimport _ \"" + importPath + "\"\n",
			})

			_, stderr, exitCode := runFixture(root)
			if exitCode == 0 {
				t.Fatal("run() exit code = 0, want failure")
			}
			if !strings.Contains(stderr, "must not import infrastructure package") {
				t.Fatalf("stderr:\n%s", stderr)
			}
		})
	}
}

func TestRunChecksInactiveBuildTaggedFiles(t *testing.T) {
	t.Parallel()

	root := writeFixture(t, map[string]string{
		"internal/core/domain/entity_windows.go": `//go:build windows

package domain

import _ "example.com/project/internal/adapter/secondary"
`,
		"internal/adapter/secondary/database.go": "package secondaryadapter\n",
	})

	_, stderr, exitCode := runFixture(root)
	if exitCode == 0 {
		t.Fatal("run() exit code = 0, want failure")
	}
	if !strings.Contains(stderr, "core/domain must not import") {
		t.Fatalf("stderr:\n%s", stderr)
	}
}

func TestRunRejectsSyntaxErrorsAndMissingProjectPackages(t *testing.T) {
	t.Parallel()

	root := writeFixture(t, map[string]string{
		"internal/core/domain/broken_linux.go": `//go:build linux

package domain

func broken(
`,
		"internal/core/services/service_windows.go": `//go:build windows

package services

import _ "example.com/project/internal/core/domain/missing"
`,
	})

	_, stderr, exitCode := runFixture(root)
	if exitCode == 0 {
		t.Fatal("run() exit code = 0, want failure")
	}
	for _, want := range []string{"parse Go source", "imports missing project package"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr missing %q:\n%s", want, stderr)
		}
	}
}

func TestRunRequiresSecondaryAdapterPortAssertions(t *testing.T) {
	t.Parallel()

	root := writeFixture(t, map[string]string{
		"internal/adapter/secondary/database.go": "package secondaryadapter\n\ntype Adapter struct{}\n",
	})

	_, stderr, exitCode := runFixture(root)
	if exitCode == 0 {
		t.Fatal("run() exit code = 0, want failure")
	}
	if !strings.Contains(stderr, "must declare a compile-time port assertion") {
		t.Fatalf("stderr:\n%s", stderr)
	}
}

func TestRunRejectsUndefinedInternalLayers(t *testing.T) {
	t.Parallel()

	root := writeFixture(t, map[string]string{
		"internal/helpers/helper.go": "package helpers\n",
	})

	_, stderr, exitCode := runFixture(root)
	if exitCode == 0 {
		t.Fatal("run() exit code = 0, want failure")
	}
	if !strings.Contains(stderr, "outside the supported internal architecture layers") {
		t.Fatalf("stderr:\n%s", stderr)
	}
}

func TestRunRejectsImportCycles(t *testing.T) {
	t.Parallel()

	root := writeFixture(t, map[string]string{
		"internal/one/one.go": `package one

import _ "example.com/project/internal/two"
`,
		"internal/two/two.go": `package two

import _ "example.com/project/internal/one"
`,
	})

	_, stderr, exitCode := runFixture(root)
	if exitCode == 0 {
		t.Fatal("run() exit code = 0, want failure")
	}
	if !strings.Contains(stderr, "import cycle:") {
		t.Fatalf("stderr:\n%s", stderr)
	}
}

func writeFixture(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	files["go.mod"] = "module " + testModule + "\n\ngo 1.26.5\n"
	for filename, content := range files {
		fullPath := filepath.Join(root, filepath.FromSlash(filename))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
			t.Fatalf("MkdirAll(%q): %v", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile(%q): %v", fullPath, err)
		}
	}
	return root
}

func runFixture(root string) (stdoutText, stderrText string, exitCode int) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode = run(root, &stdout, &stderr)
	return stdout.String(), stderr.String(), exitCode
}

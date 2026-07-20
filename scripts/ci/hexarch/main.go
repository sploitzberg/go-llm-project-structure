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

// Command hexarch validates the repository's hexagonal dependency rules.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type layer string

const (
	layerUnknown          layer = "unknown"
	layerDomain           layer = "core/domain"
	layerPorts            layer = "core/ports"
	layerPortsPrimary     layer = "core/ports/primary"
	layerPortsSecondary   layer = "core/ports/secondary"
	layerServices         layer = "core/services"
	layerAdapterPrimary   layer = "adapter/primary"
	layerAdapterSecondary layer = "adapter/secondary"
	layerConfig           layer = "config"
)

type sourceFile struct {
	filename         string
	importPath       string
	layer            layer
	isTest           bool
	hasPortAssertion bool
	imports          []fileImport
}

type fileImport struct {
	path string
	line int
}

type violation struct {
	filename string
	line     int
	message  string
}

func main() {
	os.Exit(run(".", os.Stdout, os.Stderr))
}

//nolint:gocognit // Orchestration keeps all validation failures in one deterministic report.
func run(root string, stdout, stderr io.Writer) int {
	modulePath, err := readModulePath(filepath.Join(root, "go.mod"))
	if err != nil {
		writef(stderr, "error: %v\n", err)
		return 1
	}

	files, parseViolations, err := parseSourceFiles(root, modulePath)
	if err != nil {
		writef(stderr, "error: inspect Go source: %v\n", err)
		return 1
	}

	violations := slices.Clone(parseViolations)
	violations = append(violations, validateImports(files, modulePath)...)
	violations = append(violations, validateAdapterAssertions(files)...)
	violations = append(violations, detectCycles(files, modulePath)...)
	slices.SortFunc(violations, func(a, b violation) int {
		if result := strings.Compare(a.filename, b.filename); result != 0 {
			return result
		}
		if result := a.line - b.line; result != 0 {
			return result
		}
		return strings.Compare(a.message, b.message)
	})

	if len(violations) > 0 {
		for _, item := range violations {
			if item.line > 0 {
				writef(stderr, "error: %s:%d: %s\n", item.filename, item.line, item.message)
				continue
			}
			writef(stderr, "error: %s: %s\n", item.filename, item.message)
		}
		writef(stderr, "error: hexagonal architecture guardrail failed with %d violation(s)\n", len(violations))
		return 1
	}

	writef(stdout, "success: hexagonal architecture guardrail passed (%d Go files checked)\n", len(files))
	return 0
}

func writef(writer io.Writer, format string, values ...any) {
	_, _ = fmt.Fprintf(writer, format, values...)
}

func readModulePath(filename string) (string, error) {
	// #nosec G304 -- filename is the repository go.mod selected by this tool, or a test fixture.
	contents, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}

	for rawLine := range strings.SplitSeq(string(contents), "\n") {
		line := strings.TrimSpace(rawLine)
		if modulePath, ok := strings.CutPrefix(line, "module "); ok {
			modulePath = strings.TrimSpace(modulePath)
			if modulePath == "" {
				return "", fmt.Errorf("go.mod has an empty module directive")
			}
			return modulePath, nil
		}
	}
	return "", fmt.Errorf("go.mod has no module directive")
}

//nolint:gocognit,funlen // Walking and parsing all build-tag variants requires explicit failure branches.
func parseSourceFiles(root, modulePath string) ([]sourceFile, []violation, error) {
	fileSet := token.NewFileSet()
	var files []sourceFile
	var violations []violation

	err := filepath.WalkDir(root, func(filename string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if filename != root && ignoredDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(filename) != ".go" {
			return nil
		}

		relativeName, err := filepath.Rel(root, filename)
		if err != nil {
			return fmt.Errorf("make %q relative to repository root: %w", filename, err)
		}
		relativeName = filepath.ToSlash(relativeName)
		parsed, err := parser.ParseFile(fileSet, filename, nil, 0)
		if err != nil {
			violations = append(violations, violation{
				filename: relativeName,
				line:     0,
				message:  fmt.Sprintf("parse Go source: %v", err),
			})
			return nil
		}

		file := sourceFile{
			filename:         relativeName,
			importPath:       packageImportPath(modulePath, filepath.Dir(relativeName)),
			layer:            classifyLayer(relativeName),
			isTest:           strings.HasSuffix(relativeName, "_test.go"),
			hasPortAssertion: hasCompileTimeAssertion(parsed),
			imports:          nil,
		}
		if file.layer == layerUnknown && hasPathPrefix(relativeName, "internal") {
			violations = append(violations, violation{
				filename: relativeName,
				line:     0,
				message:  "Go source is outside the supported internal architecture layers",
			})
		}
		for _, spec := range parsed.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				position := fileSet.Position(spec.Pos())
				violations = append(violations, violation{
					filename: relativeName,
					line:     position.Line,
					message:  fmt.Sprintf("invalid import path %s", spec.Path.Value),
				})
				continue
			}
			file.imports = append(file.imports, fileImport{
				path: importPath,
				line: fileSet.Position(spec.Pos()).Line,
			})
		}
		files = append(files, file)
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("walk repository: %w", err)
	}
	return files, violations, nil
}

//nolint:gocognit // AST node filtering is clearer as direct guards than as type-specific helpers.
func hasCompileTimeAssertion(file *ast.File) bool {
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.VAR {
			continue
		}
		for _, spec := range general.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok || value.Type == nil || len(value.Values) == 0 {
				continue
			}
			for _, name := range value.Names {
				if name.Name == "_" {
					return true
				}
			}
		}
	}
	return false
}

func ignoredDirectory(name string) bool {
	return name == ".git" || name == "bin" || name == "dist" || name == "vendor"
}

func packageImportPath(modulePath, relativeDirectory string) string {
	relativeDirectory = filepath.ToSlash(relativeDirectory)
	if relativeDirectory == "." {
		return modulePath
	}
	return modulePath + "/" + relativeDirectory
}

func classifyLayer(filename string) layer {
	directory := filepath.ToSlash(filepath.Dir(filename))
	switch {
	case hasPathPrefix(directory, "internal/core/domain"):
		return layerDomain
	case hasPathPrefix(directory, "internal/core/ports/primary"):
		return layerPortsPrimary
	case hasPathPrefix(directory, "internal/core/ports/secondary"):
		return layerPortsSecondary
	case hasPathPrefix(directory, "internal/core/ports"):
		return layerPorts
	case hasPathPrefix(directory, "internal/core/services"):
		return layerServices
	case hasPathPrefix(directory, "internal/adapter/primary"):
		return layerAdapterPrimary
	case hasPathPrefix(directory, "internal/adapter/secondary"):
		return layerAdapterSecondary
	case hasPathPrefix(directory, "internal/config"):
		return layerConfig
	default:
		return layerUnknown
	}
}

//nolint:gocognit // The branches directly encode the documented dependency matrix and test exceptions.
func validateImports(files []sourceFile, modulePath string) []violation {
	packages := make(map[string]bool)
	for _, file := range files {
		if !file.isTest {
			packages[file.importPath] = true
		}
	}

	var violations []violation
	for _, file := range files {
		if file.layer == layerUnknown {
			continue
		}
		for _, item := range file.imports {
			if file.isTest && item.path == file.importPath {
				continue
			}
			if relativeImport, local := projectImport(item.path, modulePath); local {
				if !packages[item.path] {
					violations = append(violations, violation{
						filename: file.filename,
						line:     item.line,
						message:  fmt.Sprintf("%s imports missing project package %q", file.layer, item.path),
					})
					continue
				}
				if !allowedProjectImport(file.layer, relativeImport) {
					violations = append(violations, violation{
						filename: file.filename,
						line:     item.line,
						message:  fmt.Sprintf("%s must not import %q", file.layer, item.path),
					})
				}
				continue
			}
			if file.isTest || allowsExternalDependencies(file.layer) {
				continue
			}
			if isInfrastructurePackage(item.path) {
				violations = append(violations, violation{
					filename: file.filename,
					line:     item.line,
					message:  fmt.Sprintf("%s must not import infrastructure package %q", file.layer, item.path),
				})
				continue
			}
			if !isStandardLibrary(item.path) {
				violations = append(violations, violation{
					filename: file.filename,
					line:     item.line,
					message:  fmt.Sprintf("%s must not import third-party package %q", file.layer, item.path),
				})
			}
		}
	}
	return violations
}

func validateAdapterAssertions(files []sourceFile) []violation {
	type packageState struct {
		filename  string
		assertion bool
	}
	packages := make(map[string]packageState)
	for _, file := range files {
		if file.layer != layerAdapterSecondary || file.isTest {
			continue
		}
		state := packages[file.importPath]
		if state.filename == "" {
			state.filename = file.filename
		}
		state.assertion = state.assertion || file.hasPortAssertion
		packages[file.importPath] = state
	}

	var violations []violation
	for importPath, state := range packages {
		if !state.assertion {
			violations = append(violations, violation{
				filename: state.filename,
				line:     0,
				message:  fmt.Sprintf("secondary adapter package %q must declare a compile-time port assertion", importPath),
			})
		}
	}
	return violations
}

func projectImport(importPath, modulePath string) (string, bool) {
	if importPath == modulePath {
		return ".", true
	}
	relativeImport, ok := strings.CutPrefix(importPath, modulePath+"/")
	return relativeImport, ok
}

func allowedProjectImport(source layer, target string) bool {
	switch source {
	case layerPorts, layerPortsPrimary, layerPortsSecondary:
		return hasPathPrefix(target, "internal/core/domain")
	case layerServices:
		return hasPathPrefix(target, "internal/core/domain") || hasPathPrefix(target, "internal/core/ports")
	case layerAdapterPrimary:
		return hasPathPrefix(target, "internal/core/domain") ||
			hasPathPrefix(target, "internal/core/ports/primary") ||
			hasPathPrefix(target, "internal/config")
	case layerAdapterSecondary:
		return hasPathPrefix(target, "internal/core/domain") || hasPathPrefix(target, "internal/core/ports/secondary")
	case layerDomain, layerConfig, layerUnknown:
		return false
	default:
		return false
	}
}

func allowsExternalDependencies(source layer) bool {
	return source == layerAdapterPrimary || source == layerAdapterSecondary
}

func isInfrastructurePackage(importPath string) bool {
	return importPath == "database/sql" || importPath == "net/http" || importPath == "net/rpc" || importPath == "net/smtp"
}

func isStandardLibrary(importPath string) bool {
	if importPath == "C" || strings.HasPrefix(importPath, ".") {
		return false
	}
	firstSegment, _, _ := strings.Cut(importPath, "/")
	return !strings.Contains(firstSegment, ".")
}

func hasPathPrefix(path, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

//nolint:gocognit,funlen // Depth-first cycle detection necessarily tracks graph, stack, and visit state.
func detectCycles(files []sourceFile, modulePath string) []violation {
	graph := make(map[string][]string)
	for _, file := range files {
		if file.isTest {
			continue
		}
		if _, ok := graph[file.importPath]; !ok {
			graph[file.importPath] = nil
		}
		for _, item := range file.imports {
			if _, local := projectImport(item.path, modulePath); local && item.path != file.importPath {
				graph[file.importPath] = append(graph[file.importPath], item.path)
			}
		}
	}

	for packagePath := range graph {
		slices.Sort(graph[packagePath])
		graph[packagePath] = slices.Compact(graph[packagePath])
	}

	const (
		unvisited = iota
		visiting
		visited
	)
	state := make(map[string]int)
	var stack []string
	var violations []violation
	var visit func(string)
	visit = func(packagePath string) {
		state[packagePath] = visiting
		stack = append(stack, packagePath)
		for _, dependency := range graph[packagePath] {
			if _, exists := graph[dependency]; !exists {
				continue
			}
			switch state[dependency] {
			case unvisited:
				visit(dependency)
			case visiting:
				start := slices.Index(stack, dependency)
				cycle := append(slices.Clone(stack[start:]), dependency)
				violations = append(violations, violation{
					filename: packagePath,
					line:     0,
					message:  "import cycle: " + strings.Join(cycle, " -> "),
				})
			}
		}
		stack = stack[:len(stack)-1]
		state[packagePath] = visited
	}

	packages := make([]string, 0, len(graph))
	for packagePath := range graph {
		packages = append(packages, packagePath)
	}
	slices.Sort(packages)
	for _, packagePath := range packages {
		if state[packagePath] == unvisited {
			visit(packagePath)
		}
	}
	return violations
}

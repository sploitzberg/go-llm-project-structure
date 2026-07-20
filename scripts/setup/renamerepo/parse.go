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
	"fmt"
	"strings"
)

const (
	githubHost     = "github.com"
	githubPathPart = 3
	maxOwnerLength = 39
	maxRepoLength  = 100
)

type repositoryID struct {
	owner  string
	name   string
	module string
}

func parseTarget(input string) (repositoryID, error) {
	cleaned := strings.TrimSpace(input)
	cleaned = strings.TrimPrefix(cleaned, "https://")
	cleaned, _ = strings.CutSuffix(cleaned, ".git")

	return parseGitHubPath(cleaned)
}

func parseCurrentModule(module string) (repositoryID, error) {
	if strings.TrimSpace(module) != module {
		return repositoryID{}, fmt.Errorf("current module path contains surrounding whitespace")
	}

	current, err := parseGitHubPath(module)
	if err != nil {
		return repositoryID{}, fmt.Errorf("current module must be github.com/<owner>/<repo>: %w", err)
	}
	return current, nil
}

func parseGitHubPath(value string) (repositoryID, error) {
	parts := strings.Split(value, "/")
	if len(parts) != githubPathPart || parts[0] != githubHost {
		return repositoryID{}, fmt.Errorf("expected github.com/<owner>/<repo>")
	}
	if err := validateOwner(parts[1]); err != nil {
		return repositoryID{}, err
	}
	if err := validateRepo(parts[2]); err != nil {
		return repositoryID{}, err
	}

	return repositoryID{
		owner:  parts[1],
		name:   parts[2],
		module: strings.Join(parts, "/"),
	}, nil
}

func validateOwner(owner string) error {
	if owner == "" {
		return fmt.Errorf("owner must not be empty")
	}
	if len(owner) > maxOwnerLength {
		return fmt.Errorf("owner must be at most %d characters", maxOwnerLength)
	}
	if !isASCIIAlphanumeric(owner[0]) || !isASCIIAlphanumeric(owner[len(owner)-1]) {
		return fmt.Errorf("owner must start and end with an ASCII letter or digit")
	}
	if strings.Contains(owner, "--") {
		return fmt.Errorf("owner must not contain consecutive hyphens")
	}
	for index := range len(owner) {
		if !isASCIIAlphanumeric(owner[index]) && owner[index] != '-' {
			return fmt.Errorf("owner may contain only ASCII letters, digits, and hyphens")
		}
	}
	return nil
}

func validateRepo(repo string) error {
	if repo == "" {
		return fmt.Errorf("repository name must not be empty")
	}
	if len(repo) > maxRepoLength {
		return fmt.Errorf("repository name must be at most %d characters", maxRepoLength)
	}
	if repo == "." || repo == ".." {
		return fmt.Errorf("repository name must not be a relative path component")
	}
	for index := range len(repo) {
		if !isSafeRepoByte(repo[index]) {
			return fmt.Errorf("repository name may contain only ASCII letters, digits, periods, hyphens, and underscores")
		}
	}
	return nil
}

func isSafeRepoByte(value byte) bool {
	return isASCIIAlphanumeric(value) || value == '.' || value == '-' || value == '_'
}

func isASCIIAlphanumeric(value byte) bool {
	return value >= 'a' && value <= 'z' ||
		value >= 'A' && value <= 'Z' ||
		value >= '0' && value <= '9'
}

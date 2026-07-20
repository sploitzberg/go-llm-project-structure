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
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

type options struct {
	yes    bool
	target string
}

type streams struct {
	input       io.Reader
	output      io.Writer
	errorOutput io.Writer
}

func run(ctx context.Context, arguments []string, streams streams) int {
	options, err := parseArguments(arguments, streams.errorOutput)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		writef(streams.errorOutput, "error: %v\n", err)
		return 2
	}

	reader := bufio.NewReader(streams.input)
	if err := performRename(ctx, options, reader, streams); err != nil {
		writef(streams.errorOutput, "error: %v\n", err)
		return 1
	}
	return 0
}

func parseArguments(arguments []string, stderr io.Writer) (options, error) {
	var parsed options
	flags := flag.NewFlagSet("rename-repo", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.BoolVar(&parsed.yes, "yes", false, "apply the rename without confirmation")
	flags.Usage = func() {
		writef(stderr, "Usage: ./scripts/setup/rename-repo.sh [--yes] [github.com/<owner>/<repo>]\n")
		writef(stderr, "       ./scripts/setup/rename-repo.sh [--yes] [https://github.com/<owner>/<repo>.git]\n")
	}
	if err := flags.Parse(arguments); err != nil {
		return options{}, fmt.Errorf("parse arguments: %w", err)
	}
	if flags.NArg() > 1 {
		return options{}, fmt.Errorf("expected at most one repository target")
	}
	if flags.NArg() == 1 {
		parsed.target = flags.Arg(0)
	}
	if parsed.yes && parsed.target == "" {
		return options{}, fmt.Errorf("--yes requires a repository target")
	}
	return parsed, nil
}

func performRename(ctx context.Context, options options, reader *bufio.Reader, streams streams) error {
	root, current, err := loadCurrentRepository(ctx)
	if err != nil {
		return err
	}
	target, err := readTarget(options, reader, streams.output, current)
	if err != nil {
		return err
	}
	plan, err := buildPlan(ctx, root, current, target)
	if err != nil {
		return err
	}
	printPreview(streams.output, &plan)
	confirmed, err := confirm(options.yes, reader, streams.output)
	if err != nil {
		return err
	}
	if !confirmed {
		writef(streams.output, "Aborted; no files changed.\n")
		return nil
	}
	if err := ensureTrackedWorktreeClean(ctx, root); err != nil {
		return err
	}
	if err := executePlan(ctx, root, &plan, streams); err != nil {
		return err
	}
	printSuccess(streams.output, &plan)
	return nil
}

func loadCurrentRepository(ctx context.Context) (string, repositoryID, error) {
	root, err := gitTopLevel(ctx, ".")
	if err != nil {
		return "", repositoryID{}, err
	}
	module, err := goModulePath(ctx, root)
	if err != nil {
		return "", repositoryID{}, err
	}
	current, err := parseCurrentModule(module)
	if err != nil {
		return "", repositoryID{}, err
	}
	if err := ensureTrackedWorktreeClean(ctx, root); err != nil {
		return "", repositoryID{}, err
	}
	return root, current, nil
}

func readTarget(options options, reader *bufio.Reader, stdout io.Writer, current repositoryID) (repositoryID, error) {
	input, err := targetInput(options, reader, stdout)
	if err != nil {
		return repositoryID{}, err
	}
	target, err := parseTarget(input)
	if err != nil {
		return repositoryID{}, fmt.Errorf("invalid repository target: %w", err)
	}
	if target.module == current.module {
		return repositoryID{}, fmt.Errorf("new module path is identical to the current module path")
	}
	return target, nil
}

func targetInput(options options, reader *bufio.Reader, stdout io.Writer) (string, error) {
	if options.target != "" {
		return options.target, nil
	}
	writef(stdout, "Repository target (github.com/<owner>/<repo>): ")
	value, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read repository target: %w", err)
	}
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("repository target must not be empty")
	}
	return value, nil
}

func confirm(assumeYes bool, reader *bufio.Reader, stdout io.Writer) (bool, error) {
	if assumeYes {
		return true, nil
	}
	writef(stdout, "Proceed? [y/N]: ")
	answer, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

func printPreview(writer io.Writer, plan *renamePlan) {
	writef(writer, "\nRepository rename plan\n")
	writef(writer, "  Module:       %s -> %s\n", plan.current.module, plan.target.module)
	writef(writer, "  Clone URL:    https://%s/%s/%s.git\n", githubHost, plan.target.owner, plan.target.name)
	writef(writer, "  Text files:   %d tracked file(s)\n", len(plan.files))
	if plan.renameCommand {
		writef(writer, "  Command path: %s -> %s\n", plan.commandDir.oldPath, plan.commandDir.newPath)
	} else {
		writef(writer, "  Command path: unchanged\n")
	}
}

func printSuccess(writer io.Writer, plan *renamePlan) {
	writef(writer, "\nRepository renamed successfully.\n")
	writef(writer, "  Module: %s\n", plan.target.module)
	writef(writer, "  Updated %d tracked text file(s):\n", len(plan.files))
	for _, change := range plan.files {
		writef(writer, "    - %s\n", change.path)
	}
	if plan.renameCommand {
		writef(writer, "  Renamed %s -> %s\n", plan.commandDir.oldPath, plan.commandDir.newPath)
	}
	writef(writer, "  go mod tidy completed successfully.\n")
	writef(writer, "Review the result with: git diff --find-renames\n")
}

func writef(writer io.Writer, format string, values ...any) {
	_, _ = fmt.Fprintf(writer, format, values...)
}

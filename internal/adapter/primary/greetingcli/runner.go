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

// Package greetingcli exposes the removable greeting use case through a CLI.
package greetingcli

import (
	"context"
	"fmt"
	"io"
	"strings"

	primarygreeting "github.com/sploitzberg/go-llm-project-structure/internal/core/ports/primary/greeting"
)

const defaultName = "World"

// Runner translates CLI arguments into greeting commands.
type Runner struct {
	greeter primarygreeting.Greeter
}

// New creates a CLI runner.
func New(greeter primarygreeting.Greeter) *Runner {
	return &Runner{greeter: greeter}
}

// Run creates a greeting and writes it to output.
func (r *Runner) Run(ctx context.Context, arguments []string, output io.Writer) error {
	name := defaultName
	if len(arguments) > 0 {
		name = strings.Join(arguments, " ")
	}
	result, err := r.greeter.Greet(ctx, primarygreeting.Command{Name: name})
	if err != nil {
		return fmt.Errorf("greet: %w", err)
	}
	if _, err := fmt.Fprintln(output, result.String()); err != nil {
		return fmt.Errorf("write greeting: %w", err)
	}
	return nil
}

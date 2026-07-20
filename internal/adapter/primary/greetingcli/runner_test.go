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

package greetingcli

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	domaingreeting "github.com/sploitzberg/go-llm-project-structure/internal/core/domain/greeting"
	primarygreeting "github.com/sploitzberg/go-llm-project-structure/internal/core/ports/primary/greeting"
)

type greeterStub struct {
	command primarygreeting.Command
}

func (g *greeterStub) Greet(_ context.Context, command primarygreeting.Command) (domaingreeting.Greeting, error) {
	g.command = command
	result, err := domaingreeting.New(command.Name)
	if err != nil {
		return domaingreeting.Greeting{}, fmt.Errorf("create greeting: %w", err)
	}
	return result, nil
}

func TestRunnerRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		arguments  []string
		wantName   string
		wantOutput string
	}{
		{name: "default", arguments: nil, wantName: "World", wantOutput: "Hello, World!\n"},
		{name: "arguments", arguments: []string{"Go", "Developer"}, wantName: "Go Developer", wantOutput: "Hello, Go Developer!\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			greeter := &greeterStub{command: primarygreeting.Command{Name: ""}}
			var output bytes.Buffer
			if err := New(greeter).Run(t.Context(), test.arguments, &output); err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if greeter.command.Name != test.wantName {
				t.Fatalf("command name = %q, want %q", greeter.command.Name, test.wantName)
			}
			if output.String() != test.wantOutput {
				t.Fatalf("output = %q, want %q", output.String(), test.wantOutput)
			}
		})
	}
}

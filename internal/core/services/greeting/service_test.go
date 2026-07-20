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

package greeting

import (
	"context"
	"errors"
	"testing"

	domaingreeting "github.com/sploitzberg/go-llm-project-structure/internal/core/domain/greeting"
	primarygreeting "github.com/sploitzberg/go-llm-project-structure/internal/core/ports/primary/greeting"
)

var errRecord = errors.New("record failed")

type recorderStub struct {
	recorded domaingreeting.Greeting
	err      error
}

func (r *recorderStub) Record(_ context.Context, value domaingreeting.Greeting) error {
	r.recorded = value
	return r.err
}

func TestServiceGreet(t *testing.T) {
	t.Parallel()

	recorder := &recorderStub{recorded: domaingreeting.Greeting{}, err: nil}
	service := New(recorder)

	result, err := service.Greet(t.Context(), primarygreeting.Command{Name: "Gopher"})
	if err != nil {
		t.Fatalf("Greet() error = %v", err)
	}
	if result.String() != "Hello, Gopher!" {
		t.Fatalf("Greet().String() = %q", result.String())
	}
	if recorder.recorded.String() != result.String() {
		t.Fatalf("recorded greeting = %q, want %q", recorder.recorded.String(), result.String())
	}
}

func TestServiceGreetErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		command   primarygreeting.Command
		recordErr error
		wantError error
	}{
		{name: "invalid name", command: primarygreeting.Command{Name: ""}, recordErr: nil, wantError: domaingreeting.ErrNameRequired},
		{name: "recorder", command: primarygreeting.Command{Name: "Gopher"}, recordErr: errRecord, wantError: errRecord},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := &recorderStub{recorded: domaingreeting.Greeting{}, err: test.recordErr}
			_, err := New(recorder).Greet(t.Context(), test.command)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("Greet() error = %v, want %v", err, test.wantError)
			}
		})
	}
}

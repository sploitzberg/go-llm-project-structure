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

package greetingmemory

import (
	"context"
	"errors"
	"testing"

	domaingreeting "github.com/sploitzberg/go-llm-project-structure/internal/core/domain/greeting"
)

func TestRecorder(t *testing.T) {
	t.Parallel()

	value, err := domaingreeting.New("Gopher")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	recorder := NewRecorder()
	if err := recorder.Record(t.Context(), value); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	got := recorder.Greetings()
	if len(got) != 1 || got[0].String() != value.String() {
		t.Fatalf("Greetings() = %v, want %q", got, value.String())
	}
	got[0] = domaingreeting.Greeting{}
	if recorder.Greetings()[0].String() != value.String() {
		t.Fatal("Greetings() exposed recorder storage")
	}
}

func TestRecorderHonorsCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	value, err := domaingreeting.New("Gopher")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = NewRecorder().Record(ctx, value)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Record() error = %v, want %v", err, context.Canceled)
	}
}

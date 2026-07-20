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

// Package greetingmemory records greetings in memory for the removable example.
package greetingmemory

import (
	"context"
	"fmt"
	"slices"
	"sync"

	domaingreeting "github.com/sploitzberg/go-llm-project-structure/internal/core/domain/greeting"
	secondarygreeting "github.com/sploitzberg/go-llm-project-structure/internal/core/ports/secondary/greeting"
)

// Recorder stores greetings in memory.
type Recorder struct {
	mu        sync.RWMutex
	greetings []domaingreeting.Greeting
}

// NewRecorder creates an empty recorder.
func NewRecorder() *Recorder {
	return &Recorder{mu: sync.RWMutex{}, greetings: nil}
}

// Record stores a greeting.
func (r *Recorder) Record(ctx context.Context, value domaingreeting.Greeting) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("record greeting: %w", err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.greetings = append(r.greetings, value)
	return nil
}

// Greetings returns a snapshot of recorded greetings.
func (r *Recorder) Greetings() []domaingreeting.Greeting {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Clone(r.greetings)
}

var _ secondarygreeting.Recorder = (*Recorder)(nil)

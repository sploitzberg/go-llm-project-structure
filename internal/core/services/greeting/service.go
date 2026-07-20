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

// Package greeting implements the removable greeting use case.
package greeting

import (
	"context"
	"fmt"

	domaingreeting "github.com/sploitzberg/go-llm-project-structure/internal/core/domain/greeting"
	primarygreeting "github.com/sploitzberg/go-llm-project-structure/internal/core/ports/primary/greeting"
	secondarygreeting "github.com/sploitzberg/go-llm-project-structure/internal/core/ports/secondary/greeting"
)

// Service creates and records greetings.
type Service struct {
	recorder secondarygreeting.Recorder
}

// New creates a greeting service.
func New(recorder secondarygreeting.Recorder) *Service {
	return &Service{recorder: recorder}
}

// Greet creates and records a greeting.
func (s *Service) Greet(ctx context.Context, command primarygreeting.Command) (domaingreeting.Greeting, error) {
	result, err := domaingreeting.New(command.Name)
	if err != nil {
		return domaingreeting.Greeting{}, fmt.Errorf("create greeting: %w", err)
	}
	if err := s.recorder.Record(ctx, result); err != nil {
		return domaingreeting.Greeting{}, fmt.Errorf("record greeting: %w", err)
	}
	return result, nil
}

var _ primarygreeting.Greeter = (*Service)(nil)

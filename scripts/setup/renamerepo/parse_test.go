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

import "testing"

func TestParseTargetAcceptsSupportedForms(t *testing.T) {
	t.Parallel()

	tests := []string{
		"github.com/acme/my-service",
		"https://github.com/acme/my-service",
		"https://github.com/acme/my-service.git",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			got, err := parseTarget(input)
			if err != nil {
				t.Fatalf("parseTarget(%q) error = %v", input, err)
			}
			if got.module != "github.com/acme/my-service" {
				t.Fatalf("parseTarget(%q).module = %q, want %q", input, got.module, "github.com/acme/my-service")
			}
		})
	}
}

func TestParseTargetRejectsUnsafeValues(t *testing.T) {
	t.Parallel()

	tests := []string{
		"",
		"github.com//repo",
		"github.com/owner/",
		"github.com/-owner/repo",
		"github.com/owner--name/repo",
		"github.com/owner/..",
		"github.com/owner/repo/name",
		"github.com/owner/repo?tab=1",
		"http://github.com/owner/repo",
		"git@github.com:owner/repo",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			if _, err := parseTarget(input); err == nil {
				t.Fatalf("parseTarget(%q) error = nil, want error", input)
			}
		})
	}
}

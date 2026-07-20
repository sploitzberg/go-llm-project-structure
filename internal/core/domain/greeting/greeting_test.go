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
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		want      string
		wantError error
	}{
		{name: "name", input: "Gopher", want: "Hello, Gopher!", wantError: nil},
		{name: "trim spaces", input: "  Gopher  ", want: "Hello, Gopher!", wantError: nil},
		{name: "missing name", input: "  ", want: "", wantError: ErrNameRequired},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result, err := New(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("New(%q) error = %v, want %v", test.input, err, test.wantError)
			}
			if result.String() != test.want {
				t.Fatalf("New(%q).String() = %q, want %q", test.input, result.String(), test.want)
			}
		})
	}
}

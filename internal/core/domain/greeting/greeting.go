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

// Package greeting contains the removable example's domain behavior.
package greeting

import (
	"errors"
	"strings"
)

// ErrNameRequired indicates that a greeting cannot be created without a name.
var ErrNameRequired = errors.New("name is required")

// Greeting is an immutable greeting message.
type Greeting struct {
	message string
}

// New creates a greeting for name.
func New(name string) (Greeting, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Greeting{}, ErrNameRequired
	}
	return Greeting{message: "Hello, " + name + "!"}, nil
}

// String returns the greeting text.
func (g Greeting) String() string {
	return g.message
}

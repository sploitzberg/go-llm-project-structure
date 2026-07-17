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
	"fmt"
	"net/http"
	_ "net/http/pprof" // #nosec G108 -- pprof is intentionally exposed for profiling (only with ENABLE_PPROF=true)
	"os"
)

func main() {
	// Enable pprof profiling if ENABLE_PPROF is set
	if os.Getenv("ENABLE_PPROF") == "true" {
		go func() {
			fmt.Println("Profiling enabled on http://localhost:6060")
			// #nosec G114 -- pprof server without timeouts is acceptable for dev profiling
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				fmt.Printf("pprof server error: %v\n", err)
			}
		}()
	}

	fmt.Println("Hello, World!")
}

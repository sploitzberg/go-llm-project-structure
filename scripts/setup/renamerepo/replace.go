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

import "strings"

type identifierReplacer struct {
	standard  *strings.Replacer
	copyright *strings.Replacer
}

func newIdentifierReplacer(current, target repositoryID) identifierReplacer {
	pairs := replacementPairs(current, target)
	copyrightPairs := append([]string{}, pairs...)
	copyrightPairs = append(copyrightPairs, " "+current.owner, " "+target.owner)

	return identifierReplacer{
		standard:  strings.NewReplacer(pairs...),
		copyright: strings.NewReplacer(copyrightPairs...),
	}
}

func replacementPairs(current, target repositoryID) []string {
	return []string{
		current.module, target.module,
		current.owner + "/" + current.name, target.owner + "/" + target.name,
		current.name, target.name,
		"author: " + current.owner, "author: " + target.owner,
		"codecov.io/gh/" + current.owner + "/", "codecov.io/gh/" + target.owner + "/",
	}
}

func (replacer identifierReplacer) replace(contents []byte) []byte {
	var result strings.Builder
	result.Grow(len(contents))

	for line := range strings.SplitAfterSeq(string(contents), "\n") {
		if strings.Contains(line, "Copyright (c) ") {
			result.WriteString(replacer.copyright.Replace(line))
			continue
		}
		result.WriteString(replacer.standard.Replace(line))
	}
	return []byte(result.String())
}

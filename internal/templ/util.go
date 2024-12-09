// Copyright © 2018 Ray Johnson <ray.johnson@gmail.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package templ

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var multiNewlineRegex = regexp.MustCompile(`\n+\n`)

func SimpleToMdoc(str string) string {
	// Guessing this is already troff - so let it pass through
	if len(str) > 1 && str[0] == '.' {
		return str
	}

	// TODO: this could certainly be more sophisticated.  Pull requests welcome!
	// Right now it is good enough for the most simple cases.
	return Backslashify(multiNewlineRegex.ReplaceAllString(str, "\n.Pp\n"))
}

func SimpleToTroff(str string) string {
	// Guessing this is already troff - so let it pass through
	if len(str) > 1 && str[0] == '.' {
		return str
	}

	// TODO: this could certainly be more sophisticated.  Pull requests welcome!
	// Right now it is good enough for the most simple cases.
	return Backslashify(multiNewlineRegex.ReplaceAllString(str, "\n.PP\n"))
}

var backslashReplacer *strings.Replacer

func Backslashify(str string) string {
	if backslashReplacer == nil {
		backslashReplacer = strings.NewReplacer("-", "\\-", "_", "\\_", "&", "\\&", "\\", "\\\\", "~", "\\~")
	}
	return backslashReplacer.Replace(str)
}

func Dashify(str string) string {
	return strings.ReplaceAll(str, " ", "-")
}

func Underscoreify(str string) string {
	return strings.ReplaceAll(str, " ", "_")
}

func TrimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// PadR adds padding to the right of a string.
func PadR(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

func Makeline(str string, char byte) string {
	b := make([]byte, len(str))
	for i := range b {
		b[i] = char
	}
	return string(b)
}

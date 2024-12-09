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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackslashify(t *testing.T) {
	cases := [][]string{
		{`foo-bar`, `foo\-bar`},
		{`foo&bar`, `foo\&bar`},
		{`foo_bar`, `foo\_bar`},
		{`foo\bar`, `foo\\bar`},
		{`foo~bar`, `foo\~bar`},
		{`-_&\~`, `\-\_\&\\\~`},
	}

	for i := 0; i < len(cases); i++ {
		str := backslashify(cases[i][0])
		expected := cases[i][1]
		assert.Equal(t, expected, str)
	}
}

func TestDashify(t *testing.T) {
	cases := [][]string{
		{`foo bar`, `foo-bar`},
		{`foo bar cat`, `foo-bar-cat`},
		{` foo bar `, `-foo-bar-`},
	}

	for i := 0; i < len(cases); i++ {
		str := dashify(cases[i][0])
		expected := cases[i][1]
		assert.Equal(t, expected, str)
	}
}

func TestUnderscoreify(t *testing.T) {
	cases := [][]string{
		{`foo bar`, `foo_bar`},
		{`foo bar cat`, `foo_bar_cat`},
		{` foo bar `, `_foo_bar_`},
	}

	for i := 0; i < len(cases); i++ {
		str := underscoreify(cases[i][0])
		expected := cases[i][1]
		assert.Equal(t, expected, str)
	}
}

func TestSimpleToTroff(t *testing.T) {
	cases := [][]string{
		{"Some test\none a line", "Some test\none a line"},
		{"Some test\n\nwith empty line", "Some test\n.PP\nwith empty line"},
		{".ignore me\n\none a line", ".ignore me\n\none a line"},
		{"Some test\n\n\nwith empty line", "Some test\n.PP\nwith empty line"},
		{"Some test\n\n\n\nwith empty line", "Some test\n.PP\nwith empty line"},
	}

	for i := 0; i < len(cases); i++ {
		str := simpleToTroff(cases[i][0])
		expected := cases[i][1]
		assert.Equal(t, expected, str)
	}
}

func TestSimpleToMdoc(t *testing.T) {
	cases := [][]string{
		{"Some test\none a line", "Some test\none a line"},
		{"Some test\n\nwith empty line", "Some test\n.Pp\nwith empty line"},
		{".ignore me\n\none a line", ".ignore me\n\none a line"},
		{"Some test\n\n\nwith empty line", "Some test\n.Pp\nwith empty line"},
		{"Some test\n\n\n\nwith empty line", "Some test\n.Pp\nwith empty line"},
	}

	for i := 0; i < len(cases); i++ {
		str := simpleToMdoc(cases[i][0])
		expected := cases[i][1]
		assert.Equal(t, expected, str)
	}
}

func TestRpad(t *testing.T) {
	cases := [][]string{
		{"foo", "10", "foo       x"},
		{"foo bar", "10", "foo bar   x"},
		{"foo bar cat", "10", "foo bar catx"},
	}

	for i := 0; i < len(cases); i++ {
		pad, _ := strconv.Atoi(cases[i][1])
		str := rpad(cases[i][0], pad) + "x"
		expected := cases[i][2]
		assert.Equal(t, expected, str)
	}
}

func TestTrimRightSpace(t *testing.T) {
	cases := [][]string{
		{"foo   ", "foox"},
		{"foo", "foox"},
	}

	for i := 0; i < len(cases); i++ {
		str := trimRightSpace(cases[i][0]) + "x"
		expected := cases[i][1]
		assert.Equal(t, expected, str)
	}
}

func TestMakeline(t *testing.T) {
	cases := [][]string{
		{"foo", "-", "---"},
		{"foo bar", "*", "*******"},
	}

	for i := 0; i < len(cases); i++ {
		str := makeline(cases[i][0], cases[i][1][0])
		expected := cases[i][2]
		assert.Equal(t, expected, str)
	}
}

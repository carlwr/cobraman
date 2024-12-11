package fjoin_test

import (
	"fmt"
	"testing"

	"github.com/carlwr/cobraman/internal/tests/fjoin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tc struct {
	args []string
	want string
}

func TestRelative(t *testing.T) {
	tcs := []tc{
		{[]string{"a"},
			"a"},
		{[]string{"a/b/"},
			"a/b"},
		{[]string{"a", "b"},
			"a/b"},
		{[]string{"a b"},
			"a b"},
		{[]string{"a<b"},
			"a_b"},
		{[]string{"ab<cd"},
			"ab_cd"},
		{[]string{"<"},
			"_"},
		{[]string{"</b"},
			"_/b"},
		{[]string{"<", ">", "/</"},
			"_/_/_"},
	}
	run(t, tcs)
}

func TestAbsolute(t *testing.T) {
	tcs := []tc{
		{[]string{"/a"},
			"/a"},
		{[]string{"/a/"},
			"/a"},
		{[]string{"/a/b/"},
			"/a/b"},
		{[]string{"//a/", "//b/"},
			"/a/b"},
	}
	run(t, tcs)
}

func TestEmptyComponents(t *testing.T) { // ...should be ignored

	t.Run(
		"empty elements",
		func(t *testing.T) {
			expected := "aA/b"
			args := [][]string{
				{"aA", "b"},
				{"aA", "", "b"},
				{"aA", "", "", "b"},
				{"", "aA", "b"},
				{"aA", "b", ""},
				{"", "", "aA", "", "", "b", "", ""},
			}
			runExpected(t, expected, args)

		})

	t.Run(
		"slashes in elements",
		func(t *testing.T) {
			expected := "a/A/b"
			args := [][]string{
				{"a/A", "b", "/"},
				{"a//A", "b", "/"},
				{"a//A/", "/", "", "/b", "//", "/"},
			}
			runExpected(t, expected, args)
		})

	t.Run(
		"absolute paths",
		func(t *testing.T) {
			expected := "/a/A/b"
			args := [][]string{
				{"/a/A", "b", "/"},
				{"/a/A", "/", "b"},
				{"/a/A", "b"},
				{"/", "a/A", "b"},
				{"//", "a/A", "b"},
				{"/a//A", "b", "/"},
				{"/a//A/", "/", "", "/b", "//", "/"},
			}
			runExpected(t, expected, args)
		})

}

func TestPeculiarities( // ...of Filenamify
	t *testing.T) {

	t.Run(
		"leading/trailing non-path chars", // ...are _removed_
		func(t *testing.T) {
			tcs := []tc{
				{[]string{"<b"},
					"b"},
				{[]string{"b>"},
					"b"},
				{[]string{"/<b"},
					"/b"},
				{[]string{"a", "<b"},
					"a/b"},
				{[]string{"a", "/<b"},
					"a/b"},
			}
			run(t, tcs)
		})

	t.Run(
		"sequences of non-path chars", // ...are replaced with a single replacement char",
		func(t *testing.T) {
			tcs := []tc{
				{[]string{"a<<b"},
					"a_b"},
				{[]string{"<<b"},
					"b"},
				{[]string{"/<<b"},
					"/b"},
				{[]string{"a", "<<b"},
					"a/b"},
				{[]string{"a", "/<<b"},
					"a/b"},
			}
			run(t, tcs)
		})

}

func run(t *testing.T, tcs []tc) {
	for i, tc := range tcs {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			got, err := fjoin.Join(tc.args...)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			if t.Failed() {
				t.Logf("arguments:\n\t%v", tc.args)
			}
		})
	}
}

func runExpected(t *testing.T, expected string, args [][]string) {
	var tcs []tc
	for _, a := range args {
		tcs = append(tcs, tc{a, expected})
	}
	run(t, tcs)
}

package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilenamifyJoin(t *testing.T) {
	tcs := []struct {
		args []string
		want string
	}{
		{[]string{"a"},
			"a"},
		{[]string{"/a"},
			"/a"},
		{[]string{"/a/"},
			"/a"},
		{[]string{"/a/b/"},
			"/a/b"},
		{[]string{"a/b/"},
			"a/b"},
		{[]string{"a", "b"},
			"a/b"},
		{[]string{"//a/", "//b/"},
			"/a/b"},
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

		// peculiarities of filenamify:
		// _removes_ non-path characters if leading or trailing:
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
		// replaces sequences of non-path characters with a single replacement character:
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

	for i, tc := range tcs {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			got, err := FilenamifyJoin(tc.args...)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			if t.Failed() {
				t.Logf("arguments:\n\t%v", tc.args)
			}
		})
	}
}

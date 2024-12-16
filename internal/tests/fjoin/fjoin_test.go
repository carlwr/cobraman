package fjoin_test

import (
	"encoding/csv"
	"fmt"
	"strings"
	"testing"

	"github.com/carlwr/cobraman/internal/tests/fjoin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// test cases csv multiline strings:
// * fields are separated by commas
// * last filed is the expexted output string, preceding fields are the elements of the input list-of-strings
// * any leading whitespaces in fields are trimmed
// * fields may be quoted with double quotes (not required)
// * double quotes in fields are escaped by doubling them
//
// chosen conventions:
// * multiline strings are prepended with whitespace, not tabs, to maintain indentation

// Whether the testlog should report the input strings, output string and expected string for each test case.
const Verbose = false

type suite struct {
	name   string
	tcsStr string
}

var std = []suite{
	{
		name: "relative paths",
		tcsStr: `
      a,                           a
      a/b/,                        a/b
      a,     b,                    a/b
      a b,                         a b
      a<b,                         a_b
      ab<cd,                       ab_cd
      <,                           _
      </b,                         _/b
      <,     >,  /</,              _/_/_`,
	},
	{
		name: "absolute paths",
		tcsStr: `
      /a,                          /a
      /a/,                         /a
      /a/b/,                       /a/b
      //a/,  //b/,                 /a/b`,
	},
	{
		name: "ignoring empty components",
		tcsStr: `
      aA,  bb,                     aA/bb
      aA,  "",  bb,                aA/bb
      aA,  "",  "",  bb,           aA/bb
      "",  aA,  bb,                aA/bb
      aA,  bb,  "",                aA/bb
      "",  "",  aA,  bb,           aA/bb
      aA,  bb,  "",  "",           aA/bb`,
	},
	{
		name: "ignoring extra slashes",
		tcsStr: `
        a/A,    b,   /,            a/A/b
        a//A,   b,   /,            a/A/b
        a//A/,  /,  "",  /b,       a/A/b
        a//A/,  /b,  //,  /,       a/A/b`,
	},
	{
		name: "ignoring extra slashes (absolute paths)",
		tcsStr: `
      /a/A,    b,    /,            /a/A/b
      /a/A,    /,    b,            /a/A/b
      /a/A,    b,                  /a/A/b
      /,       a/A,  b,            /a/A/b
      //,      a/A,  b,            /a/A/b
      /a//A,   b,    /,            /a/A/b
      /a/A/b,  /,    "",           /a/A/b
      /a/A/,   "",   /b,           /a/A/b
      /a/A/b,  //,   /,            /a/A/b`,
	},
}

func TestStd(t *testing.T) {
	runSuites(t, std)
}

var peculiarities = /* of filenamify */ []suite{
	{
		name: "leading+trailing non-path chars trimmed",
		tcsStr: `
      <b,                          b
      b>,                          b
      /<b,                         /b
      a,    <b,                    a/b
      a,    /<b,                   a/b`,
	},
	{
		name: "sequences of such are too",
		tcsStr: `
      a<<b,                        a_b
      <<b,                         b
      /<<b,                        /b
      a,     <<b,                  a/b
      a,     /<<b,                 a/b`,
	},
}

func TestPeculiarities(t *testing.T) {
	runSuites(t, peculiarities)
}

type tc struct {
	args []string
	want string
}

func runTc(t *testing.T, tcs []tc) {
	for i, tc := range tcs {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			got, err := fjoin.Join(tc.args...)

			logTC := func() {
				t.Logf("\nargs:  %#v\ngot:   %#v\nwant:  %#v", tc.args, got, tc.want)
			}

			require.NoError(t, err)
			passed := assert.Equal(t, tc.want, got)

			if Verbose || !passed {
				logTC()
			}
		})
	}
}

func runSuite(t *testing.T, s suite) {
	t.Run(s.name, func(t *testing.T) {
		runTc(t, loadCSV(s.tcsStr))
	})
}

func runSuites(t *testing.T, s []suite) {
	for _, suite := range s {
		runSuite(t, suite)
	}
}

func loadCSV(csvStr string) []tc {
	r := csv.NewReader(strings.NewReader(csvStr))
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		panic(fmt.Errorf("loadFromCSV: %w", err))
	}

	var tcs []tc
	for i, rec := range records {
		if len(rec) < 2 {
			panic(fmt.Errorf("record %d: got %d field(s), need >=2", i, len(rec)))
		}
		args := rec[:len(rec)-1]
		want := rec[len(rec)-1]
		tcs = append(tcs, tc{args, want})
	}

	return tcs
}

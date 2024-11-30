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

package cobraman

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/flytam/filenamify"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loop variables not captured, so must be compiled with Go version >= 1.22

/*
for testing test robustness:
* parallel tests
    t.Parallel
* `test` switches
    -race
		-count=
		-shuffle=on


*/

type preserveTmpPolicy int

const (
	pFailing preserveTmpPolicy = iota
	pAlways
	pNever
)

type preserveTmpCfg struct {
	policy preserveTmpPolicy
	dir    string
}

type testsCfgT struct {
	preserve preserveTmpCfg
}

var testsCfg = testsCfgT{
	preserve: preserveTmpCfg{
		policy: pFailing,
		dir:    "/tmp/cobraman",
	},
}

var testInvokedAt time.Time

func TestMain(m *testing.M) {
	testInvokedAt = time.Now()
	os.Exit(m.Run())
}

// Like `filepath.Join()`, but additionally filenamifies each individual path component.
func filenamifyJoin(parts ...string) (string, error) {

	opts := filenamify.Options{Replacement: "_"}

	var fixeds []string
	isAbs := filepath.IsAbs(parts[0])

	for _, part := range parts {
		partCl := filepath.Clean(part)
		splitted := strings.Split(partCl, "/")
		for _, elem := range splitted {
			if elem == "" {
				continue
			}
			fixed, err := filenamify.Filenamify(elem, opts)
			if err != nil {
				return "", err
			}
			fixeds = append(fixeds, fixed)
		}
	}

	joined := filepath.Join(fixeds...)
	if isAbs {
		joined = string(filepath.Separator) + joined
	}
	return joined, nil
}

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
			got, err := filenamifyJoin(tc.args...)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			if t.Failed() {
				t.Logf("arguments:\n\t%v", tc.args)
			}
		})
	}
}

func preserve(t *testing.T, dir string) {

	sinceAlways := (testsCfg.preserve.policy == pAlways)
	sinceFailin := (testsCfg.preserve.policy == pFailing) && t.Failed()
	doPreserve := sinceAlways || sinceFailin

	if doPreserve {
		var err error

		toDir, err := filenamifyJoin(
			testsCfg.preserve.dir,
			testInvokedAt.Format("Mon_150405.0000"),
			t.Name(),
		)
		if err != nil {
			t.Logf("WARNING: failed to filenamify:\n  '%v'", err)
			return
		}

		err = os.CopyFS(toDir, os.DirFS(dir))
		if err != nil {
			t.Logf("WARNING: failed to preserve:\n  '%v'", err)
			return
		}
		t.Logf("info: preserved temp dir:\n  %s", toDir)
	}
}

// Returns a temp dir for the test, automatically deleted when the test is complete. If the tests are configured to preserve
// if test fails, the dir is preserved, and a message printed about where
func tempDir(t *testing.T) string {
	tmpDir := t.TempDir()
	t.Cleanup(func() { preserve(t, tmpDir) })
	return tmpDir
}

type runFunc func(*cobra.Command, []string)

func mkMockRunFunc() runFunc {
	f := func(cmd *cobra.Command, args []string) {}
	return f
}

func mkCobraCmd(name string, setRunFunc bool) *cobra.Command {
	cmd := &cobra.Command{Use: name}
	if setRunFunc {
		cmd.Run = mkMockRunFunc()
	}
	return cmd
}

// removes any tabs + whitespaces following a newline
func dedent(s string) string {
	return regexp.MustCompile(`\n\t+ *`).ReplaceAllString(s, "\n")
}

func expectedFname(dir, cmd, subCmdSep, subCmd, extFstr, sec string) string {
	stem := cmd
	if subCmd != "" {
		stem += subCmdSep + subCmd
	}
	ext := extFstr
	if strings.Contains(extFstr, "%s") {
		ext = fmt.Sprintf(extFstr, sec)
	}
	return filepath.Join(dir, stem+ext)
}

func expectedSec(opts Options) string {
	if opts.Section == "" {
		return "1"
	}
	return opts.Section

}

type fmtCfg struct {
	ext       string
	subCmdSep string
}

var fmts = map[string]fmtCfg{
	"troff": {
		ext:       ".%s",
		subCmdSep: "-"},
	"mdoc": {
		ext:       ".%s",
		subCmdSep: "-"},
	"markdown": {
		ext:       ".md",
		subCmdSep: "_"},
}

func TestFileCreation(t *testing.T) {

	tcCmds := []struct {
		cmd string
		opt Options
	}{
		{"fooCmd", Options{}},
		{"barCmd", Options{Section: "1"}},
		{"bazCmd", Options{Section: "8"}},
		{"fo sub", Options{}},
	}

	for fmt, fmtCfg := range fmts {
		fmt := fmt
		fmtCfg := fmtCfg

		t.Run(fmt, func(t *testing.T) {

			t.Run("no-cmd", func(t *testing.T) {
				err := GenerateDocs(&cobra.Command{}, &Options{}, tempDir(t), fmt)
				assert.Equal(t, "you need a command name to have a man page", err.Error())
			})

			for _, tcCmd := range tcCmds {
				tcCmd := tcCmd
				t.Run(tcCmd.cmd, func(t *testing.T) {
					tmpD := tempDir(t)

					// some shorthand functions:
					genDocs := func(cobraCmd *cobra.Command) error {
						optCopy := tcCmd.opt
						return GenerateDocs(cobraCmd, &optCopy, tmpD, fmt)
					}
					expectedFname_ := func(cmd, subCmd string) string {
						return expectedFname(tmpD, cmd, fmtCfg.subCmdSep, subCmd, fmtCfg.ext, expectedSec(tcCmd.opt))
					}

					cmdHierarchy := strings.Fields(tcCmd.cmd)

					switch len(cmdHierarchy) {
					case 1:
						assert.Nil(t, genDocs(&cobra.Command{Use: cmdHierarchy[0]}))
						assert.FileExists(t, expectedFname_(cmdHierarchy[0], ""))
					case 2:
						mainCmd := &cobra.Command{Use: cmdHierarchy[0]}
						subCmd := &cobra.Command{Use: cmdHierarchy[1], Run: mkMockRunFunc()}
						mainCmd.AddCommand(subCmd)

						assert.Nil(t, genDocs(mainCmd))
						assert.FileExists(t, expectedFname_(cmdHierarchy[0], ""))
						assert.FileExists(t, expectedFname_(cmdHierarchy[0], cmdHierarchy[1]))
					default:
						t.Fatalf("invalid cmdHierarchy: %v", cmdHierarchy)
					}
				})
			}

		})

	}
}

func TestDefaultOpts(t *testing.T) {
	opts := Options{}

	validate(&opts, "troff")
	assert.Equal(t, opts.Section, "1")
	assert.Equal(t, opts.fileCmdSeparator, "-")
	assert.Equal(t, opts.fileSuffix, "1")

	validate(&opts, "mdoc")
	assert.Equal(t, opts.Section, "1")
	assert.Equal(t, opts.fileCmdSeparator, "-")
	assert.Equal(t, opts.fileSuffix, "1")

	opts = Options{}
	validate(&opts, "markdown")
	assert.Equal(t, opts.Section, "1")
	assert.Equal(t, opts.fileCmdSeparator, "_")
	assert.Equal(t, opts.fileSuffix, "md")

	delta := time.Since(*opts.Date)
	if delta.Seconds() >= 1 {
		assert.Fail(t, "time difference too large")
	}

	opts = Options{}
	assert.Panics(t, func() { validate(&opts, "no exist") }, "should have paniced")
}

func mkDate(s string) *time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return &t
}

type format int

const (
	troff format = iota
	mdoc
	md
)

var templName = map[format]string{
	troff: "troff",
	mdoc:  "mdoc",
	md:    "markdown",
}

func (ss format) String() string {
	return templName[ss]
}

type wantRegexes map[format]([]string)

func TestOptions_New(t *testing.T) {

	cmd := cobra.Command{Use: "foo"}

	var tcs = []struct {
		opt  Options
		want wantRegexes
	}{
		{
			opt: Options{},
			want: wantRegexes{
				troff: {`\.TH "FOO" "1"`},
				mdoc:  {`\.Dt FOO 1`}},
		}, {
			opt: Options{Section: "3"},
			want: wantRegexes{
				troff: {`\.TH "FOO" "3"`},
				mdoc:  {`\.Dt FOO 3`}},
		}, {
			opt: Options{
				LeftFooter:   "left footer",
				CenterHeader: "centerHeader",
				CenterFooter: "centerFooter",
				Section:      "3",
			},
			want: wantRegexes{
				troff: {`\.TH "FOO" "3" "centerFooter" "left footer" "centerHeader"`},
				mdoc:  {`\.Dt FOO 3`}, // custom header/footer not supported
			},
		}, {
			opt: Options{Date: mkDate("1968-06-21T15:04:05Z")},
			want: wantRegexes{
				troff: {`\.TH "FOO" "1" "Jun(e?) 1968"`},
				mdoc:  {`\.Dd Jun(e?) 1968`},
			},
		},
	}

	for i, tc := range tcs {
		t_Run(t, i, cmd, tc.opt, tc.want)
	}
}

func genDoc(cmd cobra.Command, opts Options, dir string, formt format) (err error) {
	cmdCopy := cmd
	optCopy := opts
	return GenerateDocs(&cmdCopy, &optCopy, dir, formt.String())
}

func genPage(cmd cobra.Command, opts Options, formt format) (buf *bytes.Buffer, err error) {
	cmdCopy := cmd
	optCopy := opts
	buf_ := new(bytes.Buffer)
	err_ := GenerateOnePage(&cmdCopy, &optCopy, formt.String(), buf_)
	return buf_, err_
}

func t_Run(t *testing.T, i int, cmd cobra.Command, opts Options, wants wantRegexes) {

	t.Run(fmt.Sprint(i), func(t *testing.T) {
		for format, wantREs := range wants {
			t.Run(format.String(), func(t *testing.T) {

				t.Run("docs_noError", func(t *testing.T) {
					tmpD := tempDir(t)
					assert.NoError(t, genDoc(cmd, opts, tmpD, format))
				})

				t.Run("onePage_regexes", func(t *testing.T) {
					buf, err := genPage(cmd, opts, format)
					require.NoError(t, err)
					for j, wantRE := range wantREs {
						t.Run(fmt.Sprint(j), func(t *testing.T) {
							assert.Regexp(t, wantRE, buf.String())
						})
					}

				})
			})
		}
	})
}

func TestMisc(t *testing.T) {
	// this function is quiet terrible; work in progress to replace it

	// extended from TestFileCreation

	type formatPatterns struct {
		fmtName  string
		patterns map[string]string
	}

	formats := []formatPatterns{
		{
			fmtName: "troff",
			patterns: map[string]string{
				"header_custom":    `\.TH "%s" "%s" "%s" "%s" "%s"`,
				"header_toolName":  `\.TH "%s" "%s"`,
				"header_date":      `\.TH .* "%s"`,
				"name":             `\.SH NAME\n%s( - )?%s\n`,
				"synopsis":         `\.SH SYNOPSIS\n\.sp\n.+%s`,
				"synopsis_subcmds": `\.SH SYNOPSIS\n\.sp(\n.+%s (%s|%s).+flags.+\n\.br){2}`,
				"synopsis_flags":   `\.SH SYNOPSIS\n\.sp\n.+%s.+\\-\\-%s.+<args>]`,
				"description":      `\.SH DESCRIPTION\n\.PP\n%s`,
				"description_long": `\.SH DESCRIPTION\n\.PP\n%s\n\.PP\n%s`,
			},
		},
		{
			fmtName: "mdoc",
			patterns: map[string]string{
				"header_toolName":  `\.Dt %s %s`,
				"header_custom":    `.*(%s%s%s%s%s)?`, // not supported by mdoc
				"header_date":      `\.Dd %s`,
				"name":             `\.Sh NAME\n\.Nm %s\n(\.Nd %s\n)?\.Sh SYNOPSIS`,
				"synopsis":         `\.Sh SYNOPSIS\n\.Nm %s\n\.Op Fl <args>\n\.Ek\n\.Sh DESCRIPTION`,
				"synopsis_subcmds": `\.Sh SYNOPSIS(\n\.Nm %s (%s|%s) Op Fl flags Op args){2}\n\.Ek\n\.Sh DESCRIPTION`,
				"synopsis_flags":   `\.Sh SYNOPSIS\n\.Nm %s\n\.Op Fl \\-%s\n\.Op Fl <args>`,
				"description":      `\.Sh DESCRIPTION\n%s`,
				"description_long": `\.Sh DESCRIPTION\n%s\n\.Pp\n%s`,
			},
		},
	}

	type expectedPattern struct {
		key  string
		args []interface{}
	}

	type testCase struct {
		description      string
		cmd              *cobra.Command
		opts             Options
		expectedPatterns []expectedPattern
	}

	testCases := []testCase{
		{
			description: "header-toolname",
			cmd:         &cobra.Command{Use: "foo"},
			opts:        Options{},
			expectedPatterns: []expectedPattern{
				{"header_toolName", []interface{}{"FOO", "1"}},
			},
		},
		{
			description: "header-custom",
			cmd:         &cobra.Command{Use: "foo"},
			opts: Options{
				LeftFooter:   "kitty kat",
				CenterHeader: "Hello",
				CenterFooter: "meow",
				Section:      "3",
			},
			expectedPatterns: []expectedPattern{
				{"header_toolName", []interface{}{"FOO", "3"}},
				{"header_custom", []interface{}{"FOO", "3", "meow", "kitty kat", "Hello"}},
			},
		},
		{
			description: "header-date",
			cmd:         &cobra.Command{Use: "foo"},
			opts: Options{
				Date: func() *time.Time {
					t, _ := time.Parse(time.RFC3339, "1968-06-21T15:04:05Z")
					return &t
				}(),
			},
			expectedPatterns: []expectedPattern{
				{"header_toolName", []interface{}{"FOO", "1"}},
				{"header_date", []interface{}{"Jun(e?) 1968"}},
			},
		},
		{
			description: "name",
			cmd:         &cobra.Command{Use: "foo"},
			opts:        Options{},
			expectedPatterns: []expectedPattern{
				{"name", []interface{}{"foo", ""}},
			},
		},
		{
			description: "name-shortDesc",
			cmd:         &cobra.Command{Use: "bar", Short: "going to"},
			opts:        Options{},
			expectedPatterns: []expectedPattern{
				{"name", []interface{}{"bar", "going to"}},
			},
		},
		{
			description: "synopsis-subcmds",
			cmd: func() *cobra.Command {
				cmdCat := &cobra.Command{Use: "cat", Run: mkMockRunFunc()}
				cmdDog := &cobra.Command{Use: "dog", Run: mkMockRunFunc()}
				cmdAnm := &cobra.Command{Use: "anm"}
				cmdAnm.AddCommand(cmdCat, cmdDog)
				return cmdAnm
			}(),
			opts: Options{},
			expectedPatterns: []expectedPattern{
				{"synopsis_subcmds", []interface{}{"anm", "cat", "dog"}},
			},
		},
		{
			description: "synopsis-flags",
			cmd: func() *cobra.Command {
				cmdFlag := &cobra.Command{Use: "foo"}
				cmdFlag.Flags().String("thing", "", "string with no default")
				return cmdFlag
			}(),
			opts: Options{},
			expectedPatterns: []expectedPattern{
				{"synopsis_flags", []interface{}{"foo", "thing"}},
			},
		},
		{
			description: "desc-shortDesc",
			cmd:         &cobra.Command{Use: "bar", Short: "going to"},
			opts:        Options{},
			expectedPatterns: []expectedPattern{
				{"description", []interface{}{"going to"}},
			},
		},
		{
			description: "desc-longDesc",
			cmd: &cobra.Command{
				Use:   "bar",
				Short: "going to",
				Long:  "Long desc\n\nThis is long & stuff.",
			},
			opts: Options{},
			expectedPatterns: []expectedPattern{
				{"description_long", []interface{}{"Long desc", "This is long \\\\& stuff\\."}},
			},
		},
	}

	for _, format := range formats {
		format := format
		t.Run(format.fmtName, func(t *testing.T) {
			for _, tc := range testCases {
				tc := tc
				t.Run(tc.description, func(t *testing.T) {
					var buf bytes.Buffer
					optCopy := tc.opts

					err := GenerateOnePage(tc.cmd, &optCopy, format.fmtName, &buf)
					assert.NoError(t, err)

					for _, ep := range tc.expectedPatterns {
						ep := ep
						patternTemplate, ok := format.patterns[ep.key]
						if !ok {
							t.Errorf("Pattern key %s not found for format %s, skipping", ep.key, format.fmtName)
							continue
						}
						regex := fmt.Sprintf(patternTemplate, ep.args...)
						if assert.NotContains(t, regex, "(EXTRA string=") {
							assert.Regexp(t, regex, buf.String())
						}
					}
				})
			}
		})
	}
}

func TestSec_required(t *testing.T) {

	cmd := &cobra.Command{Use: "foo"}
	cmd_shrt := &cobra.Command{Use: "bar", Short: "going to"}

	tcs := []struct {
		fmt          string
		header       string
		sec_name     string
		sec_synopsis string
	}{
		{
			fmt:          "troff",
			header:       "\\.TH \"%s\" \"%s\" \"%s\" \"%s\" \"%s\"",
			sec_name:     "\\.SH NAME\n%s( - )?%s\n",
			sec_synopsis: "\\.SH SYNOPSIS\n.sp\n.+%s",
		},
		{
			fmt:          "mdoc",
			header:       "\\.Dt %s %s(%s%s%s){0}",
			sec_name:     "\\.Sh NAME\n\\.Nm %s\n(\\.Nd %s\n)?\\.Sh SYNOPSIS",
			sec_synopsis: "\\.Sh SYNOPSIS\n\\.Nm %s\n\\.Op Fl <args>\n",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.fmt, func(t *testing.T) {

			t.Run("header", func(t *testing.T) {

				buf := new(bytes.Buffer)
				opts := Options{}
				assert.NoError(t, GenerateOnePage(cmd, &opts, tc.fmt, buf))
				assert.Regexp(t, fmt.Sprintf(tc.header, "FOO", "1", ".*", "", ""), buf.String())

				buf.Reset()
				opts = Options{
					LeftFooter:   "kitty kat",
					CenterHeader: "Hello",
					CenterFooter: "meow",
					Section:      "3"}
				assert.NoError(t, GenerateOnePage(cmd, &opts, tc.fmt, buf))
				assert.Regexp(t, fmt.Sprintf(tc.header, "FOO", "3", "meow", "kitty kat", "Hello"), buf.String())

				buf.Reset()
				opts = Options{
					Date: func() *time.Time {
						t, _ := time.Parse(time.RFC3339, "1968-06-21T15:04:05Z")
						return &t
					}(),
				}
				assert.NoError(t, GenerateOnePage(cmd, &opts, tc.fmt, buf))
				assert.Regexp(t, fmt.Sprintf(tc.header, "FOO", "1", "Jun 1968", "", ""), buf.String())

			})

			t.Run("sec-name", func(t *testing.T) {
				buf := new(bytes.Buffer)
				opts := Options{}
				assert.NoError(t, GenerateOnePage(cmd, &opts, tc.fmt, buf))
				assert.Regexp(t, fmt.Sprintf(tc.sec_name, "foo", ""), buf.String())

				buf.Reset()
				assert.NoError(t, GenerateOnePage(cmd_shrt, &opts, tc.fmt, buf))
				assert.Regexp(t, fmt.Sprintf(tc.sec_name, "bar", "going to"), buf.String())
			})

			t.Run("sec-synopsis", func(t *testing.T) {

				t.Run("shrt", func(t *testing.T) {
					buf := new(bytes.Buffer)
					opts := Options{}
					assert.NoError(t, GenerateOnePage(cmd_shrt, &opts, tc.fmt, buf))
					// Test Synopsis
					assert.Regexp(t, fmt.Sprintf(tc.sec_synopsis, "bar"), buf.String())
				})

			})

		})
	}
}

func TestSec_Options(t *testing.T) {
	buf := new(bytes.Buffer)

	cmd := &cobra.Command{Use: "foo"}
	opts := Options{}

	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.NotRegexp(t, "\\.SH OPTIONS\n", buf.String()) // No OPTIONS section if no flags

	cmd.Flags().String("flag", "", "string with no default")
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH OPTIONS\n.TP\n.+flag.+\nstring with no default", buf.String()) // No OPTIONS section if no flags

	cmd.Flags().String("hello", "world", "default is world")
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.TP\n.+flag.+\nstring with no default", buf.String()) // No OPTIONS section if no flags
}

func TestSec_Alt(t *testing.T) {
	buf := new(bytes.Buffer)

	cmd := &cobra.Command{Use: "foo"}
	opts := Options{}

	// ENVIRONMENT
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.NotRegexp(t, "\\.SH ENVIRONMENT\n", buf.String()) // No ENVIRONMENT section if not in opts

	opts = Options{Environment: "This uses ENV"}
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH ENVIRONMENT\n.PP\nThis uses ENV\n", buf.String())

	annotations := make(map[string]string)
	annotations["man-environment-section"] = "Override at cmd level"
	cmd.Annotations = annotations
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH ENVIRONMENT\n.PP\nOverride at cmd", buf.String())

	// FILES
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.NotRegexp(t, "\\.SH FILES\n", buf.String()) // No FILES section if not in opts

	opts = Options{Files: "This uses files"}
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH FILES\n.PP\nThis uses files\n", buf.String())

	annotations = make(map[string]string)
	annotations["man-files-section"] = "Override at cmd level"
	cmd.Annotations = annotations
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH FILES\n.PP\nOverride at cmd", buf.String())

	// BUGS
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.NotRegexp(t, "\\.SH BUGS\n", buf.String()) // No BUGS section if not in opts

	opts = Options{Bugs: "You bet."}
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH BUGS\n.PP\nYou bet.\n", buf.String())

	annotations = make(map[string]string)
	annotations["man-bugs-section"] = "Override at cmd level"
	cmd.Annotations = annotations
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH BUGS\n.PP\nOverride at cmd", buf.String())

	// EXAMPLES
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.NotRegexp(t, "\\.SH EXAMPLES\n", buf.String()) // No EXAMPLES section if not in opts

	cmd.Example = "Here is example"
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH EXAMPLES\n.PP\nHere is example\n", buf.String())

	annotations = make(map[string]string)
	annotations["man-examples-section"] = "Override at cmd level"
	cmd.Annotations = annotations
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH EXAMPLES\n.PP\nOverride at cmd", buf.String())

	// AUTHOR
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	// assert.Regexp(t, "\\.SH AUTHOR\n.PP\n.SM Page auto-generated by rayjohnson/cobraman", buf.String()) // Always have AUTHOR SECTION

	opts = Options{Author: "Written by Ray Johnson"}
	buf.Reset()
	assert.NoError(t, GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH AUTHOR\nWritten by Ray Johnson\n.PP", buf.String()) // No OPTIONS section if not in opts
}

func TestBiggerExample(t *testing.T) {
	cmd1 := mkCobraCmd("bob", false)
	cmd2 := mkCobraCmd("bar", true)
	cmd3 := mkCobraCmd("foo", true)

	cmdH := mkCobraCmd("hidden", true)
	cmdH.Hidden = true

	cmd4 := mkCobraCmd("ash", true)

	cmd1.AddCommand(cmd2, cmd3, cmd4, cmdH)

	cmd5 := mkCobraCmd("dog", true)

	cmd6 := mkCobraCmd("cat", true)
	cmd6.Flags().Bool("a-boolflag", false, "Blah")
	cmd6.Flags().Bool("hiddenflag", false, "Blah")
	cmd6.Flags().Lookup("hiddenflag").Hidden = true
	cmd6.Flags().String("file", "", "Blah")
	annotation := []string{"path"}
	cmd6.Flags().SetAnnotation("file", "man-arg-hints", annotation)

	cmd3.AddCommand(cmd5, cmd6)

	opts := Options{}

	t.Run("created-files", func(t *testing.T) {
		for _, fmt := range []string{"troff", "mdoc"} {
			t.Run(fmt, func(t *testing.T) {
				tmpD := tempDir(t)

				assert.Nil(t, GenerateDocs(cmd1, &opts, tmpD, fmt))
				assert.NoFileExists(t, filepath.Join(tmpD, "bob-hidden.1"))

				for _, want := range []string{
					"bob.1",
					"bob-ash.1",
					"bob-bar.1",
					"bob-foo-cat.1",
					"bob-foo-dog.1",
					"bob-foo.1",
				} {
					assert.FileExists(t, filepath.Join(tmpD, want))
				}

				// TODO test hidden flag does not exist in bob-foo-cat.1
				// TODO: test annotation exists

			})
		}

	})

	t.Run("see-also_cmd", func(t *testing.T) {
		for _, tc := range []struct {
			fmt       string
			secHeader string
			secBody   string
		}{
			{
				fmt:       "troff",
				secHeader: "\\.SH SEE ALSO",
				secBody: dedent(
					`\.BR bob\\-ash \(1\)
					 \.BR bob\\-bar \(1\)
					 \.BR bob\\-foo \(1\)
					 `),
			},
			{
				fmt:       "mdoc",
				secHeader: `\.Sh SEE ALSO`,
				secBody: dedent(
					`\.Xr bob\\-ash 1 ,
					 \.Xr bob\\-bar 1 ,
					 \.Xr bob\\-foo 1
					 `),
			},
		} {
			t.Run(tc.fmt, func(t *testing.T) {

				tmpDir := tempDir(t)

				t.Run("genDocs-noerror", func(t *testing.T) {
					assert.Nil(t, GenerateDocs(cmd1, &Options{}, tmpDir, tc.fmt))
				})

				t.Run("genPage-text", func(t *testing.T) {
					buf := new(bytes.Buffer)
					require.NoError(t, GenerateOnePage(cmd1, &Options{}, tc.fmt, buf))
					require.Regexp(t, tc.secHeader, buf.String())
					assert.Regexp(t, tc.secBody, buf.String())
				})
			})
		}
	})

}

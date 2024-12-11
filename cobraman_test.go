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

package cobraman_test

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/carlwr/cobraman"
	"github.com/carlwr/cobraman/internal/templ"
	"github.com/carlwr/cobraman/internal/tests/tempdir"
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

var testInvokedAt time.Time

func TestMain(m *testing.M) {
	testInvokedAt = time.Now()
	os.Exit(m.Run())
}

type testsCfgT struct {
	preserve tempdir.PreserveCfg
}

var testsCfg = testsCfgT{
	preserve: tempdir.PreserveCfg{
		Policy: tempdir.P_Failing,
		Dir:    "/tmp/cobraman",
	},
}

var tempDir = tempdir.TempDirFunc(testsCfg.preserve, &testInvokedAt)

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

func expectedSec(opts cobraman.Options) string {
	if opts.Section == "" {
		return "1"
	}
	return opts.Section
}

func mkDate(s string) *time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return &t
}

// The file format for generated documentation.
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

type fmtCfg struct {
	ext       string
	subCmdSep string
}

var fmts = map[format]fmtCfg{
	troff: {
		ext:       ".%s",
		subCmdSep: "-"},
	mdoc: {
		ext:       ".%s",
		subCmdSep: "-"},
	md: {
		ext:       ".md",
		subCmdSep: "_"},
}

func TestFileCreation(t *testing.T) {

	tcCmds := []struct {
		cmd string
		opt cobraman.Options
	}{
		{"fooCmd", cobraman.Options{}},
		{"barCmd", cobraman.Options{Section: "1"}},
		{"bazCmd", cobraman.Options{Section: "8"}},
		{"fo sub", cobraman.Options{}},
	}

	for fmt, fmtCfg := range fmts {

		t.Run(fmt.String(), func(t *testing.T) {

			t.Run("no-cmd", func(t *testing.T) {
				err := cobraman.GenerateDocs(&cobra.Command{}, &cobraman.Options{}, tempDir(t), fmt.String())
				assert.Equal(t, "you need a command name to have a man page", err.Error())
			})

			for _, tcCmd := range tcCmds {
				t.Run(tcCmd.cmd, func(t *testing.T) {
					tmpD := tempDir(t)

					// some shorthand functions:
					genDocs := func(cobraCmd *cobra.Command) error {
						optCopy := tcCmd.opt
						return cobraman.GenerateDocs(cobraCmd, &optCopy, tmpD, fmt.String())
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

type wantRegexes map[format]([]string)

func TestOptions_New(t *testing.T) {

	cmd := cobra.Command{Use: "foo"}

	var tcs = []struct {
		opt  cobraman.Options
		want wantRegexes
	}{
		{
			opt: cobraman.Options{},
			want: wantRegexes{
				troff: {`\.TH "FOO" "1"`},
				mdoc:  {`\.Dt FOO 1`}},
		}, {
			opt: cobraman.Options{Section: "3"},
			want: wantRegexes{
				troff: {`\.TH "FOO" "3"`},
				mdoc:  {`\.Dt FOO 3`}},
		}, {
			opt: cobraman.Options{
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
			opt: cobraman.Options{Date: mkDate("1968-06-21T15:04:05Z")},
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

func genDoc(cmd cobra.Command, opts cobraman.Options, dir string, formt format) (err error) {
	cmdCopy := cmd
	optCopy := opts
	return cobraman.GenerateDocs(&cmdCopy, &optCopy, dir, formt.String())
}

func genPage(cmd cobra.Command, opts cobraman.Options, formt format) (buf *bytes.Buffer, err error) {
	cmdCopy := cmd
	optCopy := opts
	buf_ := new(bytes.Buffer)
	err_ := cobraman.GenerateOnePage(&cmdCopy, &optCopy, formt.String(), buf_)
	return buf_, err_
}

func t_Run(t *testing.T, i int, cmd cobra.Command, opts cobraman.Options, wants wantRegexes) {

	t.Run(fmt.Sprint(i), func(t *testing.T) {
		for format, wantREs := range wants {
			t.Run(format.String(), func(t *testing.T) {

				t.Run("docs_noError", func(t *testing.T) {
					tmpD := tempDir(t)
					assert.NoError(t, genDoc(cmd, opts, tmpD, format))
					// TODO: add optional check on names of created files
				})

				if len(wantREs) == 0 {
					return
				}
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
		opts             cobraman.Options
		expectedPatterns []expectedPattern
	}

	testCases := []testCase{
		{
			description: "header-toolname",
			cmd:         &cobra.Command{Use: "foo"},
			opts:        cobraman.Options{},
			expectedPatterns: []expectedPattern{
				{"header_toolName", []interface{}{"FOO", "1"}},
			},
		},
		{
			description: "header-custom",
			cmd:         &cobra.Command{Use: "foo"},
			opts: cobraman.Options{
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
			opts: cobraman.Options{
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
			opts:        cobraman.Options{},
			expectedPatterns: []expectedPattern{
				{"name", []interface{}{"foo", ""}},
			},
		},
		{
			description: "name-shortDesc",
			cmd:         &cobra.Command{Use: "bar", Short: "going to"},
			opts:        cobraman.Options{},
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
			opts: cobraman.Options{},
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
			opts: cobraman.Options{},
			expectedPatterns: []expectedPattern{
				{"synopsis_flags", []interface{}{"foo", "thing"}},
			},
		},
		{
			description: "desc-shortDesc",
			cmd:         &cobra.Command{Use: "bar", Short: "going to"},
			opts:        cobraman.Options{},
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
			opts: cobraman.Options{},
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

					err := cobraman.GenerateOnePage(tc.cmd, &optCopy, format.fmtName, &buf)
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
				opts := cobraman.Options{}
				assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, tc.fmt, buf))
				assert.Regexp(t, fmt.Sprintf(tc.header, "FOO", "1", ".*", "", ""), buf.String())

				buf.Reset()
				opts = cobraman.Options{
					LeftFooter:   "kitty kat",
					CenterHeader: "Hello",
					CenterFooter: "meow",
					Section:      "3"}
				assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, tc.fmt, buf))
				assert.Regexp(t, fmt.Sprintf(tc.header, "FOO", "3", "meow", "kitty kat", "Hello"), buf.String())

				buf.Reset()
				opts = cobraman.Options{
					Date: func() *time.Time {
						t, _ := time.Parse(time.RFC3339, "1968-06-21T15:04:05Z")
						return &t
					}(),
				}
				assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, tc.fmt, buf))
				assert.Regexp(t, fmt.Sprintf(tc.header, "FOO", "1", "Jun 1968", "", ""), buf.String())

			})

			t.Run("sec-name", func(t *testing.T) {
				buf := new(bytes.Buffer)
				opts := cobraman.Options{}
				assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, tc.fmt, buf))
				assert.Regexp(t, fmt.Sprintf(tc.sec_name, "foo", ""), buf.String())

				buf.Reset()
				assert.NoError(t, cobraman.GenerateOnePage(cmd_shrt, &opts, tc.fmt, buf))
				assert.Regexp(t, fmt.Sprintf(tc.sec_name, "bar", "going to"), buf.String())
			})

			t.Run("sec-synopsis", func(t *testing.T) {

				t.Run("shrt", func(t *testing.T) {
					buf := new(bytes.Buffer)
					opts := cobraman.Options{}
					assert.NoError(t, cobraman.GenerateOnePage(cmd_shrt, &opts, tc.fmt, buf))
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
	opts := cobraman.Options{}

	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.NotRegexp(t, "\\.SHcobraman.Options\n", buf.String()) // Nocobraman.Options section if no flags

	cmd.Flags().String("flag", "", "string with no default")
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH OPTIONS\n.TP\n.+flag.+\nstring with no default", buf.String()) // Nocobraman.Options section if no flags

	cmd.Flags().String("hello", "world", "default is world")
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.TP\n.+flag.+\nstring with no default", buf.String()) // Nocobraman.Options section if no flags
}

func TestSec_Alt(t *testing.T) {
	buf := new(bytes.Buffer)

	cmd := &cobra.Command{Use: "foo"}
	opts := cobraman.Options{}

	// ENVIRONMENT
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.NotRegexp(t, "\\.SH ENVIRONMENT\n", buf.String()) // No ENVIRONMENT section if not in opts

	opts = cobraman.Options{Environment: "This uses ENV"}
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH ENVIRONMENT\n.PP\nThis uses ENV\n", buf.String())

	annotations := make(map[string]string)
	annotations["man-environment-section"] = "Override at cmd level"
	cmd.Annotations = annotations
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH ENVIRONMENT\n.PP\nOverride at cmd", buf.String())

	// FILES
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.NotRegexp(t, "\\.SH FILES\n", buf.String()) // No FILES section if not in opts

	opts = cobraman.Options{Files: "This uses files"}
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH FILES\n.PP\nThis uses files\n", buf.String())

	annotations = make(map[string]string)
	annotations["man-files-section"] = "Override at cmd level"
	cmd.Annotations = annotations
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH FILES\n.PP\nOverride at cmd", buf.String())

	// BUGS
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.NotRegexp(t, "\\.SH BUGS\n", buf.String()) // No BUGS section if not in opts

	opts = cobraman.Options{Bugs: "You bet."}
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH BUGS\n.PP\nYou bet.\n", buf.String())

	annotations = make(map[string]string)
	annotations["man-bugs-section"] = "Override at cmd level"
	cmd.Annotations = annotations
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH BUGS\n.PP\nOverride at cmd", buf.String())

	// EXAMPLES
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.NotRegexp(t, "\\.SH EXAMPLES\n", buf.String()) // No EXAMPLES section if not in opts

	cmd.Example = "Here is example"
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH EXAMPLES\n.PP\nHere is example\n", buf.String())

	annotations = make(map[string]string)
	annotations["man-examples-section"] = "Override at cmd level"
	cmd.Annotations = annotations
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH EXAMPLES\n.PP\nOverride at cmd", buf.String())

	// AUTHOR
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	// assert.Regexp(t, "\\.SH AUTHOR\n.PP\n.SM Page auto-generated by rayjohnson/cobraman", buf.String()) // Always have AUTHOR SECTION

	opts = cobraman.Options{Author: "Written by Ray Johnson"}
	buf.Reset()
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "troff", buf))
	assert.Regexp(t, "\\.SH AUTHOR\nWritten by Ray Johnson\n.PP", buf.String()) // Nocobraman.Options section if not in opts
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

	opts := cobraman.Options{}

	t.Run("created-files", func(t *testing.T) {
		for _, fmt := range []string{"troff", "mdoc"} {
			t.Run(fmt, func(t *testing.T) {
				tmpD := tempDir(t)

				assert.Nil(t, cobraman.GenerateDocs(cmd1, &opts, tmpD, fmt))
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
					assert.Nil(t, cobraman.GenerateDocs(cmd1, &cobraman.Options{}, tmpDir, tc.fmt))
				})

				t.Run("genPage-text", func(t *testing.T) {
					buf := new(bytes.Buffer)
					require.NoError(t, cobraman.GenerateOnePage(cmd1, &cobraman.Options{}, tc.fmt, buf))
					require.Regexp(t, tc.secHeader, buf.String())
					assert.Regexp(t, tc.secBody, buf.String())
				})
			})
		}
	})

}

func TestCustomerTemplate(t *testing.T) {
	buf := new(bytes.Buffer)
	templ.RegisterTemplate("good", "-", "txt", "Hello {{ \"world\" }} ")
	cmd := &cobra.Command{Use: "foo"}
	opts := cobraman.Options{}
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "good", buf))
	assert.Regexp(t, "Hello world", buf.String())

}

func TestAddTemplateFunc(t *testing.T) {

	hello := func(str string) string {
		return "Hello " + str + "!"
	}

	templ.AddTemplateFunc("lower", strings.ToLower)

	var templateFuncs = template.FuncMap{
		"hello":  hello,
		"repeat": strings.Repeat,
	}

	templ.AddTemplateFuncs(templateFuncs)

	// Register template using these new functions
	templ.RegisterTemplate("tester", "-", "txt", `{{ hello "World" | lower }} {{ repeat "x" 5 }}`)
	cmd := &cobra.Command{Use: "foo"}
	opts := cobraman.Options{}
	buf := new(bytes.Buffer)
	assert.NoError(t, cobraman.GenerateOnePage(cmd, &opts, "tester", buf))
	assert.Regexp(t, "hello world!", buf.String())
	assert.Regexp(t, "xxxxx", buf.String())
}

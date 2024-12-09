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

// Package cobraman is a library for generating documentation out of a command
// line structure created by the github.com/spf13/cobra library.
package cobraman

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/carlwr/cobraman/internal/templ"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ErrMissingCommandName is returned with no command is provided.
var ErrMissingCommandName = errors.New("you need a command name to have a man page")

// Options is used configure how GenerateManPages will
// do its job.
type Options struct {
	// What section to generate the pages for (1 is the default if not set)
	Section string

	// CenterFooter used across all pages (defaults to current month and year)
	// If you just want to set the date used in the center footer use Date
	CenterFooter string

	// If you just want to set the date used in the center footer use Date
	// Will default to Now
	Date *time.Time

	// LeftFooter used across all pages
	LeftFooter string

	// CenterHeader used across all pages
	CenterHeader string

	// Files if set with content will create a FILES section for all
	// pages.  If you want this section only for a single command add
	// it as an annotation: cmd.Annotations["man-files-section"]
	// The field will be sanitized for troff output. However, if
	// it starts with a '.' we assume it is valid troff and pass it through.
	Files string

	// Bugs if set with content will create a BUGS section for all
	// pages.  If you want this section only for a single command add
	// it as an annotation: cmd.Annotations["man-bugs-section"]
	// The field will be sanitized for troff output. However, if
	// it starts with a '.' we assume it is valid troff and pass it through.
	Bugs string

	// Environment if set with content will create a ENVIRONMENT section for all
	// pages.  If you want this section only for a single command add
	// it as an annotation: cmd.Annotations["man-environment-section"]
	// The field will be sanitized for troff output. However, if
	// it starts with a '.' we assume it is valid troff and pass it through.
	Environment string

	// Author if set will create a Author section with this content.
	Author string

	// Private fields

	// fileCmdSeparator defines what character to use to separate the
	// sub commands in the man page file name.  The '-' char is the default.
	fileCmdSeparator string

	// fileSuffix is the file extension to use for file name.  Defaults to the section
	// for man templates and .md for the MarkdownTemplate template.
	fileSuffix string

	// CustomData allows passing custom data into the template
	CustomData map[string]interface{}
}

// GenerateDocs - build man pages for the passed in cobra.Command
// and all of its children.
func GenerateDocs(cmd *cobra.Command, opts *Options, directory string, templateName string) (err error) {
	// Set defaults
	validate(opts, templateName)
	if directory == "" {
		directory = "."
	}

	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenerateDocs(c, opts, directory, templateName); err != nil {
			return err
		}
	}

	// Generate file name and open the file
	basename := strings.ReplaceAll(cmd.CommandPath(), " ", opts.fileCmdSeparator)
	if basename == "" {
		return ErrMissingCommandName
	}
	filename := filepath.Join(directory, basename+"."+opts.fileSuffix)
	f, err := os.Create(filename) //nolint:gosec // the file is constructed safely
	if err != nil {
		return err
	}
	defer func() {
		err = f.Close()
	}()

	// Generate the documentation
	return GenerateOnePage(cmd, opts, templateName, f)
}

func validate(opts *Options, templateName string) {
	if opts.Section == "" {
		opts.Section = "1"
	}
	if opts.Date == nil {
		now := time.Now()
		opts.Date = &now
	}

	sep, ext, t := templ.GetTemplate(templateName)
	if t == nil {
		panic("template could not be found: " + templateName)
	}
	opts.fileCmdSeparator = sep
	opts.fileSuffix = ext
	if ext == "use_section" {
		opts.fileSuffix = opts.Section
	}
}

type manStruct struct {
	Date             *time.Time
	Section          string
	CenterFooter     string
	LeftFooter       string
	CenterHeader     string
	UseLine          string
	CommandPath      string
	ShortDescription string
	Description      string
	NoArgs           bool

	AllFlags          []manFlag
	InheritedFlags    []manFlag
	NonInheritedFlags []manFlag
	SeeAlsos          []seeAlso
	SubCommands       []*cobra.Command

	Author      string
	Environment string
	Files       string
	Bugs        string
	Examples    string

	CobraCmd *cobra.Command

	CustomData map[string]interface{}
}

type manFlag struct {
	Shorthand   string
	Name        string
	NoOptDefVal string
	DefValue    string
	Usage       string
	ArgHint     string
}

type seeAlso struct {
	CmdPath   string
	Section   string
	IsParent  bool
	IsChild   bool
	IsSibling bool
}

// GenerateOnePage will generate one documentation page and output the result to w
// TODO: document use of this function in README.
//
//nolint:funlen,gocognit,cyclop // method is readable
func GenerateOnePage(cmd *cobra.Command, opts *Options, templateName string, w io.Writer) error {
	// Set defaults - these would already be set unless GenerateOnePage called directly
	validate(opts, templateName)

	values := manStruct{}

	// Header fields
	values.LeftFooter = opts.LeftFooter
	values.CenterHeader = opts.CenterHeader
	values.Section = opts.Section
	values.Date = opts.Date
	values.CenterFooter = opts.CenterFooter
	if opts.CenterFooter == "" {
		// TODO: should this be part of template instead?
		values.CenterFooter = values.Date.Format("Jan 2006")
	}

	values.CobraCmd = cmd
	values.ShortDescription = cmd.Short
	values.UseLine = cmd.UseLine()
	values.CommandPath = cmd.CommandPath()

	// Use reflection to see if cobra.NoArgs was set
	argFuncName := runtime.FuncForPC(reflect.ValueOf(cmd.Args).Pointer()).Name()
	values.NoArgs = strings.HasSuffix(argFuncName, "cobra.NoArgs")

	if cmd.HasSubCommands() {
		subCmdArr := make([]*cobra.Command, 0, len(cmd.Commands()))
		for _, c := range cmd.Commands() {
			if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
				continue
			}
			subCmdArr = append(subCmdArr, c)
		}
		values.SubCommands = subCmdArr
	}

	// DESCRIPTION
	description := cmd.Long
	if description == "" {
		description = cmd.Short
	}
	values.Description = description

	// Flag arrays
	values.AllFlags = genFlagArray(cmd.Flags())
	values.InheritedFlags = genFlagArray(cmd.InheritedFlags())
	values.NonInheritedFlags = genFlagArray(cmd.NonInheritedFlags())

	// ENVIRONMENT section
	altEnvironmentSection := cmd.Annotations["man-environment-section"]
	if opts.Environment != "" || altEnvironmentSection != "" {
		if altEnvironmentSection != "" {
			values.Environment = altEnvironmentSection
		} else {
			values.Environment = opts.Environment
		}
	}

	// FILES section
	altFilesSection := cmd.Annotations["man-files-section"]
	if opts.Files != "" || altFilesSection != "" {
		if altFilesSection != "" {
			values.Files = altFilesSection
		} else {
			values.Files = opts.Files
		}
	}

	// BUGS section
	altBugsSection := cmd.Annotations["man-bugs-section"]
	if opts.Bugs != "" || altBugsSection != "" {
		if altBugsSection != "" {
			values.Bugs = altBugsSection
		} else {
			values.Bugs = opts.Bugs
		}
	}

	// EXAMPLES section
	altExampleSection := cmd.Annotations["man-examples-section"]
	if cmd.Example != "" || altExampleSection != "" {
		if altExampleSection != "" {
			values.Examples = altExampleSection
		} else {
			values.Examples = cmd.Example
		}
	}

	// AUTHOR section
	values.Author = opts.Author

	// SEE ALSO section
	values.SeeAlsos = generateSeeAlsos(cmd, values.Section)

	// Custom Data
	values.CustomData = opts.CustomData

	// Get template and generate the documentation page
	_, _, t := templ.GetTemplate(templateName)

	err := t.Execute(w, values)
	if err != nil {
		return err
	}
	return nil
}

func genFlagArray(flags *pflag.FlagSet) []manFlag {
	flagArray := make([]manFlag, 0, 15)
	flags.VisitAll(
		func(flag *pflag.Flag) {
			if len(flag.Deprecated) > 0 || flag.Hidden {
				return
			}
			thisFlag := manFlag{
				Name:        flag.Name,
				NoOptDefVal: flag.NoOptDefVal,
				DefValue:    flag.DefValue,
				Usage:       flag.Usage,
			}
			if flag.ShorthandDeprecated == "" {
				thisFlag.Shorthand = flag.Shorthand
			}
			hintArr, exists := flag.Annotations["man-arg-hints"]
			if exists && len(hintArr) > 0 {
				thisFlag.ArgHint = hintArr[0]
			}
			flagArray = append(flagArray, thisFlag)
		},
	)

	return flagArray
}

func generateSeeAlsos(cmd *cobra.Command, section string) []seeAlso {
	seealsos := make([]seeAlso, 0)
	if cmd.HasParent() {
		see := seeAlso{
			CmdPath:  cmd.Parent().CommandPath(),
			Section:  section,
			IsParent: true,
		}
		seealsos = append(seealsos, see)
		siblings := cmd.Parent().Commands()
		for _, c := range siblings {
			if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() || c.Name() == cmd.Name() {
				continue
			}
			see := seeAlso{
				CmdPath:   c.CommandPath(),
				Section:   section,
				IsSibling: true,
			}
			seealsos = append(seealsos, see)
		}
	}
	children := cmd.Commands()
	for _, c := range children {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		see := seeAlso{
			CmdPath: c.CommandPath(),
			Section: section,
			IsChild: true,
		}
		seealsos = append(seealsos, see)
	}

	return seealsos
}

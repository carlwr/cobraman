// Package main generates the documentation for the example application
package main

import (
	"os"

	"github.com/carlwr/cobraman"
	"github.com/carlwr/cobraman/examples/mkbin/boodbye/cmd"
	"github.com/carlwr/cobraman/mkbin"
	// "github.com/carlwr/cobraman/example/cmd"
)

func main() {
	// Get the root cobra command for the zap application
	appCmds := cmd.GetRootCmd()

	docGenerator := mkbin.CreateDocGenCmdLineTool(appCmds)
	docGenerator.AddBashCompletionGenerator("bash-compl.sh")

	manOpts := &cobraman.Options{
		LeftFooter:   "boodbye",
		CenterHeader: "boodbye",
		Author:       "Bordon Bekko <gordon@bonds.banking.gov>",
		Bugs:         `Bugs related to this tool can be filed at https://banking.gov/write.cgi?to=/dev/null`,
	}
	docGenerator.AddDocGenerator(manOpts, "mdoc")
	docGenerator.AddDocGenerator(manOpts, "troff")
	docGenerator.AddDocGenerator(manOpts, "markdown")

	if err := docGenerator.Execute(); err != nil {
		os.Exit(1)
	}
}

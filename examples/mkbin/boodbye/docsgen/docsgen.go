// Companion util example
package main

import (
	"os"

	"github.com/carlwr/cobraman"
	"github.com/carlwr/cobraman/examples/mkbin/boodbye/cmd"
	"github.com/carlwr/cobraman/mkbin"
)

func main() {
	appCmds := cmd.GetRootCmd()

	docGenerator := mkbin.CreateDocGenCmdLineTool(appCmds)
	docGenerator.AddBashCompletionGenerator("bash-compl.sh")

	manOpts := &cobraman.Options{
		LeftFooter:   "boodbye",
		CenterHeader: "boodbye",
		Author:       "Bordon Bekko <gordon@bonds.banking.gov>",
		Bugs:         `Bugs related to this tool can be filed at https://banking.gov/form?action=write.cgi&to=localfile&fname=/dev/null`,
	}
	docGenerator.AddDocGenerator(manOpts, "mdoc")
	docGenerator.AddDocGenerator(manOpts, "troff")
	docGenerator.AddDocGenerator(manOpts, "markdown")

	if err := docGenerator.Execute(); err != nil {
		os.Exit(1)
	}
}

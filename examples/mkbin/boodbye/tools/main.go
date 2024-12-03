// Package main generates the documentation for the example application
package main

import (
	"os"

	"github.com/carlwr/cobraman"
	"github.com/carlwr/cobraman/example/cmd"
)

func main() {
	// Get the root cobra command for the zap application
	appCmds := cmd.GetRootCmd()

	docGenerator := cobraman.CreateDocGenCmdLineTool(appCmds)
	docGenerator.AddBashCompletionGenerator("zap.sh")

	manOpts := &cobraman.Options{
		LeftFooter:   "Example",
		CenterHeader: "Example Manual",
		Author:       "Author Name <author@email.com>",
		Bugs:         `Bugs related to cobraman can be filed at https://github.com/carlwr/cobraman `,
	}
	docGenerator.AddDocGenerator(manOpts, "mdoc")
	docGenerator.AddDocGenerator(manOpts, "troff")
	docGenerator.AddDocGenerator(manOpts, "markdown")

	if err := docGenerator.Execute(); err != nil {
		os.Exit(1)
	}
}

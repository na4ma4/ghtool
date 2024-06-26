package main

import (
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/google/go-github/v53/github"
	"github.com/spf13/cobra"
)

var cmdRunners = &cobra.Command{
	Use:     "runner",
	Aliases: []string{"runners", "r"},
	Short:   "Runner Commands",
}

func init() {
	rootCmd.AddCommand(cmdRunners)
}

func printRunnerList(tmpl *template.Template, forceDisableHeader bool, runnerChan chan *github.Runner) {
	twOut := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0) //nolint:gomnd // standard terminal output

	if !strings.Contains(tmpl.Root.String(), "json") && strings.Contains(tmpl.Root.String(), "\t") && !forceDisableHeader {
		if err := tmpl.Execute(twOut, map[string]interface{}{
			"ID":     "ID",
			"Name":   "Name",
			"OS":     "OS",
			"Status": "Status",
			"Busy":   "Busy",
			"Labels": "Labels",
		}); err != nil {
			log.Printf("error parsing template: %s", err.Error())
		}
	}

	defer func() { _ = twOut.Flush() }()

	for in := range runnerChan {
		if err := tmpl.Execute(twOut, in); err != nil {
			log.Printf("error displaying host: %s", err.Error())
		}
	}
}

func simplePrintRunnerList(tmpl *template.Template, runnerChan chan *github.Runner) {
	for in := range runnerChan {
		if err := tmpl.Execute(os.Stdout, in); err != nil {
			log.Printf("error displaying host: %s", err.Error())
		}
	}
}

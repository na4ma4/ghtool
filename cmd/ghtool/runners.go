package main

import (
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/google/go-github/v43/github"
	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals // cobra uses globals in main
var cmdRunners = &cobra.Command{
	Use:     "runner",
	Aliases: []string{"runners", "r"},
	Short:   "Runner Commands",
}

//nolint:gochecknoinits // init is used in main for cobra
func init() {
	rootCmd.AddCommand(cmdRunners)
}

func printRunnerList(tmpl *template.Template, forceDisableHeader bool, runnerChan chan *github.Runner) {
	// sort.Slice(runnerList, func(i, j int) bool { return runnerList[i].GetID() < runnerList[j].GetID() })

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !strings.Contains(tmpl.Root.String(), "json") && strings.Contains(tmpl.Root.String(), "\t") && !forceDisableHeader {
		if err := tmpl.Execute(w, map[string]interface{}{
			"ID":     "ID",
			"Name":   "Name",
			"OS":     "OS",
			"Status": "Status",
			"Busy":   "Busy",
			"Labels": "Labels",
		}); err != nil {
			log.Printf("error pparsing template: %s", err.Error())
		}
	}

	for in := range runnerChan {
		if err := tmpl.Execute(w, in); err != nil {
			log.Printf("error displaying host: %s", err.Error())
		}
	}

	_ = w.Flush()
}

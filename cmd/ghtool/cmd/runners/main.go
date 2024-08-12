package runners

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/google/go-github/v63/github"
	"github.com/na4ma4/config"
	"github.com/na4ma4/ghtool/internal/mainconfig"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var CmdRunners = &cobra.Command{
	Use:     "runner",
	Aliases: []string{"runners", "r"},
	Short:   "Runner Commands",
}

func init() {
	CmdRunners.AddCommand(cmdRunnersList)
	CmdRunners.AddCommand(cmdRunnersWatch)
}

func getGithubClient(ctx context.Context, cfg config.Conf) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GetString("github.token")},
	)
	tkc := oauth2.NewClient(ctx, ts)

	client, err := github.NewClient(tkc).WithEnterpriseURLs(
		cfg.GetString("github.url"),
		cfg.GetString("github.url"),
	)
	if err != nil {
		return client, fmt.Errorf("unable to create github API client: %w", err)
	}

	return client, nil
}

var errConfigMissing = errors.New("config key missing")

func checkConfig(cfg config.Conf, keys ...string) error {
	missing := []string{}

	for _, key := range keys {
		if cfg.GetString(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: %s", errConfigMissing, strings.Join(missing, ", "))
	}

	return nil
}

func getTemplateFromConfig(format string, extraFunc ...template.FuncMap) (*template.Template, error) {
	if strings.Contains(format, "\\t") {
		format = strings.ReplaceAll(format, "\\t", "\t")
	}

	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}

	tmpl, err := template.New("").Funcs(mainconfig.BasicFunctions(extraFunc...)).Parse(format)
	if err != nil {
		return tmpl, fmt.Errorf("unable to create template: %w", err)
	}

	return tmpl, nil
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
	state := map[int64]*github.Runner{}
	for in := range runnerChan {
		state[in.GetID()] = in
		if v := statePrefix(state); v != "" {
			fmt.Fprint(os.Stdout, v)
		}
		if err := tmpl.Execute(os.Stdout, in); err != nil {
			log.Printf("error displaying host: %s", err.Error())
		}
	}
}

func statePrefix(state map[int64]*github.Runner) string {
	total, online, offline, busy := 0, 0, 0, 0
	for idx := range state {
		switch state[idx].GetStatus() {
		case "online":
			online++
		case "offline":
			offline++
		case "shutdown":
			delete(state, idx)
			continue
		}
		if state[idx].GetBusy() {
			busy++
		}

		total++
	}

	return fmt.Sprintf("[Onl:%02d Ofl:%02d Busy:%02d T:%02d] ", online, offline, busy, total)
}

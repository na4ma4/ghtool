package runners

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/google/go-github/v70/github"
	"github.com/na4ma4/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cmdRunnersList = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List runners and their statuses",
	RunE:    runnerListCommand,
	Args:    cobra.NoArgs,
}

func init() {
	cmdRunnersList.PersistentFlags().StringP("format", "f",
		"{{.ID}}\t{{.Name}}\t{{.OS}}\t{{.Status}}\t{{tf .Busy}}\t{{labels .Labels}}",
		"Output format (go template)",
	)
	cmdRunnersList.PersistentFlags().BoolP("raw", "r", false,
		"Raw output (no headers)",
	)

	_ = viper.BindPFlag("runner.list.raw", cmdRunnersList.PersistentFlags().Lookup("raw"))
	_ = viper.BindPFlag("runner.list.format", cmdRunnersList.PersistentFlags().Lookup("format"))
}

func runnerListCommand(_ *cobra.Command, _ []string) error {
	cfg := config.NewViperConfigFromViper(viper.GetViper(), "ghtool")

	if err := checkConfig(
		cfg,
		"github.url",
		"github.token",
		"github.enterprise",
	); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := getGithubClient(ctx, cfg)
	if err != nil {
		logrus.Panicf("unable to connect to github enterprise: %s", err)
	}

	tmpl, err := getTemplateFromConfig(
		cfg.GetString("runner.list.format"),
		template.FuncMap{"labels": templateLabels},
	)
	if err != nil {
		logrus.Panicf("unable to parse format: %s", err)
	}

	runnerChan := listEnterpriseRunners(ctx, client, cfg.GetString("github.enterprise"))

	printRunnerList(tmpl, cfg.GetBool("runner.list.raw"), runnerChan)

	return nil
}

func listEnterpriseRunners(ctx context.Context, client *github.Client, enterprise string) chan *github.Runner {
	runnerChan := make(chan *github.Runner)

	go func() {
		defer close(runnerChan)

		opts := &github.ListRunnersOptions{}

		for {
			runners, resp, err := client.Enterprise.ListRunners(ctx, enterprise, opts)
			if err != nil {
				logrus.Errorf("unable to get org runners: %s", err)

				return
			}

			for _, runner := range runners.Runners {
				// logrus.Debugf("Sending Runner to Channel: %d", runner.GetID())
				runnerChan <- runner
			}

			if opts.Page = resp.NextPage; resp.NextPage == 0 {
				return
			}
		}
	}()

	return runnerChan
}

func templateLabels(labels interface{}) string {
	switch typedLabels := labels.(type) {
	case []*github.RunnerLabels:
		out := []string{}
		for _, label := range typedLabels {
			// switch label.GetType() {
			// case "read-only":
			// 	// built-in default labels
			// 	o = append(o, fmt.Sprintf("%s (ro) [%d]", label.GetName(), label.GetID()))
			// case "custom":
			// 	// custom labels
			// 	o = append(o, fmt.Sprintf("%s [%d]", label.GetName(), label.GetID()))
			// }
			out = append(out, label.GetName())
		}

		return strings.Join(out, ", ")
	default:
		return fmt.Sprintf("%s", typedLabels)
	}
}

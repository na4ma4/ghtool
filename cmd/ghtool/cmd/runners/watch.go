package runners

import (
	"context"
	"text/template"
	"time"

	"github.com/google/go-github/v70/github"
	"github.com/na4ma4/config"
	"github.com/na4ma4/ghtool/internal/runnerlist"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cmdRunnersWatch = &cobra.Command{
	Use:     "watch",
	Aliases: []string{"w"},
	Short:   "Watch runners and their statuses",
	RunE:    runnerWatchCommand,
	Args:    cobra.NoArgs,
}

const (
	defaultTickPeriod = 30 * time.Second
)

func init() {
	cmdRunnersWatch.PersistentFlags().StringP("format", "f",
		"[{{padlen .ID 5}}] {{padlen .Name 25}} Status:{{.Status}}\tBusy:{{tf .Busy}}\t(Labels:{{labels .Labels}})",
		"Output format (go template)",
	)
	cmdRunnersWatch.PersistentFlags().BoolP("raw", "r", false,
		"Raw output (no headers)",
	)
	cmdRunnersWatch.PersistentFlags().DurationP("tick", "t", defaultTickPeriod,
		"Interval between polling for runner updates",
	)

	_ = viper.BindPFlag("runner.watch.raw", cmdRunnersWatch.PersistentFlags().Lookup("raw"))
	_ = viper.BindPFlag("runner.watch.format", cmdRunnersWatch.PersistentFlags().Lookup("format"))
	_ = viper.BindPFlag("runner.watch.tick", cmdRunnersWatch.PersistentFlags().Lookup("tick"))
}

func runnerWatchCommand(_ *cobra.Command, _ []string) error {
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
		cfg.GetString("runner.watch.format"),
		template.FuncMap{"labels": templateLabels},
	)
	if err != nil {
		logrus.Panicf("unable to parse format: %s", err)
	}

	runnerList := runnerlist.NewRunners()

	go simplePrintRunnerList(tmpl, runnerList.Channel())

	watchEnterpriseRunners(ctx, cfg, client, cfg.GetString("github.enterprise"), runnerList)

	return nil
}

func fetchEnterpriseRunners(
	ctx context.Context,
	client *github.Client,
	enterprise string,
	runnerList *runnerlist.Runners,
) {
	opts := &github.ListRunnersOptions{}

	for {
		runners, resp, err := client.Enterprise.ListRunners(ctx, enterprise, opts)
		if err != nil {
			logrus.Errorf("unable to get org runners: %s", err)

			return
		}

		runnerList.FreshnessReset()

		for _, runner := range runners.Runners {
			// logrus.Debugf("Sending Runner to Channel: %d", runner.GetID())
			_ = runnerList.Add(runner)
		}

		runnerList.PushUnfresh()

		if opts.Page = resp.NextPage; resp.NextPage == 0 {
			return
		}
	}
}

func watchEnterpriseRunners(
	ctx context.Context,
	cfg config.Conf,
	client *github.Client,
	enterprise string,
	runnerList *runnerlist.Runners,
) {
	defer runnerList.Close()

	fetchEnterpriseRunners(ctx, client, enterprise, runnerList)

	runnerTick := time.NewTicker(cfg.GetDuration("runner.watch.tick"))

	for {
		select {
		case ts := <-runnerTick.C:
			func() {
				logrus.Debugf("tick received: %s", ts.String())

				fetchEnterpriseRunners(ctx, client, enterprise, runnerList)
			}()
		case <-ctx.Done():
			return
		}
	}
}

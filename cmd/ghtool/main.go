package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/google/go-github/v43/github"
	"github.com/na4ma4/config"
	"github.com/na4ma4/ghtool/internal/mainconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

//nolint:gochecknoglobals // cobra uses globals in main
var rootCmd = &cobra.Command{
	Use: "ghtool",
}

//nolint:gochecknoinits // init is used in main for cobra
func init() {
	cobra.OnInitialize(mainconfig.ConfigInit)

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug output")
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindEnv("debug", "DEBUG")
}

func main() {
	_ = rootCmd.Execute()
}

func getGithubClient(ctx context.Context, cfg config.Conf) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GetString("github.token")},
	)
	tc := oauth2.NewClient(ctx, ts)

	// client := github.NewClient(tc)
	client, err := github.NewEnterpriseClient(
		cfg.GetString("github.url"),
		cfg.GetString("github.url"),
		tc,
	)

	return client, err
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
		format = fmt.Sprintf("%s\n", format)
	}

	tmpl, err := template.New("").Funcs(basicFunctions(extraFunc...)).Parse(format)

	return tmpl, err
}

package main

import (
	"github.com/dosquad/go-cliversion"
	"github.com/na4ma4/ghtool/cmd/ghtool/cmd/runners"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use: "ghtool",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func Help() error {
	return rootCmd.Help()
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	_ = rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug output")
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindEnv("debug", "DEBUG")

	rootCmd.AddCommand(runners.CmdRunners)

	rootCmd.Version = cliversion.Get().VersionString()
}

func main() {
	_ = Execute()
}

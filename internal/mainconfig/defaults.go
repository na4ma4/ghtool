package mainconfig

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// ConfigInit is the common config initialisation for the commands.
func ConfigInit() {
	viper.SetConfigName("ghtool")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./artifacts")
	viper.AddConfigPath("./test")
	viper.AddConfigPath("$HOME/.ghtool")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath("/run/secrets")
	viper.AddConfigPath("/etc/ghtool")
	viper.AddConfigPath("/usr/local/etc")
	viper.AddConfigPath("/usr/local/ghtool/etc")
	viper.AddConfigPath("/opt/homebrew/etc")
	viper.AddConfigPath(".")

	_ = viper.BindEnv("github.token", "GITHUB_TOKEN")
	_ = viper.BindEnv("github.url", "GITHUB_URL")
	_ = viper.BindEnv("github.enterprise", "GITHUB_ENTERPRISE")

	if viper.GetBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	_ = viper.ReadInConfig()
}

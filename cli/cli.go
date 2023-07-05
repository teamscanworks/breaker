package cli

import (
	"flag"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func RunCLI() {
	flag.String("config", "", "path to config file")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}

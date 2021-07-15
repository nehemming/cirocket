package cmd

import (
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (cli *cli) runFireCmd(cmd *cobra.Command, args []string) error {

	// Check that the init process found a config file
	if cli.initError != nil {
		return cli.initError
	}

	// Attempt to fly mission
	return rocket.Default().FlyMission(cli.ctx, viper.ConfigFileUsed(), viper.AllSettings(), args...)
}

func (cli *cli) bindLaunchFlagsAndConfig(cmd *cobra.Command) {

}

package cmd

import (
	"fmt"
	"strings"

	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func getCliParams(cmd *cobra.Command) ([]rocket.Param, error) {
	valueParams, err := cmd.Flags().GetStringArray(flagParams)
	if err != nil {
		return nil, err
	}
	return parseParams(valueParams)
}

func parseParams(valueParams []string) ([]rocket.Param, error) {
	params := make([]rocket.Param, len(valueParams))

	for i, nv := range valueParams {
		slice := strings.SplitN(nv, "=", 2)

		if len(slice) != 2 {
			return nil, fmt.Errorf("param[%d] %s is not formed as name=value", i, nv)
		}

		params[i] = rocket.Param{Name: slice[0], Value: slice[1]}
	}

	return params, nil
}

func (cli *cli) runFireCmd(cmd *cobra.Command, args []string) error {
	// Check that the init process found a config file
	if cli.initError != nil {
		return cli.initError
	}

	// Handle params
	params, err := getCliParams(cmd)
	if err != nil {
		return err
	}

	// Attempt to launch mission
	return rocket.Default().
		LaunchMissionWithParams(cli.ctx, viper.ConfigFileUsed(),
			viper.AllSettings(), params, args...)
}

const flagParams = "params"

func (cli *cli) bindLaunchFlagsAndConfig(cmd *cobra.Command) {
	cmd.Flags().StringArray(flagParams, nil, "supply parameter values to the mission, multiple params flags can be provided")
}

/*
Copyright (c) 2021 The cirocket Authors (Neil Hemming)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/spf13/cobra"
)

func (cli *cli) newLaunchCommand() *cobra.Command {
	launchCmd := &cobra.Command{
		Use:           "launch [{flightSequence}]",
		Short:         "launch the CI rocket \U0001F680 ",
		Long:          "runs the CI config, if the config uses sequences, one or more can be specified as additional args",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  false,
		RunE:          cli.runLaunchCmd,
	}

	cli.addFlagMission(launchCmd)
	return addFlagParam(launchCmd)
}

func (cli *cli) runLaunchCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	// Check that the init process found a config file
	if cli.missionFileError != nil {
		return cli.missionFileError
	}

	// Handle params
	params, err := cli.getCliParams(cmd)
	if err != nil {
		return err
	}

	// Attempt to launch mission
	return rocket.Default().
		LaunchMissionWithParams(cli.ctx, cli.missionFile,
			cli.mission.AllSettings(), params, args...)
}

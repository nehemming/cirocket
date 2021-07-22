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
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/spf13/cobra"
)

const (
	flagParam       = "param"
	flagRunbook     = "runbook"
	flagOutput      = "output"
	flagShortOutput = "o"
	flagDebug       = "debug"
	flagSilent      = "silent"
	flagMission     = "mission"
	flagConfig      = "config"
	flagWorkingDir  = "dir"
	flagOverwrite   = "replace"
)

func (cli *cli) addFlagMission(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().StringVar(&cli.missionFile, flagMission, "",
		filepath.FromSlash(fmt.Sprintf("specify a mission file (default is ./.%s)",
			cli.appName)))

	return cmd
}

func addFlagParam(cmd *cobra.Command) *cobra.Command {
	parts := strings.SplitN(cmd.Use, " ", 2)
	cmd.Flags().StringArray(flagParam, nil, fmt.Sprintf("supply parameter values to a %s, multiple params flags can be provided", parts[0]))
	return cmd
}

func (cli *cli) getCliParams(cmd *cobra.Command) ([]rocket.Param, error) {
	valueParams, err := cmd.Flags().GetStringArray(flagParam)
	if err != nil {
		return nil, err
	}

	params, err := parseParams(valueParams)
	if err != nil {
		return nil, err
	}

	// Add in special params
	if cli.debug {
		params = append(params, rocket.Param{Name: flagDebug, Value: "yes", Print: true})
	}
	if cli.silent {
		params = append(params, rocket.Param{Name: flagSilent, Value: "yes", Print: true})
	}
	return params, nil
}

func parseParams(valueParams []string) ([]rocket.Param, error) {
	params := make([]rocket.Param, len(valueParams))

	for i, nv := range valueParams {
		slice := strings.SplitN(nv, "=", 2)

		if len(slice) != 2 {
			return nil, fmt.Errorf("param[%d] %s is not formed as name=value", i, nv)
		}

		params[i] = rocket.Param{Name: slice[0], Value: slice[1], Print: true}
	}

	return params, nil
}

func addFlagRunbook(cmd *cobra.Command) *cobra.Command {
	parts := strings.SplitN(cmd.Use, " ", 2)
	cmd.Flags().String(flagRunbook, "", fmt.Sprintf("supply a runbook to %s", parts[0]))
	return cmd
}

func (cli *cli) addFlagOverwrite(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Bool(flagOverwrite, false, "overwrite existing output file")

	return cmd
}

func (cli *cli) addFlagOutput(cmd *cobra.Command) *cobra.Command {
	parts := strings.SplitN(cmd.Use, " ", 2)
	cmd.Flags().StringP(flagOutput, flagShortOutput, "", fmt.Sprintf("output location for %s", parts[0]))

	_ = cli.config.BindPFlag(flagOutput, cmd.Flags().Lookup(flagOutput))

	return cmd
}

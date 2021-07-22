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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (cli *cli) newAssemblyCommand() *cobra.Command {
	assembleCmd := &cobra.Command{
		Use:           "assemble [blueprint]",
		Short:         "assemble a mission from a blueprint \U0001F527",
		Long:          "assemble a mission from a blueprint",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          cli.runAssembleCmd,
	}

	return addFlagRunbook(addFlagParam(assembleCmd))
}

type assemblyPrep struct {
	blueprintName   string
	sources         []string
	params          []rocket.Param
	runbookLocation string
}

func (cli *cli) prepAssembleCmd(cmd *cobra.Command, args []string) (*assemblyPrep, error) {
	prep := &assemblyPrep{}

	prep.blueprintName = args[0] // cobra cmd validates there is one arg

	// Get assembly sources
	prep.sources = cli.config.GetStringSlice(configAssemblySources)

	// Handle params
	params, err := cli.getCliParams(cmd)
	if err != nil {
		return nil, err
	}
	prep.params = params

	// spec
	runbookLocation, err := cmd.Flags().GetString(flagRunbook)
	if err != nil {
		return nil, errors.Wrap(err, flagRunbook)
	}
	prep.runbookLocation = runbookLocation

	return prep, nil
}

func (cli *cli) runAssembleCmd(cmd *cobra.Command, args []string) error {
	prep, err := cli.prepAssembleCmd(cmd, args)
	if err != nil {
		return err
	}

	return rocket.Default().Assemble(cli.ctx, prep.blueprintName, prep.sources, prep.runbookLocation, prep.params)
}

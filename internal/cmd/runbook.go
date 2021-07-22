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
	"os"
	"path/filepath"
	"strings"

	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/spf13/cobra"
)

func (cli *cli) newInitRunbookCommand() *cobra.Command {
	runbookCmd := &cobra.Command{
		Use:           "runbook [blueprint]",
		Short:         "initialise a new runbook for a blueprint specification \U0001F4C3",
		Long:          `initialise a new runbook for a blueprint specification. A blueprint runbook describes all the input parameters and sequencing needed to assemble a blueprint`,
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  false,
		RunE:          cli.runInitRunbookCmd,
	}

	cli.addFlagOverwrite(runbookCmd)
	return cli.addFlagOutput(runbookCmd)
}

func (cli *cli) runInitRunbookCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	blueprintName := args[0] // cobra cmd validates there is one arg

	// Set th default path
	cli.setDefaultOutputPath(blueprintName)

	// Get assembly sources
	sources := cli.config.GetStringSlice(configAssemblySources)

	runbook, err := rocket.Default().GetRunbook(cli.ctx, blueprintName, sources)
	if err != nil {
		return err
	}

	overwrite, _ := cmd.Flags().GetBool(flagOverwrite)

	return cli.writeRunbook(runbook, overwrite)
}

func (cli *cli) setDefaultOutputPath(blueprintName string) {
	// flagOutput may have been set, but create a default as a fallback
	runbookName := fmt.Sprintf("%s_runbook.yml", blueprintName)
	cli.mission.SetDefault(flagOutput, runbookName)
}

func (cli *cli) writeRunbook(body string, overwrite bool) error {
	// Save runbook
	var err error
	runbookPath := cli.mission.GetString(flagOutput)
	if !overwrite {
		runbookPath, err = getUniqueName(runbookPath)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(runbookPath, []byte(body), 0666)
}

func getUniqueName(path string) (string, error) {
	working := path
	counter := 1
	dir, name := filepath.Split(path)
	ext := filepath.Ext(name)
	baseName := strings.TrimSuffix(name, ext)

	for {
		_, err := os.Stat(working)
		if err != nil {
			if os.IsNotExist(err) {
				// unique
				if dir != "" {
					if err := os.MkdirAll(dir, 0777); err != nil {
						return "", err
					}
				}
				return working, nil
			}
			return "", err
		}

		working = filepath.Join(dir, fmt.Sprintf("%s_%d%s", baseName, counter, ext))
		counter++
	}
}

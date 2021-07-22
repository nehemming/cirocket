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
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

//go:embed initconfig.yml
var initConfig embed.FS

func (cli *cli) newInitMissionCommand() *cobra.Command {
	missionCmd := &cobra.Command{
		Use:           "mission",
		Short:         "initialize a new mission file \U0001F58B ",
		Long:          "creates a new mission file",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  false,
		RunE:          cli.runInitMissionCmd,
	}

	return cli.addFlagMission(missionCmd)
}

func (cli *cli) runInitMissionCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	if cli.missionFileError == nil {
		// Opened an already existing file, error
		return fmt.Errorf("mission file %s already exists", cli.missionFile)
	}

	// Create a default config file
	f, err := initConfig.Open("initconfig.yml")
	if err != nil {
		// Developer build issue
		panic(err)
	}
	defer f.Close()

	if b, err := ioutil.ReadAll(f); err != nil {
		// Bad as incorrect config, developer issue
		panic(err)
	} else {
		if err := os.WriteFile(filepath.FromSlash(cli.missionFile), b, 0o666); err != nil {
			return errors.Wrap(err, "write mission file")
		}
	}

	return nil
}

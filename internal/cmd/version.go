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

	"github.com/nehemming/cirocket/pkg/buildinfo"
	"github.com/spf13/cobra"
)

func (cli *cli) newVersionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:           "version",
		Short:         "version information",
		Long:          "runs detailed version information",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          cli.runVersionCmd,
	}

	return versionCmd
}

func (cli *cli) runVersionCmd(cmd *cobra.Command, args []string) error {
	info := buildinfo.GetBuildInfo(cli.ctx)

	fmt.Println(info.TabularVersion())

	return nil
}

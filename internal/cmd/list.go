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
	"bytes"
	"fmt"
	"runtime"

	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/nehemming/yaff/cliflags"
	"github.com/nehemming/yaff/textformatter"
	"github.com/spf13/cobra"

	term "github.com/buger/goterm"
)

const listOptions = "list.opt"

func (cli *cli) newListCommand() *cobra.Command {
	listCmd := &cobra.Command{
		Use:           "list [blueprints|tasktypes]",
		Short:         "list lists resources (types: blueprints, tasktypes) \U0001F517",
		Long:          "list lists resources of the specified type.\nCurrently 'blueprints' and 'tasktypes' are supported",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  false,
		RunE:          cli.runListCmd,
	}

	cliflags.AddFormattingFlags(listCmd.Flags())
	err := cliflags.BindFormattingParamsToFlags(listCmd.Flags(), cli.config, listOptions)
	if err != nil {
		cli.configError = err
	}

	return listCmd
}

func (cli *cli) runListCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		panic("args not cli checked")
	}

	fn, err := cli.lookuplistType(args[0])
	if err == nil {
		cmd.SilenceUsage = true
		err = fn(cmd)
	}
	return err
}

func (cli *cli) lookuplistType(listType string) (func(*cobra.Command) error, error) {
	if listType == "blueprint" || listType == "blueprints" {
		return cli.listBlueprints, nil
	}

	if listType == "types" || listType == "tasktypes" {
		return cli.listTaskTypes, nil
	}
	return nil, fmt.Errorf("unknown type: %s", listType)
}

func getwidth() int {
	width := term.Width()
	if runtime.GOOS == "windows" {
		// windows seems too wide in tests
		width -= 2
	}
	return width
}

func (cli *cli) setStandardReportTypes() {
	cli.config.SetDefault(listOptions+"."+cliflags.ParamTtyWidth, getwidth())
	cli.config.SetDefault(listOptions+"."+cliflags.ParamsReportingStyle, "grid")
}

func (cli *cli) runReport(cmd *cobra.Command, getData func() (interface{}, error)) error {
	cli.setStandardReportTypes()

	formatter, options, err := cliflags.GetFormmatterFromFlags(cmd.Flags(), cli.config, textformatter.Text, listOptions)
	if err != nil {
		return err
	}

	list, err := getData()
	if err != nil {
		return err
	}

	// Generate some output
	buf := new(bytes.Buffer)

	err = formatter.Format(buf, options, list)
	if err != nil {
		return err
	}

	// display
	fmt.Println(buf.String())

	return err
}

func (cli *cli) listTaskTypes(cmd *cobra.Command) error {
	return cli.runReport(cmd, func() (interface{}, error) {
		return rocket.Default().ListTaskTypes(cli.ctx)
	})
}

func (cli *cli) listBlueprints(cmd *cobra.Command) error {
	return cli.runReport(cmd, func() (interface{}, error) {
		// Get assembly sources
		sources := cli.config.GetStringSlice(configAssemblySources)

		return rocket.Default().ListBlueprints(cli.ctx, sources)
	})
}

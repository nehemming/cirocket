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
	"context"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestNewAssemblyCommand(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())
	cmd := cli.newAssemblyCommand()

	if cmd.Use != "assemble [blueprint]" {
		t.Error("unexpected use", cmd.Use)
	}
	if !cmd.Flags().HasFlags() {
		t.Error("unexpected flags not defined")
	}

	if cmd.Flags().NFlag() != 0 {
		t.Error("unexpected flags set", cmd.Flags().NFlag())
	}
}

func TestRunAssemble(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())
	cmd := cli.newAssemblyCommand()

	// set sources here to stop config issus
	cli.config.Set(configAssemblySources, []string{"."})

	prep, err := cli.prepAssembleCmd(cmd, []string{"testrun"})
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	if prep.blueprintName != "testrun" {
		t.Error("unexpected", prep.blueprintName)
	}

	if len(prep.sources) != 1 || prep.sources[0] != "." {
		t.Errorf("sources %v", prep.sources)
	}
}

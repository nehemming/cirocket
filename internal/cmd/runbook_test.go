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
	"os"
	"path/filepath"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestNewInitRunbookCommand(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())
	cmd := cli.newInitRunbookCommand()

	if cmd.Use != "runbook [blueprint]" {
		t.Error("unexpected use", cmd.Use)
	}
	if !cmd.Flags().HasFlags() {
		t.Error("expected flags defined")
	}

	if cmd.Flags().NFlag() != 0 {
		t.Error("unexpected flags set", cmd.Flags().NFlag())
	}
}

func TestGetUniqueName(t *testing.T) {
	n, err := getUniqueName("random")
	if err != nil || n != "random" {
		t.Error("unexpected", err, n)
	}
}

func TestGetUniqueNameDeep(t *testing.T) {
	un := filepath.Join("testdata", "deep")
	n, err := getUniqueName(un)
	if err != nil || n != un {
		t.Error("unexpected", err, n)
	}
}

func TestGetUniqueNameNewName(t *testing.T) {
	n, err := getUniqueName("launch.go")
	if err != nil || n != "launch_1.go" {
		t.Error("unexpected", err, n)
	}
}

func TestRunInitRunBook(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())
	cmd := cli.newInitRunbookCommand()

	// set sources here to stop config issus
	cli.config.Set(configAssemblySources, []string{"."})

	// simulate ore run
	err := cli.preRunCheckInitErrors(cmd, []string{})
	if err != nil {
		t.Error("unexpected", err)
	}

	err = cli.runInitRunbookCmd(cmd, []string{"cli-book"})
	if err == nil || err.Error() != "blueprint cli-book: cannot find in sources" {
		t.Error("unexpected", err)
	}
}

func TestSetDefaultOutputPath(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())

	cli.setDefaultOutputPath("testing")

	runbookPath := cli.mission.GetString(flagOutput)

	if runbookPath != "testing_runbook.yml" {
		t.Error("unexpected runbookPath", runbookPath)
	}
}

func TestWriteRunbook(t *testing.T) {
	blueprint := filepath.Join("testdata", "testing")
	outFile := blueprint + "_runbook.yml"
	_ = os.Remove(outFile)
	defer func() { _ = os.Remove(outFile) }()

	cli := newCli(context.Background(), stdlog.New())
	cli.setDefaultOutputPath(blueprint)
	err := cli.writeRunbook("test body", false)
	if err != nil {
		t.Error("unexpected ", err)
	}

	_, err = os.Stat(outFile)
	if err != nil {
		t.Error("unexpected missing file", err)
	}
}

func TestWriteRunbookOverwrite(t *testing.T) {
	blueprint := filepath.Join("testdata", "testing")
	outFile := blueprint + "_runbook.yml"
	_ = os.WriteFile(outFile, []byte("hello"), 0666)
	defer func() { _ = os.Remove(outFile) }()

	cli := newCli(context.Background(), stdlog.New())
	cli.setDefaultOutputPath(blueprint)
	err := cli.writeRunbook("test body", true)
	if err != nil {
		t.Error("unexpected ", err)
	}

	_, err = os.Stat(outFile)
	if err != nil {
		t.Error("unexpected missing file", err)
	}

	_, err = os.Stat(blueprint + "_runbook_1.yml")
	if err == nil {
		t.Error("unexpected missing file")
	}
}

func TestWriteRunbookunique(t *testing.T) {
	blueprint := filepath.Join("testdata", "testing")
	outFile := blueprint + "_runbook.yml"
	unique := blueprint + "_runbook_1.yml"
	_ = os.WriteFile(outFile, []byte("hello"), 0666)
	defer func() { _ = os.Remove(outFile) }()
	defer func() { _ = os.Remove(unique) }()
	_ = os.Remove(unique)

	cli := newCli(context.Background(), stdlog.New())
	cli.setDefaultOutputPath(blueprint)
	err := cli.writeRunbook("test body", false)
	if err != nil {
		t.Error("unexpected ", err)
	}

	_, err = os.Stat(outFile)
	if err != nil {
		t.Error("unexpected missing outFile file", outFile, err)
	}

	_, err = os.Stat(unique)
	if err != nil {
		t.Error("unexpected missing unique file", unique, err)
	}
}

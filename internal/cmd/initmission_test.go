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
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestNewInitMission(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())
	cmd := cli.newInitMissionCommand()

	if cmd.Use != "mission" {
		t.Error("unexpected use", cmd.Use)
	}
	if !cmd.Flags().HasFlags() {
		t.Error("unexpected flags not defined")
	}

	if cmd.Flags().NFlag() != 0 {
		t.Error("unexpected flags set", cmd.Flags().NFlag())
	}
}

func TestRunInitMission(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())
	cmd := cli.newInitMissionCommand()
	err := os.MkdirAll("testdata", 0777)
	if err != nil {
		panic(err)
	}
	cp := filepath.Join("testdata", "mission.yml")
	_ = os.Remove(cp)
	defer func() { _ = os.Remove(cp) }()

	// create dummy error to allow init to run, no err means it exists
	cli.missionFileError = errors.New("ignore")

	cli.missionFile = cp

	err = cli.runInitMissionCmd(cmd, []string{})
	if err != nil {
		t.Error("unexpected", err)
	}

	// check exists
	_, err = os.Stat(cp)
	if err != nil {
		t.Error("unexpected missing file", err)
	}

	if !cmd.SilenceErrors {
		t.Error("not silent")
	}
}

func TestRunInitMissionFileExists(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())
	cmd := cli.newInitMissionCommand()

	cp := filepath.Join("testdata", "mission.yml")

	cli.missionFile = cp

	err := cli.runInitMissionCmd(cmd, []string{})
	if err == nil || filepath.ToSlash(err.Error()) != "mission file testdata/mission.yml already exists" {
		t.Error("unexpected", err)
	}

	if !cmd.SilenceErrors {
		t.Error("not silent")
	}
}

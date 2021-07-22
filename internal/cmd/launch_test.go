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

func TestParseParamsEmpty(t *testing.T) {
	var list []string

	r, err := parseParams(list)
	if err != nil || len(r) > 0 {
		t.Error("unexpected", err, len(r))
	}
}

func TestParseParamsSingle(t *testing.T) {
	list := []string{"abc=123"}

	r, err := parseParams(list)
	if err != nil || len(r) != 1 {
		t.Error("unexpected", err, len(r))
		return
	}

	if r[0].Name != "abc" {
		t.Error("unexpected name", r[0].Name)
	}
	if r[0].Value != "123" {
		t.Error("unexpected value", r[0].Name, r[0].Value)
	}
}

func TestParseParamsMultiple(t *testing.T) {
	list := []string{"abc=123", "def=456,7=8"}

	r, err := parseParams(list)
	if err != nil || len(r) != 2 {
		t.Error("unexpected", err, len(r))
		return
	}

	if r[0].Name != "abc" {
		t.Error("unexpected name", r[0].Name)
	}
	if r[0].Value != "123" {
		t.Error("unexpected value", r[0].Name, r[0].Value)
	}

	if r[1].Name != "def" {
		t.Error("unexpected name", r[1].Name)
	}
	if r[1].Value != "456,7=8" {
		t.Error("unexpected value", r[1].Name, r[0].Value)
	}
}

func TestNewLaunchMission(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())
	cmd := cli.newLaunchCommand()

	if cmd.Use != "launch [{flightSequence}]" {
		t.Error("unexpected use", cmd.Use)
	}
	if !cmd.Flags().HasFlags() {
		t.Error("expected flags defined")
	}

	if cmd.Flags().NFlag() != 0 {
		t.Error("unexpected flags set", cmd.Flags().NFlag())
	}
}

func TestGetCliParams(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())
	cmd := cli.newLaunchCommand()

	params, err := cli.getCliParams(cmd)
	if err != nil || len(params) != 0 {
		t.Error("unexpected", err, len(params))
	}
}

func TestGetCliParamsSilentDebug(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())
	cmd := cli.newLaunchCommand()

	cli.silent = true
	cli.debug = true

	params, err := cli.getCliParams(cmd)
	if err != nil || len(params) != 2 {
		t.Error("unexpected", err, len(params))
	}
}

func TestRunMission(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())
	cmd := cli.newLaunchCommand()

	// simulate ore run
	err := cli.preRunCheckInitErrors(cmd, []string{})
	if err != nil {
		t.Error("unexpected", err)
	}

	err = cli.runLaunchCmd(cmd, []string{})
	if err != nil {
		t.Error("unexpected", err)
	}
}

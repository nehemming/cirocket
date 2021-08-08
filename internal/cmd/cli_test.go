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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/nehemming/cirocket/pkg/buildinfo"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
	"github.com/nehemming/cirocket/pkg/resource"
)

func TestNewCli(t *testing.T) {
	ctx := context.Background()

	cli := newCli(ctx, stdlog.New())

	if cli.rootCmd == nil {
		t.Error("no root cmd")
	}

	if cli.ctx != ctx {
		t.Error("wrong context")
	}
}

func TestNewCliRootCmd(t *testing.T) {
	ctx := buildinfo.NewInfo("1.0", "", "", "", "").NewContext(context.Background())

	cli := newCli(ctx, stdlog.New())

	rootCmd := cli.rootCmd

	if rootCmd.Args == nil {
		t.Error("Args")
	}
	if rootCmd.SilenceErrors == false {
		t.Error("SilenceErrors")
	}

	if rootCmd.Version == "" {
		t.Error("Version")
	}
}

func TestInitConfigurationsBlankAppNameErrors(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())

	if cli.configError != nil {
		t.Error("profileError pre init error", cli.configError)
	}
	if cli.missionFileError != nil {
		t.Error("configError pre init error", cli.missionFileError)
	}
	cli.appName = ""

	cli.loadMissionAndConfig()
	if cli.configError == nil {
		t.Error("expected error")
	}
	if cli.missionFileError != nil {
		t.Error("configError post init error", cli.missionFileError)
	}
}

func TestLoadProfileAndConfig(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())

	if cli.configError != nil {
		t.Error("profileError pre init error", cli.configError)
	}
	if cli.missionFileError != nil {
		t.Error("configError pre init error", cli.missionFileError)
	}
	cli.appName = "notknown"
	cli.homeDir = "testdata"

	cli.loadMissionAndConfig()
	if cli.configError != nil {
		t.Error("profileError post init error", cli.configError)
	}
	if cli.missionFileError == nil || !strings.HasPrefix(cli.missionFileError.Error(), "Config File \".notknown\"") {
		t.Error("unexpected", cli.missionFileError)
	}
}

func TestRun(t *testing.T) {
	ctx := buildinfo.NewInfo("1.0", "", "", "", "").NewContext(context.Background())

	exitCode := runWithArgs(ctx, []string{"--version"}, stdlog.New())

	if exitCode != ExitCodeSuccess {
		t.Error("unexpected exit code", exitCode)
	}
}

func TestPreRunCheckInitErrorsOk(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())
	cmd := cli.newInitCommand()

	cli.missionFileError = errors.New("ignore")

	err := cli.preRunCheckInitErrors(cmd, []string{})
	if err != nil {
		t.Error("unexpected}", err)
	}
}

func TestPreRunCheckInitErrorsReturnsError(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())
	cmd := cli.newInitCommand()

	cli.configError = errors.New("ok")

	err := cli.preRunCheckInitErrors(cmd, []string{})
	if err == nil || err.Error() != "ok" {
		t.Error("unexpected}", err)
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	cli := newCli(context.Background(), stdlog.New())

	err := os.MkdirAll("testdata", 0777)
	if err != nil {
		panic(err)
	}

	// simulate init
	cli.config.SetConfigType("yml")

	cn := "test"
	cp := filepath.Join("testdata", cn+".yml")
	_ = os.Remove(cp)
	defer func() { _ = os.Remove(cp) }()

	// create default file
	cli.configFile = cp
	err = cli.createDefaultConfig()
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	// check exists
	_, err = os.Stat(cp)
	if err != nil {
		t.Error("unexpected", err)
	}

	using := resource.Relative(cli.config.ConfigFileUsed())
	if using != cp {
		t.Error("unexpected", using, cp, cn)
	}
}

func TestMultipleErrorLoggingBlank(t *testing.T) {
	loggee.SetMultiErrorFormatting(stdMultiErrorLogger(stdlog.New()))

	var err error
	for _, m := range []string{} {
		err = multierror.Append(err, fmt.Errorf("item - %s", m))
	}
}

func TestMultipleErrorLoggingOne(t *testing.T) {
	log := stdlog.New()
	loggee.SetMultiErrorFormatting(stdMultiErrorLogger(log))

	var err error
	for _, m := range []string{"red"} {
		err = multierror.Append(err, fmt.Errorf("item - %s", m))
	}

	err = loggee.BindMultiErrorFormatting(err)

	if err, ok := err.(*multierror.Error); !ok {
		t.Error("no multi error", err)
	}

	log.Warn(err.Error())
}

func TestMultipleErrorLoggingMulti(t *testing.T) {
	log := stdlog.New()
	loggee.SetMultiErrorFormatting(stdMultiErrorLogger(log))

	var err error
	for _, m := range []string{"red", "green"} {
		err = multierror.Append(err, fmt.Errorf("item - %s", m))
	}

	err = loggee.BindMultiErrorFormatting(err)

	if err, ok := err.(*multierror.Error); !ok {
		t.Error("no multi error", err)
	}

	log.Warn(err.Error())
}

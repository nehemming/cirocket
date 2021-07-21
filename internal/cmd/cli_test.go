package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/nehemming/cirocket/pkg/buildinfo"
)

func TestNewCli(t *testing.T) {
	ctx := context.Background()

	cli := newCli(ctx)

	if cli.rootCmd == nil {
		t.Error("no root cmd")
	}

	if cli.ctx != ctx {
		t.Error("wrong context")
	}
}

func TestNewCliRootCmd(t *testing.T) {
	ctx := buildinfo.NewInfo("1.0", "", "", "", "").NewContext(context.Background())

	cli := newCli(ctx)

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

func TestInitConfigBlankAppNameErrors(t *testing.T) {
	cli := newCli(context.Background())

	if cli.initError != nil {
		t.Error("pre init error")
	}

	cli.appName = ""

	cli.initConfig()
	if cli.initError == nil {
		t.Error("expected error")
	}
}

func TestInitConfig(t *testing.T) {
	cli := newCli(context.Background())

	if cli.initError != nil {
		t.Error("pre init error")
	}

	cli.appName = "notknown"

	cli.initConfig()
	if cli.initError == nil || !strings.HasPrefix(cli.initError.Error(), "Config File \".notknown.yml\"") {
		t.Error("unexpected", cli.initError)
	}
}

func TestRun(t *testing.T) {
	ctx := buildinfo.NewInfo("1.0", "", "", "", "").NewContext(context.Background())

	exitCode := runWithArgs(ctx, []string{"--version"})

	if exitCode != ExitCodeSuccess {
		t.Error("unexpected exit code", exitCode)
	}
}

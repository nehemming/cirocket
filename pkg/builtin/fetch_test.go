package builtin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
	"github.com/nehemming/cirocket/pkg/rocket"
)

func TestFetchType(t *testing.T) {
	var ft fetchType

	if ft.Type() != "fetch" {
		t.Error("Wrong fetch type", ft.Type())
	}
}

func TestFetchDesc(t *testing.T) {
	var rt fetchType

	if rt.Description() == "" {
		t.Error("needs description", rt.Type())
	}
}

func TestFetchRun(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := rocket.NewMissionControl()
	RegisterAll(mc)

	mission, cfgFile := loadMission("fetch")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err != nil {
		t.Error("failure", err)
	}

	file := filepath.Join("testdata", "readme.tmp")
	if _, err := os.Stat(file); err != nil {
		t.Error("file missing", file, err)
	}
	_ = os.Remove(file)
}

func TestFetchRunWithError(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := rocket.NewMissionControl()
	RegisterAll(mc)

	mission, cfgFile := loadMission("badfetch")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err == nil {
		t.Error("expected an error")
	}

	file := filepath.Join("testdata", "readme.tmp")
	if _, err := os.Stat(file); err == nil {
		t.Error("file present", file, err)
	}
	_ = os.Remove(file)

	file = filepath.Join("testdata", "readme2.tmp")
	_ = os.Remove(file)
}

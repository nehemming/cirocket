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

func TestMkDirType(t *testing.T) {
	var mk mkDirType

	if mk.Type() != "mkdir" {
		t.Error("Wrong mkdir type", mk.Type())
	}
}

func TestMkDirRun(t *testing.T) {

	loggee.SetLogger(stdlog.New())

	mc := rocket.NewMissionControl()
	RegisterAll(mc)

	dir := filepath.Join("testdata", "dirtest")

	_ = os.RemoveAll(dir)

	mission, cfgFile := loadMission("mkdir")

	if err := mc.FlyMission(context.Background(), cfgFile, mission); err != nil {
		t.Error("failure", err)
	}

	//Check dir exists
	if _, err := os.Stat(dir); err != nil {
		t.Error("dir missing", err)
	}
	_ = os.RemoveAll(dir)
}

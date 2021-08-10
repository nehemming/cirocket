package builtin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
	"github.com/nehemming/cirocket/pkg/rocket"
	"gopkg.in/yaml.v2"
)

func TestRunType(t *testing.T) {
	var rt runType

	if rt.Type() != "run" {
		t.Error("Wrong Run type", rt.Type())
	}
}

func TestRunDesc(t *testing.T) {
	var rt runType

	if rt.Description() == "" {
		t.Error("needs description", rt.Type())
	}
}

func loadMission(missionName string) (map[string]interface{}, string) {
	fileName := filepath.Join(".", "testdata", missionName+".yml")
	fh, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	m := make(map[string]interface{})

	err = yaml.NewDecoder(fh).Decode(&m)
	if err != nil {
		panic(err)
	}

	return m, fileName
}

func TestRunGo(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := rocket.NewMissionControl()
	RegisterAll(mc)

	mission, cfgFile := loadMission("rungo")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err != nil {
		t.Error("Run go mission failure", err)
	}
}

func TestRunGoWithError(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := rocket.NewMissionControl()
	RegisterAll(mc)

	mission, cfgFile := loadMission("rungowitherror")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err == nil {
		t.Error("Run go mission no error")
	} else if err.Error() != "stage: testing: task: run go with error: process go exit code 2" {
		t.Error("Run go mission failure unknown error", err)
	}
}

func TestRunEchoFilter(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := rocket.NewMissionControl()
	RegisterAll(mc)

	mission, cfgFile := loadMission("runecho")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err != nil {
		t.Error("Run go mission failure", err)
	}
}

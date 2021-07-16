package builtin

import (
	"context"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
	"github.com/nehemming/cirocket/pkg/rocket"
)

func TestTemplateType(t *testing.T) {
	var rt templateType

	if rt.Type() != "template" {
		t.Error("Wrong Run type", rt.Type())
	}
}

func TestTemplateHelloWorld(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := rocket.NewMissionControl()
	RegisterAll(mc)

	mission, cfgFile := loadMission("hello")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err != nil {
		t.Error("failure", err)
	}
}

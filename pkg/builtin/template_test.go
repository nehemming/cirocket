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

func TestTemplateDesc(t *testing.T) {
	var rt templateType

	if rt.Description() == "" {
		t.Error("needs description", rt.Type())
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

func TestTemplateMissingArgIsBalk(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	capComm := rocket.NewCapComm("testdata/test.yml", stdlog.New())

	ctx := context.Background()

	templateCfg := &Template{
		Template: rocket.InputSpec{
			Path: "{{.notfound}}testdata/hello.yml",
		},
	}

	if err := capComm.AttachInputSpec(ctx, templateResourceID, templateCfg.Template); err != nil {
		t.Error("unexpected", err)
		return
	}

	_, err := loadTemplate(ctx, capComm, "test", templateCfg)
	if err != nil {
		t.Error("unexpected", err)
	}
}

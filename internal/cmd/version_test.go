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

	"github.com/nehemming/cirocket/pkg/buildinfo"
	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestNewVersionCommand(t *testing.T) {
	ctx := buildinfo.NewInfo("1.0", "", "", "", "").NewContext(context.Background())

	cli := newCli(ctx, stdlog.New())

	cmd := cli.newVersionCommand()

	if cmd == nil {
		t.Error("unexpected nil cmd")
		return
	}

	if cmd.Use != "version" {
		t.Error("unexpected use", cmd.Use)
	}
}

func TestRunVersionCommand(t *testing.T) {
	ctx := buildinfo.NewInfo("1.0", "", "", "", "").NewContext(context.Background())

	cli := newCli(ctx, stdlog.New())
	if err := cli.runVersionCmd(nil, nil); err != nil {
		t.Error("unexpected err", err)
	}
}

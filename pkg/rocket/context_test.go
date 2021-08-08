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

package rocket

import (
	"context"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestContextRoundTrip(t *testing.T) {
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())

	ctx := NewContextWithCapComm(context.Background(), capComm)

	ret := GetCapCommContext(ctx)

	if ret != capComm {
		t.Error("No round trip")
	}
}

func TestGetCapCommContextWithNoneSet(t *testing.T) {
	ret := GetCapCommContext(context.Background())

	if ret != nil {
		t.Error("ret should be nil")
	}
}

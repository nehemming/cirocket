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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func TestNewMissionControl(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	mc := NewMissionControl()

	if mc == nil {
		t.Error("nil Mission Control")
	}

	impl := mc.(*missionControl)
	if impl.types == nil || len(impl.types) != 0 {
		t.Error("mission types not empty")
	}
}

type testTaskType struct {
	runCount int
	t        *testing.T
	ch       chan struct{}
}

func (tt *testTaskType) Type() string        { return "testTask" }
func (tt *testTaskType) Description() string { return "testing task" }

// Prepare the task from the input details.
func (tt *testTaskType) Prepare(ctx context.Context, capComm *CapComm, task Task) (ExecuteFunc, error) {
	if task.Name == "fail" {
		return nil, errors.
			New("failed")
	}

	var runError error

	if task.Definition["breakInRun"] != nil {
		runError = errors.New("Error in run")
	}

	return func(ctx context.Context) error {
		tt.runCount++
		fmt.Println("run:", task.Name)
		if tt.ch != nil {
			close(tt.ch)

			// Wait for done to test cancel
			<-ctx.Done()
		}

		return runError
	}, nil
}

func TestRegisterTaskTypes(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	mc := NewMissionControl()

	mc.RegisterTaskTypes()

	// Check nothin weird happened
	impl := mc.(*missionControl)
	if impl.types == nil || len(impl.types) != 0 {
		t.Error("mission types not empty")
	}

	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	if r, ok := impl.types[tt.Type()]; !ok || r != tt {
		t.Error("task type rg fail", ok, r)
	}

	if tt.runCount != 0 {
		t.Error("run count wrong", tt.runCount)
	}
}

func TestLaunchMissionZero(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	mc := NewMissionControl()

	if err := mc.LaunchMission(context.Background(), "", nil); err != nil {
		t.Error("Mission zero has a error")
	}
}

func TestLaunchMissionOne(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	mc := NewMissionControl()

	mission := make(map[string]interface{})
	mission["name"] = "one"

	if err := mc.LaunchMission(context.Background(), "", mission); err != nil {
		t.Error("Mission error", err)
	}
}

func TestLaunchMissionTwo(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	mc := NewMissionControl()

	mission := make(map[string]interface{})

	if err := mc.LaunchMission(context.Background(), "two", mission); err != nil {
		t.Error("Mission error", err)
	}
}

func loadMission(missionName string) (map[string]interface{}, string) {
	fileName := filepath.Join(".", "testdata", missionName+".yml")
	fh, err := os.Open(filepath.FromSlash(fileName))
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

func TestLaunchMissionThree(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, missionLocation := loadMission("three")

	if err := mc.LaunchMission(context.Background(), missionLocation, mission); err != nil {
		t.Error("Mission error", err)
	}

	if tt.runCount != 1 {
		t.Error("run count post mission is", tt.runCount)
	}
}

func TestLaunchMissionFour(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, missionLocation := loadMission("four")

	if err := mc.LaunchMission(context.Background(), missionLocation, mission); err == nil {
		t.Error("Nomission error, for unknown type")
	}
}

func TestLaunchMissionFive(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, missionLocation := loadMission("five")

	if err := mc.LaunchMission(context.Background(), missionLocation, mission); err == nil {
		t.Error("Nomission error, fail prepare")
	}
}

func TestLaunchMissionSix(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// test cancellation logic
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()

	ch := make(chan struct{})
	done := make(chan struct{})

	tt := &testTaskType{t: t, ch: ch}

	mc.RegisterTaskTypes(tt)

	mission, missionLocation := loadMission("six")

	go func() {
		if err := mc.LaunchMission(ctx, missionLocation, mission); err != nil {
			t.Error("Mission error", err)
		}

		if tt.runCount != 1 {
			t.Error("run count post mission is", tt.runCount)
		}

		cancel()
		close(done)
	}()

	// Wait for running
	<-ch

	// signal cancel
	cancel()

	// wait for test to complete
	<-done
}

func TestLaunchMissionSeven(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, missionLocation := loadMission("seven")

	if err := mc.LaunchMission(context.Background(), missionLocation, mission); err != nil {
		t.Error("Mission error for a try stage", err)
	}
}

func TestLaunchMissionTwelve(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, missionLocation := loadMission("twelve")

	if err := mc.LaunchMission(context.Background(), missionLocation, mission); err != nil {
		t.Error("Mission error for a ref group task", err)
	}
}

func TestLaunchMissionThirteen(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	// expect to break
	mission, missionLocation := loadMission("thirteen")

	if err := mc.LaunchMission(context.Background(), missionLocation, mission); err == nil {
		t.Error("Mission no error for a fail task")
	}
}

func TestLaunchMissionFourteen(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, missionLocation := loadMission("fourteen")

	if err := mc.LaunchMission(context.Background(), missionLocation, mission); err != nil {
		t.Error("Mission error for a ref group task", err)
	}
}

func TestLaunchMissionWithSequencesNoneSpecified(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, missionLocation := loadMission("eight")

	if err := mc.LaunchMission(context.Background(), missionLocation, mission); err != nil {
		if err.Error() != "no flight sequence specified for a configuration that uses sequences" {
			t.Error("Wrong error message", err)
		}
	} else if err == nil {
		t.Error("Expected error message")
	}
}

func TestLaunchMissionWithSequencesMissing(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, missionLocation := loadMission("eight")

	if err := mc.LaunchMission(context.Background(), missionLocation, mission, "wings"); err != nil {
		if err.Error() != "sequence wings cannot be found" {
			t.Error("Wrong error message", err)
		}
	} else if err == nil {
		t.Error("Expected error message")
	}
}

func TestLaunchMissionWithSequencesMatches(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, missionLocation := loadMission("eight")

	if err := mc.LaunchMission(context.Background(), missionLocation, mission, "run"); err != nil {
		t.Error("Unexpected error message", err)
	}
}

func TestLaunchMissionWithSequencesInclude(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, missionLocation := loadMission("nine")

	if err := mc.LaunchMission(context.Background(), missionLocation, mission, "run"); err != nil {
		t.Error("Unexpected error message", err)
	}
}

func TestProcessGlobalsPassedInParams(t *testing.T) {
	ctx := context.Background()

	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())

	mission := new(Mission)
	mission.Params = []Param{
		{Name: "mission_value", Value: "0000"},
	}

	params := []Param{
		{Name: "test", Value: "1234"},
		{Name: "mission_value", Value: "9999"},
	}

	capCommMission, err := processGlobals(ctx, capComm, mission, params)
	if err != nil {
		t.Error("unexpected")
	}

	if capCommMission.params.Get("test") != "1234" {
		t.Error("unmatched passed in params")
	}

	if capCommMission.params.Get("mission_value") != "0000" {
		t.Error("mission_value from params, should be mission")
	}
}

func TestCheckMustHaveParamsEmpty(t *testing.T) {
	paramKvg := NewKeyValueGetter(nil)

	must := MustHaveParams{}

	err := checkMustHaveParams(paramKvg, must)
	if err != nil {
		t.Error("unexpected", err)
	}
}

func TestCheckMustHaveParamsMissing(t *testing.T) {
	paramKvg := NewKeyValueGetter(nil)

	must := MustHaveParams{"red", "green"}

	err := checkMustHaveParams(paramKvg, must)
	if err == nil {
		t.Error("unexpected non error")
	}
}

func TestCheckMustHaveParamsParams(t *testing.T) {
	paramKvg := NewKeyValueGetter(nil)
	paramKvg.kv["red"] = "yes"

	must := MustHaveParams{}

	err := checkMustHaveParams(paramKvg, must)
	if err != nil {
		t.Error("unexpected", err)
	}
}

func TestCheckMustHaveParamsMatchParams(t *testing.T) {
	paramKvg := NewKeyValueGetter(nil)
	paramKvg.kv["red"] = "yes"
	paramKvg.kv["green"] = "no"

	must := MustHaveParams{"red", "green"}

	err := checkMustHaveParams(paramKvg, must)
	if err != nil {
		t.Error("unexpected", err)
	}
}

func TestPrepareTaskKindTypeDirExpandIssue(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	mc := NewMissionControl().(*missionControl)
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	task := Task{Type: tt.Type(), Dir: "{{testdata }}", Name: "exp"}

	op, err := mc.prepareTaskKindType(ctx, capComm, task)

	if err == nil || op != nil || err.Error() != "exp dir expand: parsing template: template: dir:1: function \"testdata\" not defined" {
		t.Error("unexpected", err, op)
	}
}

func TestPrepareTaskKindTypeUnknownTaskType(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	mc := NewMissionControl().(*missionControl)
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	task := Task{Type: "fred"}

	op, err := mc.prepareTaskKindType(ctx, capComm, task)

	if err == nil || op != nil || err.Error() != "unknown task type fred" {
		t.Error("unexpected", err, op)
	}
}

func TestPrepareTaskKindType(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	mc := NewMissionControl().(*missionControl)
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	task := Task{Type: tt.Type(), Dir: "testdata", OnFail: &Task{Type: tt.Type()}}

	op, err := mc.prepareTaskKindType(ctx, capComm, task)

	if err != nil || op == nil {
		t.Error("unexpected", err, op)
	}
}

func TestPrepareConcurrentTaskListExpandIssue(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	mc := NewMissionControl().(*missionControl)
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	task := Tasks{Task{Type: tt.Type()}}
	fail := &Task{Type: tt.Type()}

	ops, err := mc.prepareConcurrentTaskList(ctx, capComm, "desc", task, fail, "{{testdata }}")

	if err == nil || ops != nil || err.Error() != "desc dir expand: parsing template: template: dir:1: function \"testdata\" not defined" {
		t.Error("unexpected", err, ops)
	}
}

func TestPrepareConcurrentTaskListNoTypeError(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	mc := NewMissionControl().(*missionControl)
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	task := Tasks{Task{Type: "bad"}}
	fail := &Task{Type: tt.Type()}

	ops, err := mc.prepareConcurrentTaskList(ctx, capComm, "desc", task, fail, "testdata")

	if err == nil || ops != nil || err.Error() != "prepare: task[0]: unknown task type bad" {
		t.Error("unexpected", err, ops)
	}
}

func TestPrepareConcurrentTaskList(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	mc := NewMissionControl().(*missionControl)
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	task := Tasks{Task{Type: tt.Type()}}
	fail := &Task{Type: tt.Type()}

	op, err := mc.prepareConcurrentTaskList(ctx, capComm, "desc", task, fail, "testdata")

	if err != nil || op == nil {
		t.Error("unexpected", err, op)
	}
}

func TestPrepareConcurrentTaskListEmpty(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	mc := NewMissionControl().(*missionControl)
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	task := Tasks{}
	fail := &Task{Type: tt.Type()}

	op, err := mc.prepareConcurrentTaskList(ctx, capComm, "desc", task, fail, "testdata")

	if err != nil || op != nil {
		t.Error("unexpected", err, op)
	}
}

func TestPrepareSequentialTaskListExpandIssue(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	mc := NewMissionControl().(*missionControl)
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	task := Tasks{Task{Type: tt.Type()}}
	fail := &Task{Type: tt.Type()}

	ops, err := mc.prepareSequentialTaskList(ctx, capComm, "desc", task, fail, false, "{{testdata }}")

	if err == nil || ops != nil || err.Error() != "desc dir expand: parsing template: template: dir:1: function \"testdata\" not defined" {
		t.Error("unexpected", err, ops)
	}
}

func TestPrepareSequentialTaskListNoTypeError(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	mc := NewMissionControl().(*missionControl)
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	task := Tasks{Task{Type: "bad"}}
	fail := &Task{Type: tt.Type()}

	ops, err := mc.prepareSequentialTaskList(ctx, capComm, "desc", task, fail, false, "testdata")

	if err == nil || ops != nil || err.Error() != "prepare: task[0]: unknown task type bad" {
		t.Error("unexpected", err, ops)
	}
}

func TestPrepareSequentialTaskList(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	mc := NewMissionControl().(*missionControl)
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	task := Tasks{Task{Type: tt.Type()}}
	fail := &Task{Type: tt.Type()}

	op, err := mc.prepareSequentialTaskList(ctx, capComm, "desc", task, fail, false, "testdata")

	if err != nil || op == nil {
		t.Error("unexpected", err, op)
	}
}

func TestPrepareSequentialTaskListEmpty(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	mc := NewMissionControl().(*missionControl)
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	task := Tasks{}
	fail := &Task{Type: tt.Type()}

	op, err := mc.prepareSequentialTaskList(ctx, capComm, "desc", task, fail, false, "testdata")

	if err != nil || op != nil {
		t.Error("unexpected", err, op)
	}
}

func TestGetTaskKindNone(t *testing.T) {
	task := Task{}

	kind, err := getTaskKind(task)
	if err == nil || kind != taskKindNone {
		t.Error("unexpected", err, kind)
	}
}

func TestGetTaskKindType(t *testing.T) {
	tt := &testTaskType{t: t}
	task := Task{Type: tt.Type()}

	kind, err := getTaskKind(task)
	if err != nil || kind != taskKindType {
		t.Error("unexpected", err, kind)
	}
}

func TestGetTaskKindTry(t *testing.T) {
	task := Task{Try: Tasks{Task{}}}

	kind, err := getTaskKind(task)
	if err != nil || kind != taskKindTry {
		t.Error("unexpected", err, kind)
	}
}

func TestGetTaskKindGroup(t *testing.T) {
	task := Task{Group: Tasks{Task{}}}

	kind, err := getTaskKind(task)
	if err != nil || kind != taskKindGroup {
		t.Error("unexpected", err, kind)
	}
}

func TestGetTaskKindConcurrent(t *testing.T) {
	task := Task{Concurrent: Tasks{Task{}}}

	kind, err := getTaskKind(task)
	if err != nil || kind != taskKindConcurrent {
		t.Error("unexpected", err, kind)
	}
}

func TestGetTaskKindMultipleError(t *testing.T) {
	task := Task{Type: "bad", Concurrent: Tasks{Task{}}}

	kind, err := getTaskKind(task)
	if err == nil || kind != taskKindNone {
		t.Error("unexpected", err, kind)
	}
}

func TestApplyTaskHandlers(t *testing.T) {
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	task := Task{
		Export:    Exports{""},
		PostVars:  VarMap{"key": "val"},
		Condition: "true",
		PreVars:   VarMap{"key": "val"},
	}
	op := &operation{makeItSo: func(_ context.Context) error { return nil }}

	applyTaskHandlers(capComm, task, op)

	err := op.makeItSo(context.Background())
	if err != nil {
		t.Error("unexpected", err)
	}
}

func TestConvertStagesToMapDup(t *testing.T) {
	sm, err := convertStagesToMap(Stages{Stage{Name: "dup"}, Stage{Name: "dup"}})
	if err == nil || len(sm) > 0 {
		t.Error("unexpected", err, sm)
	}
}

func TestConvertStagesToMapIgnoreNoName(t *testing.T) {
	sm, err := convertStagesToMap(Stages{Stage{Name: ""}, Stage{Name: "one"}})
	if err != nil || len(sm) != 1 {
		t.Error("unexpected", err, sm)
	}
	if _, ok := sm[""]; ok {
		t.Error("unexpected empotty name", sm)
	}
}

func TestGetStagesTooRunNoSequences(t *testing.T) {
	stages, err := getStagesTooRun(&Mission{}, StageMap{}, []string{"one"})
	if err == nil || len(stages) > 0 {
		t.Error("unexpected", err, stages)
	}
}

func TestGetStagesTooRunNoStage(t *testing.T) {
	stages, err := getStagesTooRun(&Mission{Sequences: map[string][]string{"one": {"notthere"}}}, StageMap{}, []string{"one"})
	if err == nil || len(stages) > 0 {
		t.Error("unexpected", err, stages)
	}
}

func TestApplyConditionHandler(t *testing.T) {
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())

	op := &operation{makeItSo: func(_ context.Context) error { return nil }}

	task := Task{
		Condition: "{{ bad",
	}

	applyConditionHandler(capComm, task, op)

	err := op.makeItSo(context.Background())

	if err == nil || err.Error() != "parsing template: template: condition:1: function \"bad\" not defined" {
		t.Error("unexpected", err)
	}
}

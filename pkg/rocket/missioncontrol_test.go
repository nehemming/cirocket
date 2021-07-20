package rocket

import (
	"context"
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

func (tt *testTaskType) Type() string { return "testTask" }

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

func TestLaunchMissionThree(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, cfgFile := loadMission("three")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err != nil {
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

	mission, cfgFile := loadMission("four")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err == nil {
		t.Error("Nomission error, for unknown type")
	}
}

func TestLaunchMissionFive(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, cfgFile := loadMission("five")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err == nil {
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

	mission, cfgFile := loadMission("six")

	go func() {
		if err := mc.LaunchMission(ctx, cfgFile, mission); err != nil {
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

	mission, cfgFile := loadMission("seven")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err != nil {
		t.Error("Mission error for a try stage", err)
	}
}

func TestLaunchMissionWithSequencesNoneSpecified(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, cfgFile := loadMission("eight")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err != nil {
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

	mission, cfgFile := loadMission("eight")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission, "wings"); err != nil {
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

	mission, cfgFile := loadMission("eight")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission, "run"); err != nil {
		t.Error("Unexpected error message", err)
	}
}

func TestLaunchMissionWithSequencesInclude(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := NewMissionControl()
	tt := &testTaskType{t: t}

	mc.RegisterTaskTypes(tt)

	mission, cfgFile := loadMission("nine")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission, "run"); err != nil {
		t.Error("Unexpected error message", err)
	}
}

func TestProcessGlobalsPassedInParams(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(testConfigFile, stdlog.New())

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

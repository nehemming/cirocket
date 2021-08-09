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
	"sync"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/resource"
	"github.com/pkg/errors"
)

var (
	// once gates the singleton Default mission control.
	once sync.Once

	// defaultControl is the Default singleton mission control.
	defaultControl MissionController
)

type (

	// Option is the interface supported by all mission options.
	Option interface {
		Name() string
	}

	// ExecuteFunc is the function signature of an activity that can be executed.
	ExecuteFunc = loggee.ActivityFunc

	// TaskType represents a specific task type.
	TaskType interface {

		// Type of the task
		Type() string

		// Prepare the task from the input details.
		Prepare(ctx context.Context, capComm *CapComm, task Task) (ExecuteFunc, error)
	}

	// MissionController seeks out new civilizations in te CI space.
	MissionController interface {
		// Set option sets options on the mission controller
		SetOptions(options ...Option) error

		// RegisterTaskTypes we only want the best.
		RegisterTaskTypes(types ...TaskType)

		// LaunchMission loads and executes the mission
		// flightSequences may be specified, each sequence is run in the order specified.
		// Location is used to indicate where the config was read from, if blank the current working directory is assumed.
		LaunchMission(ctx context.Context, location string, spaceDust map[string]interface{}, flightSequences ...string) error

		// LaunchMissionWithParams loads and executes the mission.
		// Params can be supplied to the mssion.  The params are loaded before the mission parameters, as such
		// any value defined in the mission will override the params passed here.
		// flightSequences may be specified, each sequence is run in the order specified.
		// Location is used to indicate where the config was read from, if blank the current working directory is assumed.

		LaunchMissionWithParams(ctx context.Context, location string,
			spaceDust map[string]interface{},
			params []Param,
			flightSequences ...string) error

		// Assemble locates a blueprint from the assembly sources, loads the runbook and builds the assembly following the runbook.
		Assemble(ctx context.Context, blueprint string, sources []string, specLocation string, params []Param) error

		// GetRunbook gets the runbook for a blueprint
		GetRunbook(ctx context.Context, blueprintName string, sources []string) (string, error)

		// ListBlueprints builds a list of all blueprints found in the passed sources
		ListBlueprints(ctx context.Context, sources []string) ([]BlueprintInfo, error)
	}

	// operations is a collection of operations.
	operations []*operation

	// operation represents an activity to execute.
	operation struct {
		description string
		makeItSo    ExecuteFunc
		try         bool
	}
)

// missionControl implements MissionControl.
type missionControl struct {
	lock  sync.Mutex
	types map[string]TaskType
	log   loggee.Logger
}

// NewMissionControl create a new mission control.
func NewMissionControl() MissionController {
	return &missionControl{
		types: make(map[string]TaskType),
	}
}

// Default returns the default shared mission control.
func Default() MissionController {
	once.Do(func() { defaultControl = NewMissionControl() })
	return defaultControl
}

func (mc *missionControl) missionLog() loggee.Logger {
	if mc.log == nil {
		mc.lock.Lock()
		defer mc.lock.Unlock()
		if mc.log == nil {
			// fallback to default log
			mc.log = loggee.Default()
		}
	}

	return mc.log
}

// RegisterActorTypes actor types.
func (mc *missionControl) RegisterTaskTypes(types ...TaskType) {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	for _, t := range types {
		mc.types[t.Type()] = t
	}
}

func (mc *missionControl) LaunchMission(ctx context.Context, location string, spaceDust map[string]interface{}, flightSequences ...string) error {
	return mc.LaunchMissionWithParams(ctx, location, spaceDust, nil, flightSequences...)
}

func (mc *missionControl) LaunchMissionWithParams(ctx context.Context, location string,
	spaceDust map[string]interface{}, params []Param,
	flightSequences ...string) error {
	missionURL, err := getStartingMissionURL(location)
	if err != nil {
		return err
	}

	// Load the mission
	mission, err := loadPreMission(ctx, spaceDust, missionURL)
	if err != nil {
		return err
	}

	// Create a cap comm object from the environment
	capComm := newCapCommFromEnvironment(missionURL, mc.missionLog())

	// Check for missing params
	if err := checkMustHaveParams(capComm.params, mission.Must); err != nil {
		return err
	}

	// Misssion has been successfully parsed, load the global settings
	capComm, err = processGlobals(ctx, capComm, mission, params)
	if err != nil {
		return errors.Wrap(err, "global settings failure")
	}

	// get the stages needed to be run
	stagesToRun, err := getStagesTooRun(mission, flightSequences)
	if err != nil {
		return err
	}

	// prepare the stages
	operations, err := mc.prepareStages(ctx, capComm, stagesToRun)
	if err != nil {
		return err
	}

	return runOperations(ctx, operations)
}

func (mc *missionControl) prepareStages(ctx context.Context, capComm *CapComm, stagesToRun []Stage) (operations, error) {
	operations := make(operations, 0)

	// prepare stages
	for index, stage := range stagesToRun {
		if stage.Name == "" {
			stage.Name = fmt.Sprintf("stage[%d]", index)
		}

		// prepare the stage
		ops, err := mc.prepareStage(ctx, capComm, stage)
		if err != nil {
			return nil, errors.Wrapf(err, "%s prepare", stage.Name)
		}

		var dir string
		if stage.Dir != "" {
			dir, err = capComm.ExpandString(ctx, "dir", stage.Dir)
			if err != nil {
				return nil, errors.Wrapf(err, "%s dir expand", stage.Name)
			}
		}

		if len(ops) > 0 {
			operations = append(operations, &operation{
				description: fmt.Sprintf("stage: %s", stage.Name),
				makeItSo:    engage(ctx, ops, dir),
				try:         stage.Try,
			})
		}
	}

	return operations, nil
}

func runOperations(ctx context.Context, operations operations) error {
	//	Run mission
	for _, op := range operations {
		if ctx.Err() != nil {
			return nil
		}

		if err := runOp(ctx, op); err != nil {
			return err
		}
	}

	return nil
}

func checkMustHaveParams(params Getter, must MustHaveParams) error {
	var err error
	for _, m := range must {
		if params.Get(m) == "" {
			err = multierror.Append(err, fmt.Errorf("param %s must bet set to a non blank value", m))
		}
	}

	return loggee.BindMultiErrorFormatting(err)
}

func (mc *missionControl) prepareStage(ctx context.Context, missionCapComm *CapComm, stage Stage) (operations, error) {
	if err := checkMustHaveParams(missionCapComm.params, stage.Must); err != nil {
		return nil, err
	}

	operations := make(operations, 0, 10)

	// Create a new CapComm for the stage
	capComm := missionCapComm.Copy(stage.NoTrust).
		MergeBasicEnvMap(stage.BasicEnv)

	if stage.Filter.IsFiltered() {
		return operations, nil
	}

	if err := capComm.MergeParams(ctx, stage.Params); err != nil {
		return nil, errors.Wrap(err, "merging params")
	}

	// Merge and expand template envs
	if err := capComm.MergeTemplateEnvs(ctx, stage.Env); err != nil {
		return nil, errors.Wrap(err, "merging template envs")
	}

	capComm.Seal()

	// Move onto tasks
	for index, task := range stage.Tasks {
		if task.Name == "" {
			task.Name = fmt.Sprintf("task[%d]", index)
		}

		// prepare the task
		op, err := mc.prepareTask(ctx, capComm, task)
		if err != nil {
			return nil, errors.Wrapf(err, "prepare: %s", task.Name)
		}

		if op != nil {
			operations = append(operations, op)
		}
	}

	return operations, nil
}

func (mc *missionControl) prepareTask(ctx context.Context, stageCapComm *CapComm, task Task) (*operation, error) {
	if err := checkMustHaveParams(stageCapComm.params, task.Must); err != nil {
		return nil, err
	}

	// Create a new CapComm for the task
	capComm := stageCapComm.Copy(task.NoTrust).
		MergeBasicEnvMap(task.BasicEnv)

	if task.Filter.IsFiltered() {
		return nil, nil
	}

	// Merge the parameters
	if err := capComm.MergeParams(ctx, task.Params); err != nil {
		return nil, errors.Wrap(err, "merging params")
	}

	// Merge and expand template envs
	if err := capComm.MergeTemplateEnvs(ctx, task.Env); err != nil {
		return nil, errors.Wrap(err, "merging template envs")
	}

	// Look up task
	tt, ok := mc.types[task.Type]
	if !ok {
		// Unknown task type
		return nil, fmt.Errorf("unknown task type %s", task.Type)
	}

	taskFunc, err := tt.Prepare(ctx, capComm.Seal(), task)
	if err != nil {
		return nil, err
	}

	if taskFunc == nil {
		return nil, nil
	}

	var dir string
	if task.Dir != "" {
		dir, err = capComm.ExpandString(ctx, "dir", task.Dir)
		if err != nil {
			return nil, errors.Wrapf(err, "%s dir expand", task.Name)
		}
	}

	return &operation{
		description: fmt.Sprintf("task: %s", task.Name),
		makeItSo:    taskSwapDir(dir, taskFunc),
		try:         task.Try,
	}, nil
}

func taskSwapDir(dir string, fn ExecuteFunc) ExecuteFunc {
	if dir == "" {
		return fn
	}
	return func(ctx context.Context) error {
		pop, err := swapDir(dir)
		if err != nil {
			return err
		}
		defer pop()

		return fn(ctx)
	}
}

func processGlobals(ctx context.Context, capComm *CapComm, mission *Mission, suppliedParams []Param) (*CapComm, error) {
	// Copy the inbound CapComm
	capComm = capComm.Copy(false).
		WithMission(mission).
		MergeBasicEnvMap(mission.BasicEnv).
		AddAdditionalMissionData(mission.Additional)

	// Merge and expand parameters
	if err := capComm.MergeParams(ctx, suppliedParams); err != nil {
		return nil, errors.Wrap(err, "merging supplied params")
	}

	// Merge and expand parameters
	if err := capComm.MergeParams(ctx, mission.Params); err != nil {
		return nil, errors.Wrap(err, "merging params")
	}

	// Merge and expand template envs
	if err := capComm.MergeTemplateEnvs(ctx, mission.Env); err != nil {
		return nil, errors.Wrap(err, "merging template envs")
	}

	// Return a sealed capComm that cannot be edited
	return capComm.Seal(), nil
}

func getStagesTooRun(mission *Mission, flightSequences []string) ([]Stage, error) {
	stageMap, err := convertStagesToMap(mission.Stages)
	if err != nil {
		return nil, err
	}

	if len(mission.Sequences) == 0 {
		// using stages alone
		if len(flightSequences) > 0 {
			return nil, errors.New("flight sequence specified for a configuration does not use sequences")
		}

		// Just run all stages in order
		return mission.Stages, nil
	}

	//	Using sequences
	if len(flightSequences) == 0 {
		return nil, errors.New("no flight sequence specified for a configuration that uses sequences")
	}

	stagesToRun := make([]Stage, 0)

	// Compile stagesToRun from the sequences specified
	alreadySpecified := make(map[string]bool)
	for _, flight := range flightSequences {
		sequence, ok := mission.Sequences[flight]
		if !ok {
			return nil, fmt.Errorf("sequence %s cannot be found", flight)
		}
		for _, stageName := range sequence {
			if stage, ok := stageMap[stageName]; !ok {
				return nil, fmt.Errorf("sequence %s cannot find stage: %s", flight, stageName)
			} else if !alreadySpecified[stageName] {
				alreadySpecified[stageName] = true
				stagesToRun = append(stagesToRun, stage)
			}
		}
	}

	return stagesToRun, nil
}

func convertStagesToMap(stages []Stage) (map[string]Stage, error) {
	m := make(map[string]Stage)

	// prepare stages
	for index, stage := range stages {
		if stage.Name == "" {
			stage.Name = fmt.Sprintf("stage[%d]", index)
		}

		if _, ok := m[stage.Name]; ok {
			return nil, fmt.Errorf("stage: %s name is duplicated", stage.Name)
		}

		m[stage.Name] = stage
	}

	return m, nil
}

func runOp(ctx context.Context, op *operation) error {
	loggee.Info(op.description)

	if err := loggee.Activity(ctx, op.makeItSo); err != nil {
		if op.try {
			loggee.Warnf("try failed: %s", errors.Wrap(err, op.description))
		} else {
			return errors.Wrap(err, op.description)
		}
	}

	return nil
}

// swapDir changes to the new directory and resurns a function to resore the current dir, or the functionreturns an error.
// If the restore function fails to restor the working dir it will panic.
func swapDir(dir string) (func(), error) {
	// return no op if no dir change requested
	if dir == "" {
		return func() {}, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// decode any urls
	u, err := resource.UltimateURL(dir)
	if err != nil {
		return nil, err
	}
	dir, err = resource.URLToPath(u)
	if err != nil {
		return nil, err
	}

	err = os.Chdir(dir)
	if err != nil {
		return nil, err
	}

	return func() {
		if e := os.Chdir(cwd); e != nil {
			panic(e)
		}
	}, nil
}

func engage(_ context.Context, ops operations, dir string) ExecuteFunc {
	return func(ctx context.Context) error {
		pop, err := swapDir(dir)
		if err != nil {
			return err
		}
		defer pop()

		for _, op := range ops {
			if ctx.Err() != nil {
				return nil
			}

			if err := loggee.Activity(ctx, func(ctx context.Context) error {
				return runOp(ctx, op)
			}); err != nil {
				return err
			}
		}

		return nil
	}
}

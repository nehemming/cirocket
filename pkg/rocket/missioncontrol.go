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

type taskKind int

const (
	taskKindNone = taskKind(iota)
	taskKindType
	taskKindTry
	taskKindGroup
)

type (

	// Option is the interface supported by all mission options.
	Option interface {
		Name() string
	}

	// ExecuteFunc is the function signature of an activity that can be executed.
	ExecuteFunc = loggee.ActivityFunc

	// StageMap is a map of stage names to stages.
	StageMap map[string]Stage

	// TaskMap maps the task name to the task.
	TaskMap map[string]Task

	// TaskType represents a specific task type.
	TaskType interface {

		// Type of the task.
		Type() string

		// Description is a free text description of the type
		Description() string

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
			params Params,
			flightSequences ...string) error

		// Assemble locates a blueprint from the assembly sources, loads the runbook and builds the assembly following the runbook.
		Assemble(ctx context.Context, blueprint string, sources []string, specLocation string, params Params) error

		// GetRunbook gets the runbook for a blueprint
		GetRunbook(ctx context.Context, blueprintName string, sources []string) (string, error)

		// ListBlueprints builds a list of all blueprints found in the passed sources
		ListBlueprints(ctx context.Context, sources []string) ([]BlueprintInfo, error)

		ListTaskTypes(ctx context.Context) (TaskTypeInfoList, error)
	}

	// operations is a collection of operations.
	operations []*operation

	// operation represents an activity to execute.
	operation struct {
		description string
		makeItSo    ExecuteFunc
		try         bool
		onFail      ExecuteFunc
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
	spaceDust map[string]interface{}, params Params,
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

	// Create a map of staage names to stages, used for ref lookups and flight sequences
	stageMap, err := convertStagesToMap(mission.Stages)
	if err != nil {
		return err
	}

	// get the stages needed to be run
	stagesToRun, err := getStagesTooRun(mission, stageMap, flightSequences)
	if err != nil {
		return err
	}

	// prepare the stages
	operations, err := mc.prepareStages(ctx, capComm, stageMap, stagesToRun)
	if err != nil {
		return err
	}

	var fallbackOp *operation
	if mission.OnFail != nil {
		fallbackOp, err = mc.prepareFailStage(ctx, capComm, stageMap, *mission.OnFail)
		if err != nil {
			return err
		}
	}

	return runOperations(ctx, operations, fallbackOp)
}

func mergeStages(stage *Stage, ref string, stageMap StageMap, circular map[string]bool) error {
	if circular[ref] {
		return fmt.Errorf("circular ref %s", ref)
	}
	circular[ref] = true

	src, ok := stageMap[ref]
	if !ok {
		return fmt.Errorf("unknown stage ref %s", ref)
	}

	// apply items from src if not set
	if stage.Description == "" {
		stage.Description = src.Description
	}

	if stage.Dir == "" {
		stage.Dir = src.Dir
	}

	if len(stage.BasicEnv) == 0 {
		stage.BasicEnv = src.BasicEnv.Copy()
	}

	if len(stage.Env) == 0 {
		stage.Env = src.Env.Copy()
	}

	if stage.Filter == nil {
		stage.Filter = src.Filter
	}

	if len(stage.Must) == 0 {
		stage.Must = src.Must.Copy()
	}

	if !stage.NoTrust {
		stage.NoTrust = src.NoTrust
	}

	if stage.OnFail == nil && src.OnFail != nil {
		c := *src.OnFail
		stage.OnFail = &c
	}

	if len(stage.Params) == 0 {
		stage.Params = src.Params.Copy()
	}

	if len(stage.Tasks) == 0 {
		stage.Tasks = src.Tasks.Copy()
	}

	if src.Ref != "" {
		return mergeStages(stage, src.Ref, stageMap, circular)
	}

	return nil
}

func mergeStageRef(stage *Stage, stageMap StageMap) error {
	circular := map[string]bool{stage.Name: true}

	return mergeStages(stage, stage.Ref, stageMap, circular)
}

func mergeDefinition(dest, src map[string]interface{}) {
	for k, v := range src {
		if _, ok := dest[k]; !ok {
			dest[k] = v // be aware this is an alias not a copy, used for decoding
		}
	}
}

func mergeTasks(task *Task, ref string, taskMap TaskMap, circular map[string]bool) error {
	if circular[ref] {
		return fmt.Errorf("circular ref %s", ref)
	}
	circular[ref] = true

	src, ok := taskMap[ref]
	if !ok {
		return fmt.Errorf("unknown task ref %s", ref)
	}

	// apply items from src if not set
	if task.Description == "" {
		task.Description = src.Description
	}

	if task.Type == "" {
		task.Type = src.Type
	}

	if len(task.Export) == 0 {
		task.Export = src.Export.Copy()
	}

	if task.Dir == "" {
		task.Dir = src.Dir
	}

	if len(task.Try) == 0 {
		task.Try = src.Try.Copy()
	}

	if len(task.Group) == 0 {
		task.Group = src.Group.Copy()
	}

	if len(task.BasicEnv) == 0 {
		task.BasicEnv = src.BasicEnv.Copy()
	}

	if len(task.Env) == 0 {
		task.Env = src.Env.Copy()
	}

	if task.Filter == nil {
		task.Filter = src.Filter
	}

	if len(task.Must) == 0 {
		task.Must = src.Must.Copy()
	}

	if !task.NoTrust {
		task.NoTrust = src.NoTrust
	}

	if len(task.Params) == 0 {
		task.Params = src.Params.Copy()
	}

	if task.OnFail == nil && src.OnFail != nil {
		c := *src.OnFail
		task.OnFail = &c
	}

	mergeDefinition(task.Definition, src.Definition)

	if src.Ref != "" {
		return mergeTasks(task, src.Ref, taskMap, circular)
	}

	return nil
}

func mergeTaskRef(task *Task, taskMap TaskMap) error {
	circular := map[string]bool{task.Name: true}

	return mergeTasks(task, task.Ref, taskMap, circular)
}

func (mc *missionControl) prepareStages(ctx context.Context, capComm *CapComm, stageMap StageMap, stagesToRun Stages) (operations, error) {
	operations := make(operations, 0)

	// prepare stages
	for index, stage := range stagesToRun {
		if stage.Name == "" {
			stage.Name = fmt.Sprintf("stage[%d]", index)
		}

		// check stage to see if it has a reference to another stage
		if stage.Ref != "" {
			if err := mergeStageRef(&stage, stageMap); err != nil {
				return nil, errors.Wrapf(err, "%s merge with ref %s", stage.Name, stage.Ref)
			}
		}

		// prepare the stage
		op, err := mc.prepareStage(ctx, capComm, stage)
		if err != nil {
			return nil, errors.Wrapf(err, "%s prepare", stage.Name)
		}

		if op != nil {
			operations = append(operations, op)
		}
	}

	return operations, nil
}

func runOperations(ctx context.Context, operations operations, onFailStage *operation) error {
	//	Run mission
	for _, op := range operations {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Handle mission failure
		if err := runOp(ctx, op); err != nil {
			if onFailStage != nil {
				runOnFail(ctx, onFailStage.makeItSo, onFailStage.description)
			}
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

func (mc *missionControl) prepareFailStage(ctx context.Context, capComm *CapComm, stageMap StageMap, stage Stage) (*operation, error) {
	if stage.Name == "" {
		stage.Name = "onfail"
	}

	// check stage to see if it has a reference to another stage
	if stage.Ref != "" {
		if err := mergeStageRef(&stage, stageMap); err != nil {
			return nil, errors.Wrapf(err, "%s merge with ref %s", stage.Name, stage.Ref)
		}
	}

	// prepare the stage
	op, err := mc.prepareStage(ctx, capComm, stage)
	if err != nil {
		return nil, errors.Wrapf(err, "%s prepare", stage.Name)
	}
	return op, nil
}

func createStageCapComm(ctx context.Context, missionCapComm *CapComm, stage Stage) (*CapComm, error) {
	// Create a new CapComm for the stage
	capComm := missionCapComm.Copy(stage.NoTrust).
		MergeBasicEnvMap(stage.BasicEnv)

	if err := capComm.MergeParams(ctx, stage.Params); err != nil {
		return nil, errors.Wrap(err, "merging params")
	}

	// Merge and expand template envs
	if err := capComm.MergeTemplateEnvs(ctx, stage.Env); err != nil {
		return nil, errors.Wrap(err, "merging template envs")
	}

	capComm.Seal()

	return capComm, nil
}

func (mc *missionControl) prepareStage(ctx context.Context, missionCapComm *CapComm, stage Stage) (*operation, error) {
	if stage.Filter.IsFiltered() {
		return nil, nil
	}

	if err := checkMustHaveParams(missionCapComm.params, stage.Must); err != nil {
		return nil, err
	}

	// Create the cap comm for the stage
	capComm, err := createStageCapComm(ctx, missionCapComm, stage)
	if err != nil {
		return nil, err
	}

	return mc.prepareTaskList(ctx, capComm, "stage: "+stage.Name, stage.Tasks, stage.OnFail, false, stage.Dir)
}

func (mc *missionControl) prepareFailTask(ctx context.Context, capComm *CapComm, task Task) (*operation, error) {
	if task.Name == "" {
		task.Name = "onfail"
	}

	// prepare the task
	op, err := mc.prepareTask(ctx, capComm, task)
	if err != nil {
		return nil, errors.Wrapf(err, "prepare: %s", task.Name)
	}

	return op, nil
}

func taskCapComm(ctx context.Context, parentCapComm *CapComm, task Task) (*CapComm, error) {
	// Create a new CapComm for the task
	capComm := parentCapComm.Copy(task.NoTrust).
		MergeBasicEnvMap(task.BasicEnv)

	// Merge the parameters
	if err := capComm.MergeParams(ctx, task.Params); err != nil {
		return nil, errors.Wrap(err, "merging params")
	}

	// Merge and expand template envs
	if err := capComm.MergeTemplateEnvs(ctx, task.Env); err != nil {
		return nil, errors.Wrap(err, "merging template envs")
	}

	return capComm, nil
}

func (mc *missionControl) prepareTask(ctx context.Context, parentCapComm *CapComm, task Task) (*operation, error) {
	if task.Filter.IsFiltered() {
		return nil, nil
	}

	if err := checkMustHaveParams(parentCapComm.params, task.Must); err != nil {
		return nil, err
	}

	capComm, err := taskCapComm(ctx, parentCapComm, task)
	if err != nil {
		return nil, err
	}

	// determin task kind
	taskKind, err := getTaskKind(task)
	if err != nil {
		return nil, err
	}

	var op *operation

	switch taskKind {
	case taskKindType:
		return mc.prepareTaskKindType(ctx, capComm, task)
	case taskKindTry:
		op, err = mc.prepareTaskList(ctx, capComm, "try: "+task.Name, task.Try, task.OnFail, true, task.Dir)
	case taskKindGroup:
		op, err = mc.prepareTaskList(ctx, capComm, "group: "+task.Name, task.Group, task.OnFail, false, task.Dir)
	}

	if err != nil {
		return nil, err
	}

	if op != nil && len(task.Export) > 0 {
		// add in the exports
		op.AddHandler(func(next ExecuteFunc) ExecuteFunc {
			return func(opCtx context.Context) error {
				err := next(opCtx)
				if err == nil {
					capComm.ExportVariables(task.Export)
				}
				return err
			}
		})
	}

	return op, nil
}

func getTaskKind(task Task) (taskKind, error) {
	kinds := 0

	if task.Type != "" {
		kinds++
	}
	if len(task.Try) > 0 {
		kinds++
	}
	if len(task.Group) > 0 {
		kinds++
	}

	if kinds == 0 {
		return taskKindNone, fmt.Errorf("task %s kind not specificed needs to be a type, group or try task", task.Name)
	}

	if kinds > 1 {
		return taskKindNone, fmt.Errorf("task %s has multiple kinds needs to be a single type, group or try task", task.Name)
	}
	if task.Type != "" {
		return taskKindType, nil
	}
	if len(task.Try) > 0 {
		return taskKindTry, nil
	}
	if len(task.Group) > 0 {
		return taskKindGroup, nil
	}
	panic("kind?")
}

func (mc *missionControl) prepareTaskList(ctx context.Context, capComm *CapComm,
	groupDesc string, tasks Tasks, onFailTask *Task,
	tryOp bool, taskDir string) (*operation, error) {
	taskMap := tasks.ToMap()
	var operations operations

	// Move onto tasks
	for index, task := range tasks {
		if task.Name == "" {
			task.Name = fmt.Sprintf("task[%d]", index)
		}

		// merge tasks if ref'd
		if task.Ref != "" {
			if err := mergeTaskRef(&task, taskMap); err != nil {
				return nil, errors.Wrapf(err, "%s merge with ref %s", task.Name, task.Ref)
			}
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

	// Is there ar failure task?
	var onFail ExecuteFunc
	if onFailTask != nil {
		onFailOp, err := mc.prepareFailTask(ctx, capComm, *onFailTask)
		if err != nil {
			return nil, err
		}

		onFail = onFailOp.makeItSo
	}

	if len(operations) == 0 {
		return nil, nil
	}

	// handle any dir change
	var dir string
	var err error
	if taskDir != "" {
		dir, err = capComm.ExpandString(ctx, "dir", taskDir)
		if err != nil {
			return nil, errors.Wrapf(err, "%s dir expand", groupDesc)
		}
	}

	return &operation{
		description: groupDesc,
		makeItSo:    engage(ctx, operations, dir),
		try:         tryOp,
		onFail:      onFail,
	}, nil
}

func (mc *missionControl) prepareTaskKindType(ctx context.Context, capComm *CapComm, task Task) (*operation, error) {
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

	// Is there ar failure task?
	var onFail ExecuteFunc
	if task.OnFail != nil {
		onFailOp, err := mc.prepareFailTask(ctx, capComm, *task.OnFail)
		if err != nil {
			return nil, err
		}

		onFail = onFailOp.makeItSo
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
		try:         false,
		onFail:      onFail,
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

func processGlobals(ctx context.Context, capComm *CapComm, mission *Mission, suppliedParams Params) (*CapComm, error) {
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

func getStagesTooRun(mission *Mission, stageMap StageMap, flightSequences []string) (Stages, error) {
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

	var stagesToRun Stages

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

func convertStagesToMap(stages Stages) (StageMap, error) {
	m := make(StageMap)

	// prepare stages
	for _, stage := range stages {
		// ignore un named stages
		if stage.Name == "" {
			continue
		}

		if _, ok := m[stage.Name]; ok {
			return nil, fmt.Errorf("stage: %s name is duplicated", stage.Name)
		}

		m[stage.Name] = stage
	}

	return m, nil
}

func runOnFail(ctx context.Context, action ExecuteFunc, description string) {
	// don't pass cancel context into fail action.
	fn := func(_ context.Context) error {
		return action(context.Background())
	}

	// Invoke failure fallback
	if err := loggee.Activity(ctx, fn); err != nil {
		loggee.Errorf("fail action failed: %s", errors.Wrap(err, description))
	}
}

func runOp(ctx context.Context, op *operation) error {
	loggee.Info(op.description)

	if err := loggee.Activity(ctx, op.makeItSo); err != nil {
		if op.try {
			loggee.Warnf("try failed: %s", errors.Wrap(err, op.description))
		} else {
			if op.onFail != nil {
				runOnFail(ctx, op.onFail, op.description)
			}

			// report original error
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

// OperationHandler is a handler function that given a operations next function returns the function replacing it.
type OperationHandler func(next ExecuteFunc) ExecuteFunc

func (op *operation) AddHandler(handler OperationHandler) {
	op.makeItSo = handler(op.makeItSo)
}

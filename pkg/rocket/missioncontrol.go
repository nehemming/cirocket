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
	"sync"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/nehemming/cirocket/pkg/loggee"
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
	taskKindConcurrent
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
		Assemble(ctx context.Context, blueprint string, sources []string, runbook string, params Params) error

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

	return engage(ctx, operations, fallbackOp, capComm.Log())
}

func mergeStages(stage *Stage, ref string, stageMap StageMap, circular map[string]bool) error { // nolint:cyclop
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

func mergeTasks(task *Task, ref string, taskMap TaskMap, circular map[string]bool) error { //nolint:cyclop
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

	if len(task.Try) == 0 {
		task.Try = src.Try.Copy()
	}

	if len(task.Group) == 0 {
		task.Group = src.Group.Copy()
	}

	if len(task.Concurrent) == 0 {
		task.Concurrent = src.Concurrent.Copy()
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

	if len(task.PreVars) == 0 {
		task.PreVars = src.PreVars.Copy()
	}

	if len(task.PostVars) == 0 {
		task.PostVars = src.PostVars.Copy()
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

	op, err := mc.prepareSequentialTaskList(ctx, capComm, "stage: "+stage.Name, stage.Tasks, stage.OnFail, false, stage.Dir)
	if err != nil {
		return nil, err
	}

	if stage.If != "" {
		return op.AddHandler(func(next ExecuteFunc) ExecuteFunc {
			return func(execCtx context.Context) error {
				ok, err := capComm.ExpandBool(ctx, "if", stage.If)
				if err != nil {
					return err
				}

				if !ok {
					return nil
				}

				return next(execCtx)
			}
		}), nil
	}

	return op, nil
}

func (mc *missionControl) prepareFailTask(ctx context.Context, capComm *CapComm, task Task, taskMap TaskMap) (*operation, error) {
	if task.Name == "" {
		task.Name = "onfail"
	}

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

func (mc *missionControl) switchTaskType(ctx context.Context, capComm *CapComm, task Task, taskKind taskKind) (*operation, error) {
	switch taskKind {
	case taskKindType:
		return mc.prepareTaskKindType(ctx, capComm, task)
	case taskKindTry:
		return mc.prepareSequentialTaskList(ctx, capComm, "try: "+task.Name, task.Try, task.OnFail, true, "")
	case taskKindGroup:
		return mc.prepareSequentialTaskList(ctx, capComm, "group: "+task.Name, task.Group, task.OnFail, false, "")
	case taskKindConcurrent:
		return mc.prepareConcurrentTaskList(ctx, capComm, "concurrent: "+task.Name, task.Concurrent, task.OnFail)
	}

	return nil, nil
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

	op, err := mc.switchTaskType(ctx, capComm, task, taskKind)
	if err != nil {
		return nil, err
	}

	if op != nil {
		applyTaskHandlers(capComm, task, op)
	}

	return op, nil
}

func evalPreVars(ctx context.Context, capComm *CapComm, task Task) error {
	// Evaluate the local variables
	for k, v := range task.PreVars {
		exp, err := capComm.ExpandString(ctx, k, v)
		if err != nil {
			return err
		}

		capComm.SetLocalVariable(k, exp)
	}

	return nil
}

func evalPostVars(ctx context.Context, capComm *CapComm, task Task) error {
	// Evaluate the variables post execution
	for k, v := range task.PostVars {
		exp, err := capComm.ExpandString(ctx, k, v)
		if err != nil {
			return err
		}

		capComm.ExportVariable(k, exp)
	}

	return nil
}

func applyPreVarsHandler(capComm *CapComm, task Task, op *operation) {
	op.AddHandler(func(next ExecuteFunc) ExecuteFunc {
		return func(opCtx context.Context) error {
			err := evalPreVars(opCtx, capComm, task)
			if err != nil {
				return err
			}
			// recalculate any environment variables with templates prior to running the task
			err = capComm.MergeTemplateEnvs(opCtx, task.Env)
			if err != nil {
				return err
			}

			return next(opCtx)
		}
	})
}

func applyConditionHandler(capComm *CapComm, task Task, op *operation) {
	op.AddHandler(func(next ExecuteFunc) ExecuteFunc {
		return func(execCtx context.Context) error {
			ok, err := capComm.ExpandBool(execCtx, "if", task.If)
			if err != nil {
				return err
			}

			if !ok {
				return nil
			}

			return next(execCtx)
		}
	})
}

func applyPostVarsHandler(capComm *CapComm, task Task, op *operation) {
	op.AddHandler(func(next ExecuteFunc) ExecuteFunc {
		return func(opCtx context.Context) error {
			err := next(opCtx)
			if err == nil {
				err = evalPostVars(opCtx, capComm, task)
			}
			return err
		}
	})
}

func applyExportHandler(capComm *CapComm, task Task, op *operation) {
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

func applyTaskHandlers(capComm *CapComm, task Task, op *operation) {
	// Handlers are midleware chain, handlers registered later wrap earlier ones
	// i.e. if a later handler fails it will not run the innder handlers above it in this list.

	// export variables
	if len(task.Export) > 0 {
		applyExportHandler(capComm, task, op)
	}

	// evaluate any post run variables
	if len(task.PostVars) > 0 {
		applyPostVarsHandler(capComm, task, op)
	}

	// Run the condition check, skips running handlers above if condition is false
	// pre variables are evaluated though.
	if task.If != "" {
		applyConditionHandler(capComm, task, op)
	}

	// Add handler for preVars if there are prevars or environment variables with tempates defined
	// templates can include vars
	if len(task.PreVars) > 0 || len(task.Env) > 0 {
		applyPreVarsHandler(capComm, task, op)
	}
}

func validateTaskKind(task Task) error {
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
	if len(task.Concurrent) > 0 {
		kinds++
	}

	if kinds == 0 {
		return fmt.Errorf("task %s kind not specificed needs to be a type, concurrent, group or try task", task.Name)
	}

	if kinds > 1 {
		return fmt.Errorf("task %s has multiple kinds needs to be a single type, concurrent, group or try task", task.Name)
	}

	return nil
}

func getTaskKind(task Task) (taskKind, error) {
	if err := validateTaskKind(task); err != nil {
		return taskKindNone, err
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
	if len(task.Concurrent) > 0 {
		return taskKindConcurrent, nil
	}
	panic("kind?")
}

func combineSequentialTaskListOperations(ctx context.Context, capComm *CapComm, operations operations, onFail ExecuteFunc,
	groupDesc string, tryOp bool, taskDir string) (*operation, error) {
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
		makeItSo:    impulseAhead(operations, dir, capComm.Log()),
		try:         tryOp,
		onFail:      onFail,
	}, nil
}

func combineConcurrentTaskListOperations(capComm *CapComm, operations operations, onFail ExecuteFunc,
	groupDesc string, tryOp bool) (*operation, error) {
	if len(operations) == 0 {
		return nil, nil
	}

	// handle any dir change
	return &operation{
		description: groupDesc,
		makeItSo:    engageWarpDrive(operations, capComm.Log()),
		try:         tryOp,
		onFail:      onFail,
	}, nil
}

func (mc *missionControl) prepareOperationsFromTaskList(ctx context.Context, capComm *CapComm,
	tasks Tasks, onFailTask *Task) (operations, ExecuteFunc, error) {
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
				return nil, nil, errors.Wrapf(err, "%s merge with ref %s", task.Name, task.Ref)
			}
		}

		// prepare the task
		op, err := mc.prepareTask(ctx, capComm, task)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "prepare: %s", task.Name)
		}

		if op != nil {
			operations = append(operations, op)
		}
	}

	// Is there ar failure task?
	var onFail ExecuteFunc
	if onFailTask != nil {
		onFailOp, err := mc.prepareFailTask(ctx, capComm, *onFailTask, taskMap)
		if err != nil {
			return nil, nil, err
		}

		onFail = onFailOp.makeItSo
	}

	return operations, onFail, nil
}

func (mc *missionControl) prepareConcurrentTaskList(ctx context.Context, capComm *CapComm,
	groupDesc string, tasks Tasks, onFailTask *Task) (*operation, error) {
	operations, onFail, err := mc.prepareOperationsFromTaskList(ctx, capComm, tasks, onFailTask)
	if err != nil {
		return nil, err
	}

	return combineConcurrentTaskListOperations(capComm, operations, onFail, groupDesc, false)
}

func (mc *missionControl) prepareSequentialTaskList(ctx context.Context, capComm *CapComm,
	groupDesc string, tasks Tasks, onFailTask *Task,
	tryOp bool, taskDir string) (*operation, error) {
	operations, onFail, err := mc.prepareOperationsFromTaskList(ctx, capComm, tasks, onFailTask)
	if err != nil {
		return nil, err
	}

	return combineSequentialTaskListOperations(ctx, capComm, operations, onFail, groupDesc, tryOp, taskDir)
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
		onFailOp, err := mc.prepareFailTask(ctx, capComm, *task.OnFail, make(TaskMap))
		if err != nil {
			return nil, err
		}

		onFail = onFailOp.makeItSo
	}

	return &operation{
		description: fmt.Sprintf("task: %s", task.Name),
		makeItSo:    taskFunc,
		try:         false,
		onFail:      onFail,
	}, nil
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

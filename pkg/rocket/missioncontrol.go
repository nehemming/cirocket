package rocket

import (
	"context"
	"fmt"
	"sync"

	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/pkg/errors"
)

var (
	// once gates the singleton Default mission control.
	once sync.Once

	// defaultControl is the Default singleton mission control.
	defaultControl MissionControl
)

type (
	// ExecuteFunc is the function signature of an activity that can be executed.
	ExecuteFunc = loggee.ActivityFunc

	// TaskType represents a specific task type.
	TaskType interface {

		// Type of the task
		Type() string

		// Prepare the task from the input details.
		Prepare(ctx context.Context, capComm *CapComm, task Task) (ExecuteFunc, error)
	}

	// MissionControl seeks out new civilizations in te CI space.
	MissionControl interface {
		// RegisterTaskTypes we only want the best.
		RegisterTaskTypes(types ...TaskType)

		// LaunchMission loads and executes the mission
		// flightSequences may be specified, each sequence is run in the order specified
		// the coonfig file is the source name iof the config provided
		// if its empty the current working 'dir/default' will be used.
		LaunchMission(ctx context.Context, configFile string, spaceDust map[string]interface{}, flightSequences ...string) error

		// LaunchMissionWithParams loads and executes the mission with user supplied parameters
		// flightSequences may be specified, each sequence is run in the order specified
		// the coonfig file is the source name iof the config provided
		// if its empty the current working 'dir/default' will be used.
		// The supplied params are default values and do not override values defined in the mission
		LaunchMissionWithParams(ctx context.Context, configFile string,
			spaceDust map[string]interface{},
			params []Param,
			flightSequences ...string) error
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
}

// NewMissionControl create a new mission control.
func NewMissionControl() MissionControl {
	return &missionControl{
		types: make(map[string]TaskType),
	}
}

// Default returns the default shared mission control.
func Default() MissionControl {
	once.Do(func() { defaultControl = NewMissionControl() })
	return defaultControl
}

// RegisterActorTypes actor types.
func (mc *missionControl) RegisterTaskTypes(types ...TaskType) {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	for _, t := range types {
		mc.types[t.Type()] = t
	}
}

func (mc *missionControl) LaunchMission(ctx context.Context, configFile string, spaceDust map[string]interface{}, flightSequences ...string) error {
	return mc.LaunchMissionWithParams(ctx, configFile, spaceDust, nil, flightSequences...)
}

func (mc *missionControl) LaunchMissionWithParams(ctx context.Context, configFile string,
	spaceDust map[string]interface{}, params []Param,
	flightSequences ...string) error {
	configFile, err := getConfigFileName(configFile)
	if err != nil {
		return err
	}

	// Load the mission
	mission, err := loadPreMission(ctx, spaceDust, configFile)
	if err != nil {
		return err
	}

	// Create a cap comm object from the environment
	capComm := newCapCommFromEnvironment(configFile, loggee.Default())

	// Misssion has been successfully parsed, load the global settings
	capComm, err = processGlobals(ctx, capComm, mission, params)
	if err != nil {
		return errors.Wrap(err, "global settings failure")
	}

	stagesToRun, err := getStagesTooRun(mission, flightSequences)
	if err != nil {
		return err
	}

	operations := make(operations, 0)

	// prepare stages
	for index, stage := range stagesToRun {
		if stage.Name == "" {
			stage.Name = fmt.Sprintf("stage[%d]", index)
		}

		// prepare the stage
		ops, err := mc.prepareStage(ctx, capComm, stage)
		if err != nil {
			return errors.Wrapf(err, "prepare %s", stage.Name)
		}

		if len(ops) > 0 {
			operations = append(operations, &operation{
				description: fmt.Sprintf("stage: %s", stage.Name),
				makeItSo:    engage(ctx, ops),
				try:         stage.Try,
			})
		}
	}

	return runOperations(ctx, operations)
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

func (mc *missionControl) prepareStage(ctx context.Context, missionCapComm *CapComm, stage Stage) (operations, error) {
	operations := make(operations, 0, 10)

	// Create a new CapComm for the stage
	capComm := missionCapComm.Copy(stage.NoTrust).
		MergeBasicEnvMap(stage.BasicEnv)

	if capComm.isFiltered(stage.Filter) {
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
	// Create a new CapComm for the task
	capComm := stageCapComm.Copy(task.NoTrust).
		MergeBasicEnvMap(task.BasicEnv)

	if capComm.isFiltered(task.Filter) {
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

	return &operation{
		description: fmt.Sprintf("task: %s", task.Name),
		makeItSo:    taskFunc,
		try:         task.Try,
	}, nil
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

func engage(_ context.Context, ops operations) ExecuteFunc {
	return func(ctx context.Context) error {
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

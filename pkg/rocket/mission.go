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

import "github.com/pkg/errors"

type (
	// Include contains details of an include file within the config.
	Include struct {
		Path string `mapstructure:"path"`
		URL  string `mapstructure:"url"`
	}

	// PreMission preprocesses missions before loading them.
	PreMission struct {
		Mission  map[string]interface{} `mapstructure:",remain"`
		Includes []Include              `mapstructure:"includes"`
	}

	// Mission is activity to complete.
	Mission struct {
		// Mission name, defaults to the config file name
		Name string `mapstructure:"name"`

		// Description is a free text description of the mission.
		Description string `mapstructure:"description"`

		// Additional contains any additional parameters specified in the
		// configuration.  They are included in the template data set
		// but will be overridden by any other duplicate keys
		Additional map[string]interface{} `mapstructure:",remain"`

		// BasicEnv is a map of additional environment variables
		// They are not template expanded
		BasicEnv VarMap `mapstructure:"basicEnv"`

		// Env is a map of additional environment variables
		// These are subject to template expansion after the params have been expanded
		Env VarMap `mapstructure:"env"`

		// Must is a slice of params that must be defined prior to the mission starting
		// Iif any are missing the mission will fail.
		Must MustHaveParams `mapstructure:"must"`

		// Params is a collection of parameters that can be used within
		// the child stages.  Parameters are template expanded and can use
		// Environment variables defined in Env
		Params Params `mapstructure:"params"`

		// Sequences specify a list of stages to run.
		// If no sequences are provides all stages are run in the order they are defined.
		// If sequences are included in the mission one must be specified or the mission will fail
		Sequences map[string][]string `mapstructure:"sequences"`

		// Stages represents the stages of the mission
		// If no sequences are included in the file all stages are executed in ordinal order.
		// If a sequence is included in the mission file the launch mission call must specify the sequence to use.
		Stages Stages `mapstructure:"stages"`

		// OnFail is a stage that is executed if the mission fails.
		OnFail *Stage `mapstructure:"onfail"`

		// Version of the mission definition
		Version string `mapstructure:"version"`
	}

	// MustHaveParams is a slice of param names that must be definedbefore a mission, stage or activity starts.
	// The list is checked prior too processing the activities own set of param definitions.
	MustHaveParams []string

	// Exports is a list of exported variables from on task context to another.
	Exports []string

	// Stages is a slice of stages.
	Stages []Stage

	// Tasks is a slice of tasks.
	Tasks []Task

	// Params is a slice of params.
	Params []Param

	// Param is an expandible parameter.
	Param struct {
		// Name is the name of the parameter.
		// Name is mandatory.
		Name string `mapstructure:"name"`

		// Description is a free text description of the parameter.
		Description string `mapstructure:"description"`

		// Filter is an optional filter on the param.
		// If the param criteria are not met the param value will not be set.
		Filter *Filter `mapstructure:"filter"`

		// Optional if true allows the file not to exist.
		Optional bool `mapstructure:"optional"`

		// Path is the location of a resource that can provide the parameter's value.
		// Paths may be either a local file system path or a url to file or http(s) resource.
		// If both Path and Value are supplied the resource value will be appended to the Value.
		// Template expressions can be used in the path.  The template is expanded to get the
		// location of the resource.
		//
		// If the resource is successfully located it is appended to the Value property.
		// If SkipExpand is false the combined value will be processed as a template to
		// obtain the final value.   If SkipExpand is true the combined value will be used without
		// any additional expansion.
		Path string `mapstructure:"path"`

		// Print if true will display the value of the parameter once expanded to the log.
		Print bool `mapstructure:"print"`

		// SkipExpand skip templating the param.
		SkipExpand bool `mapstructure:"skipExpand"`

		// Value is the value of the parameter.  If SkipExpand is false the value will
		// be transformed using template expansion.
		Value string `mapstructure:"value"`
	}

	// VarMap is a map of variables to their values.
	VarMap map[string]string

	// Stage is a collection of tasks that can share a common set of parameters.
	// All tasks within a stage are executed sequently.
	Stage struct {
		// Name of the stage.
		// If it is not provided it default to the ordinal ID of the stage within the mission
		Name string `mapstructure:"name"`

		// Stage is impmented by another named stage.
		Ref string `mapstructure:"ref"`

		// Description is a free text description of the stage.
		Description string `mapstructure:"description"`

		// BasicEnv is a map of additional environment variables
		// They are not template expanded
		BasicEnv VarMap `mapstructure:"basicEnv"`

		// If is evaluated prior to running a stage.  If the condition template expression evaluates to true/yes/1 the
		// stage will be run.  If the template is blank or non true value the stage will not be run and the step will be skipped.
		If string `mapstructure:"if"`

		// Dir is the directory to execute the stage in.
		Dir string `mapstructure:"dir"`

		// Env is a map of additional environment variables
		// These are subject to template expansion after the params have been expanded.
		Env VarMap `mapstructure:"env"`

		// Filter is an optional filter on the stage.
		// If the filter criteria are not met the stage will not be executed.
		Filter *Filter `mapstructure:"filter"`

		// Must is a slice of params that must be defined prior to the stage starting
		// Iif any are missing the mission will fail.
		Must MustHaveParams `mapstructure:"must"`

		// NoTrust indicates the stage should not inherit environment
		// variables or parameters from its parent.  This can be used with a run stage
		// where you do not want the process to receive API tokens etc.
		NoTrust bool `mapstructure:"noTrust"`

		// OnFail is a task that is executed if the stage fails.
		OnFail *Task `mapstructure:"onfail"`

		// Params is a collection of parameters that can be used within
		// the child stages.  Parameters are template expanded and can use
		// Environment variables defined in Env
		Params Params `mapstructure:"params"`

		// Tasks is a collection of one or more tasks to execute
		// Tasks are executed sequentially
		Tasks Tasks `mapstructure:"tasks"`
	}

	// Task is an activity that is executed.
	Task struct {
		// Name of the task.
		// If it is not provided it default to the ordinal ID of the task within the stage
		Name string `mapstructure:"name"`

		// Task is impmented by another named task in the same stage.
		Ref string `mapstructure:"ref"`

		// BasicEnv is a map of additional environment variables.
		// They are not template expanded.
		BasicEnv VarMap `mapstructure:"basicEnv"`

		// Concurrent is a list of tasks to execute concurrently.
		Concurrent Tasks `mapstructure:"concurrent"`

		// If is evaluated prior to running a task.  If the condition template expression evaluates to true/yes/1 the
		// task will be run.  If the template is blank or non true value the task will not be run and the step will be skipped.
		If string `mapstructure:"if"`

		// Description is a free text description of the task.
		Description string `mapstructure:"description"`

		// Definition contains the additional data required to process the task type
		Definition map[string]interface{} `mapstructure:",remain"`

		// Env is a map of additional environment variables.
		// These are subject to template expansion after the params have been expanded.
		Env VarMap `mapstructure:"env"`

		// Export is a list of variables to export. This list can be used by try and group task types
		// to export their variables (output from sub tasks) to their parent stage or task.
		Export Exports `mapstructure:"export"`

		// Filter is an optional filter on the task.
		// If the filter criteria are not met the task will not be executed.
		Filter *Filter `mapstructure:"filter"`

		// Try is a list of tasks to try.
		Group Tasks `mapstructure:"group"`

		// Must is a slice of params that must be defined prior to the task starting
		// Iif any are missing the mission will fail.
		Must MustHaveParams `mapstructure:"must"`

		// NoTrust indicates the task should not inherit environment
		// variables or parameters from the parent.  This can be used with a run task
		// where you do not want the process to receive API tokens etc.
		NoTrust bool `mapstructure:"noTrust"`

		// OnFail is a task that is executed if the stage fails.
		OnFail *Task `mapstructure:"onfail"`

		// Params is a collection of parameters that can be used within
		// the child stages.  Parameters are template expanded and can use
		// Environment variables defined in Env.
		// Params and Env variables are template expanded during the preparation phase of a mission.  That means their
		// values are calculated prior to any task running in any stage.
		// If values need to be calculated before or after a run use pre or post variaBLES
		Params Params `mapstructure:"params"`

		// PostVars are variable evaluated after a task has run
		// Post variable are automatically exported to the parent task/stage.
		PostVars VarMap `mapstructure:"postvars"`

		// PreVars are variables calculated immediately prior to a run.  They are are not exported to the parent context.
		// If the variable needs to be used by other tasks it should be explicitly exported (See Export above).
		PreVars VarMap `mapstructure:"prevars"`

		// Try is a list of tasks to try.
		Try Tasks `mapstructure:"try"`

		// Type is the type of the task.  The task type must have been registered
		// with the mission control.  Tasks not registered will fail the mission.
		Type string `mapstructure:"type"`
	}

	// Filter restricts running an activity
	// The filter applis to the OS and Architecture of the machine running
	// rocket.   This allows OS specific scripts to be used.
	Filter struct {
		// IncludeOS is a list of operating systems to include.
		IncludeOS []string `mapstructure:"includeOS"`

		// IncludeArch is a list of architectures to permit.
		IncludeArch []string `mapstructure:"includeArch"`

		// ExcludeOS restricts an operating system from running.
		ExcludeOS []string `mapstructure:"excludeOS"`

		// ExcludeArch restricts specific architectures from running.
		ExcludeArch []string `mapstructure:"excludeArch"`

		// Skip prevents theactivity from running if true.
		Skip bool `mapstructure:"skip"`
	}

	// OutputSpec defines the method of outputtting for a given resource.  The choice is
	// either variables or files.
	OutputSpec struct {
		// Variable is an exported variable available to later tasks in the same stage.
		Variable string `mapstructure:"variable"`

		// Output is a path to a file replacing STDOUT.
		Path string `mapstructure:"path"`

		// AppendOutput specifies if output should append.
		Append bool `mapstructure:"append"`

		// SkipExpand when true skips template expansion of the runbook.
		SkipExpand bool `mapstructure:"skipExpand"`

		// OS File permissions
		FileMode uint `mapstructure:"fileMode"`
	}

	// InputSpec is a resource input specificsation.  Input data can be provided from
	// inline valuses, exported stage variables, local files or a web url.
	InputSpec struct {
		// Variable name to import from.
		Variable string `mapstructure:"variable"`

		Inline string `mapstructure:"inline"`

		// Path provides the path to the input file.
		Path string `mapstructure:"path"`

		// URl provides a url to th input data.
		URL string `mapstructure:"url"`

		// Optional is true if resource can be missing.
		Optional bool `mapstructure:"optional"`

		// URLTimeout request timeout, default is 30 seconds.
		URLTimeout uint `mapstructure:"timeout"`

		// SkipExpand when true skips template expansion of the runbook.
		SkipExpand bool `mapstructure:"skipExpand"`
	}

	// Redirection is provided to a task to interpret
	// Redirection strings need to be expanded by the task.
	Redirection struct {
		// Input runbook
		Input *InputSpec `mapstructure:"input"`

		// Output runbook
		Output *OutputSpec `mapstructure:"output"`

		// Error runbook
		Error *OutputSpec `mapstructure:"error"`

		// MergeErrorWithOutput specifies if error output should go to outputt
		// if specified Error and AppendError are ignored
		MergeErrorWithOutput bool `mapstructure:"merge"`

		// LogOutput if true will cause output to be logged rather than going to go to std output.
		// If an output file is specified it will be used instead.
		LogOutput bool `mapstructure:"logStdOut"`

		// DirectError when true causes the commands std error output to go direct to running processes std error
		// When DirectError is false std error output is logged.
		DirectError bool `mapstructure:"directStdErr"`
	}
)

// Validate checks that include has one and only one resource identifier defined.
func (l *Include) Validate() error {
	count := 0
	if l.Path != "" {
		count++
	}
	if l.URL != "" {
		count++
	}
	if count > 1 {
		return errors.New("more than one source was specified, only one is permitted")
	}
	if count == 0 {
		return errors.New("no source was specified")
	}

	return nil
}

// Copy copies an environment map.
func (em VarMap) Copy() VarMap {
	c := make(VarMap)
	for k, v := range em {
		c[k] = v
	}
	return c
}

// Copy copies the must have params.
func (params MustHaveParams) Copy() MustHaveParams {
	c := make(MustHaveParams, len(params))
	copy(c, params)
	return c
}

// Copy copies the must have params.
func (exports Exports) Copy() Exports {
	c := make(Exports, len(exports))
	copy(c, exports)
	return c
}

// Copy the params.
func (params Params) Copy() Params {
	c := make(Params, len(params))
	copy(c, params)
	return c
}

// Copy the tasks.
func (tasks Tasks) Copy() Tasks {
	c := make(Tasks, len(tasks))
	copy(c, tasks)
	return c
}

// ToMap converts the task slice to a map using the task name as the key.
func (tasks Tasks) ToMap() TaskMap {
	c := make(TaskMap)

	for _, v := range tasks {
		if v.Name != "" {
			c[v.Name] = v
		}
	}

	return c
}

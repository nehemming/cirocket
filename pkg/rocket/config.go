package rocket

type (
	Include struct {
		Path string `mapstructure:"path"`
		Url  string `mapstructure:"url"`
	}

	// PreMission preprocesses missions before loading them
	PreMission struct {
		Mission  map[string]interface{} `mapstructure:",remain"`
		Includes []Include              `mapstructure:"includes"`
	}

	// Mission is activity to complete
	Mission struct {
		// Mission name, defsaults to the config file name
		Name string `mapstructure:"name"`

		// Additional contains any additional paramters specified in the
		// configuration.  They are included in the template data set
		// but will be overridden by any other duplicate keys
		Additional map[string]interface{} `mapstructure:",remain"`

		// BasicEnv is a map of additional environment variables
		// They are not template expanded
		BasicEnv EnvMap `mapstructure:"basicEnv"`

		// Env is a map of additional environment variables
		// These are subject to template expansion after the params have been expanded
		Env EnvMap `mapstructure:"env"`

		// Params is a collection of parameters that can be used within
		// the child stages.  Paramters are template expanded and can use
		// Environment variables defined in Env
		Params []Param `mapstructure:"params"`

		// Sequences specify a list of stages to run.
		// If no sequences are provides all stages are run in the order they are defined.
		// If sequences are included in the mission one must be specified or the mission will fail
		Sequences map[string][]string `mapstructure:"sequences"`

		// Stages represents the stages of the mission
		// If no sequences are included in the file all stages are executed in ordinal order.
		// If a sequence is included in the mission file the fly mission call must specify the sequence to use.
		Stages []Stage `mapstructure:"stages"`

		// Version of the mission definition
		Version string `mapstructure:"version"`
	}

	// Param is an expandible parameter
	Param struct {
		// Name is the name of the parameter
		// Name is mandatory
		Name string `mapstructure:"name"`

		// Value is the value of the parameter and is subject to expandion
		Value string `mapstructure:"value"`

		// File is a path to file containing the value.
		// If both File and Value are supplied the file will bee appended to the Valued
		// The combined value will undergo template expansion
		File string `mapstructure:"file"`

		// SkipTemplate skip templating the param
		SkipTemplate bool `mapstructure:"skipTemplate"`

		// FileOptional if truew allows the file not to exist
		FileOptional bool `mapstructure:"optional"`
	}

	// EnvMap is a map of environment variables to their values
	EnvMap map[string]string

	// State iis a collection of tasks that can share a common set of parameters
	// All tasks within a stage are executed sequently
	Stage struct {
		// Name of th stage.
		// If it is not provided it default to the ordinal ID of the stage within the mission
		Name string `mapstructure:"name"`

		// BasicEnv is a map of additional environment variables
		// They are not template expanded
		BasicEnv EnvMap `mapstructure:"basicEnv"`

		// Env is a map of additional environment variables
		// These are subject to template expansion after the params have been expanded
		Env EnvMap `mapstructure:"env"`

		// Filter is an optional filter on the stage
		// If the filter criteria are not met the stage will not be executed
		Filter *Filter `mapstructure:"filter"`

		// NoTrust indicates the stage should not inherit environment
		// variables or paramters from its parent.  This can be used with a run stage
		// where you do not want the process to receive API tokens etc
		NoTrust bool `mapstructure:"noTrust"`

		// Params is a collection of parameters that can be used within
		// the child stages.  Paramters are template expanded and can use
		// Environment variables defined in Env
		Params []Param `mapstructure:"params"`

		// Tasks is a collection of one or more tasks to complete
		// Tasks are executed sequentally
		Tasks []Task `mapstructure:"tasks"`

		// Try to run the stage but if it fails do no abort the whole run
		Try bool `mapstructure:"try"`
	}

	// Task is an activity that is executed
	Task struct {
		// Type is the type of the task.  The task type must have been rejiggered
		// with the mission control.  Tasks not registered will fail the mission.
		Type string `mapstructure:"type"`

		// Name of the task.
		// If it is not provided it default to the ordinal ID of the task within the stage
		Name string `mapstructure:"name"`

		// Definition contains the additional data required to process the task type
		Definition map[string]interface{} `mapstructure:",remain"`

		// BasicEnv is a map of additional environment variables
		// They are not template expanded
		BasicEnv EnvMap `mapstructure:"basicEnv"`

		// Env is a map of additional environment variables
		// These are subject to template expansion after the params have been expanded
		Env EnvMap `mapstructure:"env"`

		// Filter is an optional filter on the task
		// If the filter criteria are not met the task will not be executed
		Filter *Filter `mapstructure:"filter"`

		// NoTrust indicates the task should not inherit environment
		// variables or paramters from the parent.  This can be used with a run task
		// where you do not want the process to receive API tokens etc
		NoTrust bool `mapstructure:"noTrust"`

		// Params is a collection of parameters that can be used within
		// the child stages.  Paramters are template expanded and can use
		// Environment variables defined in Env
		Params []Param `mapstructure:"params"`

		// Try to run the task but if it fails do no abort the whole run
		Try bool `mapstructure:"try"`
	}

	// Filter restricts running an activity
	// The filter applis to the OS and Architecture of the machine running
	// rocket.   This allows OS specific scripts to be used.
	Filter struct {
		// IncludeOS is a list of operating systems to include
		IncludeOS []string `mapstructure:"includeOS"`
		// IncludeArch is a list of architectures to permit
		IncludeArch []string `mapstructure:"includeArch"`

		// ExcludeOS restricts an operating system from running
		ExcludeOS []string `mapstructure:"excludeOS"`

		// ExcludeArch restricts specific architectures from running
		ExcludeArch []string `mapstructure:"excludeArch"`

		// Skip prevents theactivity from running if true.
		Skip bool `mapstructure:"skip"`
	}

	OutputSpec struct {
		// Output is a path to a file replacing STDOUT
		Output string `mapstructure:"output"`

		// AppendOutput specifices if output should append
		AppendOutput bool `mapstructure:"appendOutput"`
	}

	// Redirection is provided to a task to interpret
	// Redirection strings need to be expanded by the task
	Redirection struct {
		OutputSpec `mapstructure:",squash"`

		// Input is a file path to an existing input file replacing STDIN
		Input string `mapstructure:"input"`

		// Error is a path to a file replacing STDERR
		Error string `mapstructure:"error"`

		// AppendError specifies if error output should append
		AppendError bool `mapstructure:"appendError"`

		// MergeErrorWithOutput specifies if error output should go to outputt
		// if specified Error and AppendError are ignored
		MergeErrorWithOutput bool `mapstructure:"merge"`
	}

	// Delims are the delimiters to use to escape template functions
	Delims struct {
		// Left is the opening delimiter
		Left string `mapstructure:"left"`

		// Right is the closing delimiter
		Right string `mapstructure:"right"`
	}
)

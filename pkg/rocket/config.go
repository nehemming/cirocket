package rocket

type (
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
		// Mission name, defsaults to the config file name
		Name string `mapstructure:"name"`

		// Additional contains any additional parameters specified in the
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
		// the child stages.  Parameters are template expanded and can use
		// Environment variables defined in Env
		Params []Param `mapstructure:"params"`

		// Sequences specify a list of stages to run.
		// If no sequences are provides all stages are run in the order they are defined.
		// If sequences are included in the mission one must be specified or the mission will fail
		Sequences map[string][]string `mapstructure:"sequences"`

		// Stages represents the stages of the mission
		// If no sequences are included in the file all stages are executed in ordinal order.
		// If a sequence is included in the mission file the launch mission call must specify the sequence to use.
		Stages []Stage `mapstructure:"stages"`

		// Version of the mission definition
		Version string `mapstructure:"version"`
	}

	// Param is an expandible parameter.
	Param struct {
		// Name is the name of the parameter
		// Name is mandatory
		Name string `mapstructure:"name"`

		// Value is the value of the parameter and is subject to expandion
		Value string `mapstructure:"value"`

		// Path is a path to file containing the value.
		// If both Path and Value are supplied the file will bee appended to the Valued
		// The combined value will undergo template expansion if SkipTemplate is false.
		Path string `mapstructure:"path"`

		// URL specifies the data should come from the response body or a web request.
		// The url body will be concatenated with the value and file values respectively.
		// The combined value will undergo template expansion if SkipTemplate is false.
		URL string `mapstructure:"url"`

		// SkipExpand skip templating the param
		SkipExpand bool `mapstructure:"skipExpand"`

		// Optional if true allows the file not to exist
		Optional bool `mapstructure:"optional"`
	}

	// EnvMap is a map of environment variables to their values.
	EnvMap map[string]string

	// Stage is a collection of tasks that can share a common set of parameters.
	// All tasks within a stage are executed sequently.
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
		// variables or parameters from its parent.  This can be used with a run stage
		// where you do not want the process to receive API tokens etc
		NoTrust bool `mapstructure:"noTrust"`

		// Params is a collection of parameters that can be used within
		// the child stages.  Parameters are template expanded and can use
		// Environment variables defined in Env
		Params []Param `mapstructure:"params"`

		// Tasks is a collection of one or more tasks to complete
		// Tasks are executed sequentally
		Tasks []Task `mapstructure:"tasks"`

		// Try to run the stage but if it fails do no abort the whole run
		Try bool `mapstructure:"try"`
	}

	// Task is an activity that is executed.
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
		// variables or parameters from the parent.  This can be used with a run task
		// where you do not want the process to receive API tokens etc
		NoTrust bool `mapstructure:"noTrust"`

		// Params is a collection of parameters that can be used within
		// the child stages.  Parameters are template expanded and can use
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
		// Variable is an exported variable available to later tasks in the same stage
		Variable string `mapstructure:"variable"`

		// Output is a path to a file replacing STDOUT
		Path string `mapstructure:"path"`

		// AppendOutput specifies if output should append
		Append bool `mapstructure:"append"`

		// SkipExpand when true skips template expansion of the spec.
		SkipExpand bool `mapstructure:"skipExpand"`

		// OS File permissions
		FileMode uint `mapstructure:"fileMode"`
	}

	InputSpec struct {
		// Variable name to import from
		Variable string `mapstructure:"variable"`

		Inline string `mapstructure:"inline"`

		// Path provides the path data.
		Path string `mapstructure:"path"`

		// URl provides the data.
		URL string `mapstructure:"url"`

		// Optional is true if resource can be missing.
		Optional bool `mapstructure:"optional"`

		// URLTimeout request timeout, default is 30 seconds.
		URLTimeout uint `mapstructure:"timeout"`

		// SkipExpand when true skips template expansion of the spec.
		SkipExpand bool `mapstructure:"skipExpand"`
	}

	// Redirection is provided to a task to interpret
	// Redirection strings need to be expanded by the task.
	Redirection struct {
		// Input specification
		Input *InputSpec `mapstructure:"input"`

		// Output specification
		Output *OutputSpec `mapstructure:"output"`

		// Error specification
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

	// Delims are the delimiters to use to escape template functions.
	Delims struct {
		// Left is the opening delimiter
		Left string `mapstructure:"left"`

		// Right is the closing delimiter
		Right string `mapstructure:"right"`
	}
)

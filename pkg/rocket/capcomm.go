package rocket

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/nehemming/cirocket/pkg/buildinfo"
	"github.com/pkg/errors"
)

const (
	// EnvTag is the root tag used in template data for environment variables
	// eg PATH is .Env.PATH.
	EnvTag = "Env"

	// ParentEnvTag is the root tag for environment variables when the task is not trusted.
	ParentEnvTag = "ParentEnv"

	// RuntimeTag is the remplate data key to the run time data.
	RuntimeTag = "Runtime"

	// BuildTag is the data template key for build information about the hosting application.
	BuildTag = "Build"

	// AdditionalMissionTag is the data template key to additional mission information.
	AdditionalMissionTag = "Additional"

	// ConfigFile is the config file.
	ConfigFile = "configFile"

	// ConfigBaseName is the base name of the config file, not path or extensions.
	ConfigBaseName = "configBaseName"

	// ConfigDir is the config dir.
	ConfigDir = "configDir"

	// ConfigFileFullPath is the full path to the config.
	ConfigFileFullPath = "configFullPath"
)

type (
	// TemplateData represents the data passed t a template.
	TemplateData map[string]interface{}

	// Runtime holds runtime data accessible to templates.
	Runtime struct {
		// GOOS is the operating system the host application is running against
		GOOS string
		// GOARCH is the architecture the host application is running against
		GOARCH string
	}

	// CapComm handles all communication between mission control and the mission stages and tasks.
	CapComm struct {
		data                  TemplateData
		env                   Getter
		entrustedParentEnv    Getter
		funcMap               template.FuncMap
		additionalMissionData TemplateData
		params                Getter
		runtime               Runtime
		sealed                bool
		mission               *Mission
		ioSettings            *ioSettings
	}
)

func setParamsFromConfigFile(kv map[string]string, configFile string) {
	dir, file := filepath.Split(configFile)
	kv[ConfigFileFullPath], _ = filepath.Abs(configFile)
	kv[ConfigFile] = file
	kv[ConfigBaseName] = strings.TrimSuffix(file, filepath.Ext(file))
	kv[ConfigDir] = strings.TrimSuffix(dir, string(filepath.Separator))
}

func initFuncMap() template.FuncMap {
	fm := make(template.FuncMap)

	fm["Indent"] = func(indent int, text string) string {
		lines := strings.Split(text, "\n")
		sb := strings.Builder{}
		spaces := strings.Repeat(" ", indent)
		for i, line := range lines {
			if i > 0 {
				sb.WriteString(spaces)
			}
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		return sb.String()
	}

	return fm
}

// newCapCommFromEnvironment creates a new capCom from the environment.
func newCapCommFromEnvironment(configFile string) *CapComm {
	paramKvg := NewKeyValueGetter(nil)
	setParamsFromConfigFile(paramKvg.kv, configFile)

	cc := &CapComm{
		sealed:                true,
		env:                   &osEnvGetter{},
		additionalMissionData: make(TemplateData),
		params:                paramKvg,
		funcMap:               initFuncMap(),
		runtime: Runtime{
			GOOS:   runtime.GOOS,
			GOARCH: runtime.GOARCH,
		},
		ioSettings: newIOSettings(),
	}
	return cc
}

// Copy creates a new unsealed CapComm instance and deep copies the current CapComm's values into it.
func (capComm *CapComm) Copy(noTrust bool) *CapComm {
	// Create the new capComm from the source
	// Do not copy data and d not seal
	newCapComm := &CapComm{
		sealed:                false,
		mission:               capComm.mission,
		additionalMissionData: capComm.additionalMissionData, // no copy, but safe as set once
		funcMap:               make(template.FuncMap),
		params:                NewKeyValueGetter(capComm.params),
		runtime: Runtime{
			GOOS:   runtime.GOOS,
			GOARCH: runtime.GOARCH,
		},
		ioSettings: capComm.ioSettings.newCopy(),
	}

	// Non trusted CapComm copies do not receive environment variables from their parent
	// instead they are copies into a separate parent env collection.
	// This ensures that if a task "execs" the child task will not receive the host applications environment variables
	// This may be useful if for example we want to protect a API_TOKEN for being passed to a child process
	// (Alternatively, this can be done by setting the sub task env variable to an empty string).
	if noTrust {
		newCapComm.env = NewKeyValueGetter(nil)
		newCapComm.entrustedParentEnv = NewKeyValueGetter(capComm.env)
	} else {
		newCapComm.env = NewKeyValueGetter(capComm.env)
	}

	// Copy the func mapping
	for k, v := range capComm.funcMap {
		newCapComm.funcMap[k] = v
	}

	return newCapComm
}

// Seal prevents further editing of the capComm
// any attempt to edit will cause a panic (as this is a development bug).
func (capComm *CapComm) Seal() *CapComm {
	capComm.sealed = true
	return capComm
}

// WithMission attaches the mission to the CapComm.
func (capComm *CapComm) WithMission(mission *Mission) *CapComm {
	capComm.mustNotBeSealed()

	capComm.mission = mission

	return capComm
}

// MergeBasicEnvMap adds environment variables into an unsealed CapComm.
func (capComm *CapComm) MergeBasicEnvMap(env EnvMap) *CapComm {
	capComm.mustNotBeSealed()

	if env == nil {
		return capComm
	}

	kvg := capComm.env.(*KeyValueGetter)

	for k, v := range env {
		kvg.kv[k] = v
	}

	return capComm
}

// AddAdditionalMissionData adds the additional mission data.
func (capComm *CapComm) AddAdditionalMissionData(missionData TemplateData) *CapComm {
	capComm.mustNotBeSealed()

	if len(capComm.additionalMissionData) > 0 {
		//	Data only allowed to be set once
		panic("additional mission data change")
	}

	if missionData != nil {
		capComm.additionalMissionData = missionData
	}

	return capComm
}

// AddFile adds a key named file specificsation into the CapComm.
// The filePath follows standard template and environment variable expansion.  The mode controls how the file will be used.
// Files can be added to sealed CapComm's.
func (capComm *CapComm) AddFile(ctx context.Context, name NamedIO, filePath string, mode IOMode) error {
	v, err := capComm.ExpandString(ctx, string(name), filePath)
	if err != nil {
		return errors.Wrapf(err, "expand %s file path", name)
	}
	capComm.ioSettings.addFilePath(name, v, mode)
	return nil
}

// GetFileDetails returns the file details of the named file.  If the file key does not exist nil is returned.
func (capComm *CapComm) GetFileDetails(name NamedIO) *FileDetail {
	return capComm.ioSettings.getFileDetails(name)
}

// AttachOutput attaches an output specification to a capComm.
func (capComm *CapComm) AttachOutput(ctx context.Context, redirect OutputSpec) error {
	// Handle output

	if redirect.Output == "" {
		return nil
	}

	mode := IOModeOutput
	if redirect.AppendOutput {
		mode |= IOModeAppend
	} else {
		mode |= IOModeTruncate
	}

	return capComm.AddFile(ctx, OutputIO, redirect.Output, mode)
}

// AttachRedirect attaches a redirection specification to the capComm
// Redirection covers in, out and error streams.
func (capComm *CapComm) AttachRedirect(ctx context.Context, redirect Redirection) error {
	// Handle Input
	if redirect.Input != "" {
		if err := capComm.AddFile(ctx, InputIO, redirect.Input, IOModeInput); err != nil {
			return err
		}
	}

	// Handle output
	if err := capComm.attachRedirectOutput(ctx, redirect); err != nil {
		return err
	}

	return capComm.attachRedirectError(ctx, redirect)
}

// MergeParams adds params into an unsealed CapComm instance.
func (capComm *CapComm) MergeParams(ctx context.Context, params []Param) error {
	capComm.mustNotBeSealed()

	// Params need to be expanded prior to merging
	expanded := make(map[string]string)

	for index, p := range params {
		if p.Name == "" {
			return fmt.Errorf("parameter %d has no name", index)
		}

		v, err := capComm.expandParam(ctx, p)
		if err != nil {
			return errors.Wrapf(err, "parameter %s", p.Name)
		}
		expanded[p.Name] = v
	}

	// safe type conversion as KeyValueGetter used all cases except root thats sealed sp not possible to be here
	kvg := capComm.params.(*KeyValueGetter)
	for k, v := range expanded {
		kvg.kv[k] = v
	}

	return nil
}

// MergeTemplateEnvs adds params into an unsealed CapComm instance.
func (capComm *CapComm) MergeTemplateEnvs(ctx context.Context, env EnvMap) error {
	capComm.mustNotBeSealed()

	// Params need to be expanded prior to merging
	expanded := make(map[string]string)

	for k, p := range env {
		v, err := capComm.ExpandString(ctx, k, p)
		if err != nil {
			return errors.Wrapf(err, "env %s", k)
		}

		expanded[k] = v
	}

	kvg := capComm.env.(*KeyValueGetter)
	for k, v := range expanded {
		kvg.kv[k] = v
	}

	return nil
}

// ExpandString expands a templated string using the capComm's template data.
func (capComm *CapComm) ExpandString(ctx context.Context, name, value string) (string, error) {
	// Is this a param template?
	if !strings.Contains(value, "{{") {
		return capComm.expandShellEnv(value), nil
	}

	// Create the template
	template, err := template.New(name).
		Funcs(capComm.funcMap).
		Parse(value)
	if err != nil {
		return "", errors.Wrap(err, "parsing template")
	}

	// Get template data and execute the template
	buf := bytes.NewBufferString("")
	err = template.Execute(buf, capComm.GetTemplateData(ctx))
	if err != nil {
		return "", errors.Wrap(err, "executing template")
	}

	// Finally expand any environment variables in the $VAR format
	return capComm.expandShellEnv(buf.String()), nil
}

// FuncMap returns the function mapping used by CapComm.
func (capComm *CapComm) FuncMap() template.FuncMap {
	m := make(template.FuncMap)

	// Make a copy
	for k, v := range capComm.funcMap {
		m[k] = v
	}

	return m
}

// GetExecEnv converts the environment variables map into the string format needed for a exec Cmd.
func (capComm *CapComm) GetExecEnv() []string {
	all := capComm.env.All()

	env := make([]string, len(all))
	i := 0
	for k, v := range all {
		env[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	return env
}

// GetTemplateData gts the data collection supplied to a template.
func (capComm *CapComm) GetTemplateData(ctx context.Context) TemplateData {
	if capComm.data != nil {
		return capComm.data
	}

	data := make(TemplateData)

	// Add mission data
	if capComm.additionalMissionData != nil {
		data[AdditionalMissionTag] = capComm.additionalMissionData
	}

	// Build data
	// Params
	data.addMaps(capComm.params.All())

	// Env
	data[EnvTag] = capComm.env.All()

	if capComm.entrustedParentEnv != nil {
		data[ParentEnvTag] = capComm.entrustedParentEnv.All()
	}

	// Runtime
	data[RuntimeTag] = capComm.runtime

	// Build
	data[BuildTag] = buildinfo.GetBuildInfo(ctx)

	// Save cached values
	if capComm.sealed {
		capComm.data = data
	}

	return data
}

// expandShellEnv expand variables in the format $VAR and ${VAR}.
func (capComm *CapComm) expandShellEnv(value string) string {
	return os.Expand(value, func(v string) string {
		return capComm.env.Get(v)
	})
}

// expandParam carries out template expansion of a parameter.
func (capComm *CapComm) expandParam(ctx context.Context, param Param) (string, error) {
	// Read param
	value := param.Value

	// If param has a file name, open it
	if param.File != "" {
		fileName, err := capComm.ExpandString(ctx, param.Name, param.File)
		if err != nil {
			return "", errors.Wrap(err, "rexpanding file name")
		}

		b, err := os.ReadFile(fileName)
		if err != nil {
			if !os.IsNotExist(err) || !param.FileOptional {
				return "", errors.Wrap(err, "reading value from file")
			}
		} else {
			value += string(b)
		}
	}

	// Skip expanding a param
	if param.SkipTemplate {
		return value, nil
	}

	return capComm.ExpandString(ctx, param.Name, value)
}

// isFiltered returns true if the filter restricts
// the item from being included.
// False means do not process any further.
func (capComm *CapComm) isFiltered(filter *Filter) bool { //nolint remain as iis for now
	if filter == nil {
		return false
	}

	if filter.Skip {
		return true
	}

	if len(filter.ExcludeArch) > 0 {
		for _, a := range filter.ExcludeArch {
			if a == capComm.runtime.GOARCH {
				return true
			}
		}
	}

	if len(filter.ExcludeOS) > 0 {
		for _, o := range filter.ExcludeOS {
			if o == capComm.runtime.GOOS {
				return true
			}
		}
	}

	if len(filter.IncludeArch) > 0 {
		included := false
		for _, a := range filter.IncludeArch {
			if a == capComm.runtime.GOARCH {
				included = true
				break
			}
		}

		if !included {
			return true
		}
	}

	if len(filter.IncludeOS) > 0 {
		included := false
		for _, o := range filter.IncludeOS {
			if o == capComm.runtime.GOOS {
				included = true
				break
			}
		}

		if !included {
			return true
		}
	}

	return false
}

// mustNotBeSealed asserts that the seal is not in place.
func (capComm *CapComm) mustNotBeSealed() {
	if capComm.sealed {
		panic("CapComm is sealed and cannot be editied")
	}
}

func (capComm *CapComm) attachRedirectError(ctx context.Context, redirect Redirection) error {
	// Handle error files
	if !redirect.MergeErrorWithOutput && redirect.Error != "" {
		mode := IOModeError
		if redirect.AppendError {
			mode |= IOModeAppend
		} else {
			mode |= IOModeTruncate
		}

		if err := capComm.AddFile(ctx, ErrorIO, redirect.Error, mode); err != nil {
			return err
		}
	}

	return nil
}

func (capComm *CapComm) attachRedirectOutput(ctx context.Context, redirect Redirection) error {
	if redirect.Output != "" {
		mode := IOModeOutput
		if redirect.AppendOutput {
			mode |= IOModeAppend
		} else {
			mode |= IOModeTruncate
		}

		// MergeErrorWithOutput uses the output spec for both out and error files.
		if redirect.MergeErrorWithOutput {
			mode |= IOModeError

			if err := capComm.AddFile(ctx, OutputIO, redirect.Output, mode); err != nil {
				return err
			}

			if err := capComm.ioSettings.duplicate(OutputIO, ErrorIO); err != nil {
				return err
			}
		} else if err := capComm.AddFile(ctx, OutputIO, redirect.Output, mode); err != nil {
			return err
		}
	}
	return nil
}

// addMaps appends map data left to right to the template data receiver.
func (td TemplateData) addMaps(maps ...map[string]string) TemplateData {
	for _, m := range maps {
		for k, v := range m {
			td[k] = v
		}
	}

	return td
}

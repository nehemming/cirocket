package rocket

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/nehemming/cirocket/pkg/buildinfo"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/providers"
	"github.com/pkg/errors"
	"golang.org/x/net/context/ctxhttp"
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

	// VariableTag is the top key for variables in the template data.
	VariableTag = "Var"

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
		resources             providers.ResourceProviderMap
		variables             exportMap
		exportTo              exportMap
		log                   loggee.Logger
	}

	exportMap map[string]string
)

// setParamsFromConfigFile adds config file entries to the environment map.
func setParamsFromConfigFile(env map[string]string, configFile string) {
	dir, file := filepath.Split(configFile)
	env[ConfigFileFullPath], _ = filepath.Abs(configFile)
	env[ConfigFile] = file
	env[ConfigBaseName] = strings.TrimSuffix(file, filepath.Ext(file))
	env[ConfigDir] = strings.TrimSuffix(dir, string(filepath.Separator))
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
func newCapCommFromEnvironment(configFile string, log loggee.Logger) *CapComm {
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
		resources: make(providers.ResourceProviderMap),
		variables: make(exportMap),
		log:       log,
	}

	// cc.resources[InputIO] = providers.NewNonClosingReaderProvider(os.Stdin)
	cc.resources[Stdin] = providers.NewNonClosingReaderProvider(os.Stdin)

	cc.resources[OutputIO] = providers.NewNonClosingWriterProvider(os.Stdout)
	cc.resources[Stdout] = providers.NewNonClosingWriterProvider(os.Stdout)
	cc.resources[Stderr] = providers.NewNonClosingWriterProvider(os.Stderr)

	cc.resources[ErrorIO] = providers.NewLogProvider(log, providers.LogWarn)

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
		resources: capComm.resources.Copy(),
		variables: make(exportMap),
		exportTo:  capComm.variables,
		log:       capComm.log,
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

// Log returns the cap com logger.
func (capComm *CapComm) Log() loggee.Logger {
	return capComm.log
}

// ExportVariable exports the passed variable to all capComm's sharing the same parent as the receiver.
func (capComm *CapComm) ExportVariable(key, value string) *CapComm {
	capComm.exportTo[key] = value
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

// AddResource adds a resource to the capComm object.
func (capComm *CapComm) AddResource(name providers.ResourceID, provider providers.ResourceProvider) {
	capComm.resources[name] = provider
}

// GetResource a resource.
func (capComm *CapComm) GetResource(name providers.ResourceID) providers.ResourceProvider {
	return capComm.resources[name]
}

// AddFileResource adds a key named file specificsation into the CapComm.
// The filePath follows standard template and environment variable expansion.  The mode controls how the file will be used.
// Files can be added to sealed CapComm's.
func (capComm *CapComm) AddFileResource(ctx context.Context, name providers.ResourceID, filePath string, mode providers.IOMode) error {
	path, err := capComm.ExpandString(ctx, string(name), filePath)
	if err != nil {
		return errors.Wrapf(err, "expand %s file path", name)
	}

	provider, err := providers.NewFileProvider(path, mode, 0666, false)
	if err != nil {
		return err
	}
	capComm.resources[name] = provider

	return nil
}

func validateInputSpec(inputSpec *InputSpec) error {
	count := 0
	if inputSpec.Variable != "" {
		count++
	}
	if inputSpec.Inline != "" {
		count++
	}
	if inputSpec.Path != "" {
		count++
	}
	if inputSpec.URL != "" {
		count++
	}

	if count > 1 {
		return errors.New("more than one input source was specified, only one is permitted")
	}
	if count == 0 {
		return errors.New("no input source was specified")
	}
	return nil
}

func (capComm *CapComm) createProviderFromInputSpec(ctx context.Context, inputSpec InputSpec) (providers.ResourceProvider, error) { //nolint:cyclop
	var rp providers.ResourceProvider
	var err error
	var v string
	if inputSpec.Variable != "" {
		v, ok := capComm.exportTo[inputSpec.Variable]
		if !ok && !inputSpec.Optional {
			return nil, fmt.Errorf("variable %s not found", inputSpec.Variable)
		}
		rp = providers.NewNonClosingReaderProvider(bytes.NewBufferString(v))
	} else if inputSpec.Inline != "" {
		if !inputSpec.SkipExpand {
			v, err = capComm.ExpandString(ctx, "inline", inputSpec.Inline)
		} else {
			v = inputSpec.Inline
		}
		rp = providers.NewNonClosingReaderProvider(bytes.NewBufferString(v))
	} else if inputSpec.Path != "" {
		if !inputSpec.SkipExpand {
			v, err = capComm.ExpandString(ctx, "path", inputSpec.Path)
			if err != nil {
				return nil, err
			}
		} else {
			v = inputSpec.Path
		}
		rp, err = providers.NewFileProvider(v, providers.IOModeInput, 0, inputSpec.Optional)
	} else if inputSpec.URL != "" {
		if !inputSpec.SkipExpand {
			v, err = capComm.ExpandString(ctx, "url", inputSpec.URL)
			if err != nil {
				return nil, err
			}
		} else {
			v = inputSpec.URL
		}
		rp, err = providers.NewURLProvider(v, time.Second*time.Duration(inputSpec.URLTimeout), inputSpec.Optional)
	} else {
		panic("validation bad input spec")
	}

	return rp, err
}

// InputSpecToResourceProvider creates a resource provider from the supplied input spec.
func (capComm *CapComm) InputSpecToResourceProvider(ctx context.Context, inputSpec InputSpec) (providers.ResourceProvider, error) {
	if err := validateInputSpec(&inputSpec); err != nil {
		return nil, err
	}

	return capComm.createProviderFromInputSpec(ctx, inputSpec)
}

// AttachInputSpec adds a named input spec to the capCom resources.
func (capComm *CapComm) AttachInputSpec(ctx context.Context, name providers.ResourceID, inputSpec InputSpec) error {
	if name == "" {
		return errors.New("name cannot be blank")
	}

	rp, err := capComm.InputSpecToResourceProvider(ctx, inputSpec)

	if err == nil {
		capComm.AddResource(name, rp)
	}

	return err
}

func validateOutputSpec(outputSpec *OutputSpec) error {
	count := 0
	if outputSpec.Variable != "" {
		count++
	}
	if outputSpec.Path != "" {
		count++
	}

	if count > 1 {
		return errors.New("more than one output source was specified, only one is permitted")
	}
	if count == 0 {
		return errors.New("no output source was specified")
	}

	return nil
}

func (capComm *CapComm) createProviderFromOutputSpec(ctx context.Context, outputSpec OutputSpec, mode providers.IOMode) (providers.ResourceProvider, error) {
	if outputSpec.Variable != "" {
		// Write to a variable
		return newVariableWriter(capComm, outputSpec.Variable), nil
	}

	// Are we appending?
	if outputSpec.Append {
		mode |= providers.IOModeAppend
	} else {
		mode |= providers.IOModeTruncate
	}

	// Get the file mode
	var fileMode os.FileMode
	if outputSpec.FileMode == 0 {
		fileMode = 0666
	} else {
		fileMode = os.FileMode(outputSpec.FileMode)
	}

	// Expand the value
	var v string
	var err error
	if !outputSpec.SkipExpand {
		v, err = capComm.ExpandString(ctx, "path", outputSpec.Path)
		if err != nil {
			return nil, err
		}
	} else {
		v = outputSpec.Path
	}

	return providers.NewFileProvider(v, mode, fileMode, false)
}

// OutputSpecToResourceProvider creates a resource provider from a output specificsation.
func (capComm *CapComm) OutputSpecToResourceProvider(ctx context.Context, outputSpec OutputSpec) (providers.ResourceProvider, error) {
	err := validateOutputSpec(&outputSpec)
	if err != nil {
		return nil, err
	}

	return capComm.createProviderFromOutputSpec(ctx, outputSpec, providers.IOModeOutput)
}

// AttachOutputSpec attaches an output specification to the capComm.
func (capComm *CapComm) AttachOutputSpec(ctx context.Context, name providers.ResourceID, outputSpec OutputSpec) error {
	if name == "" {
		return errors.New("name cannot be blank")
	}

	rp, err := capComm.OutputSpecToResourceProvider(ctx, outputSpec)

	if err == nil {
		capComm.AddResource(name, rp)
	}

	return err
}

func validateRedirection(redirect *Redirection) error { //nolint
	if redirect.LogOutput && redirect.Output != nil {
		return errors.New("cannot both redirect to the log and also provide an output specification")
	}
	if redirect.DirectError && redirect.Error != nil {
		return errors.New("cannot both redirect to stderr and also provide an error specification")
	}
	if redirect.MergeErrorWithOutput && redirect.Error != nil {
		return errors.New("cannot merge errors with output and specify an error specification")
	}
	if redirect.Input != nil {
		if err := validateInputSpec(redirect.Input); err != nil {
			return errors.Wrap(err, string(InputIO))
		}
	}
	if redirect.Output != nil {
		err := validateOutputSpec(redirect.Output)
		if err != nil {
			return errors.Wrap(err, string(OutputIO))
		}
	}
	if redirect.Error != nil {
		err := validateOutputSpec(redirect.Error)
		if err != nil {
			return errors.Wrap(err, string(ErrorIO))
		}
	}
	return nil
}

// AttachRedirect attaches a redirection specification to the capComm
// Redirection covers in, out and error streams.
func (capComm *CapComm) AttachRedirect(ctx context.Context, redirect Redirection) error { //nolint
	// Pre validate
	if err := validateRedirection(&redirect); err != nil {
		return err
	}

	if redirect.Input != nil {
		rp, err := capComm.createProviderFromInputSpec(ctx, *redirect.Input)
		if err != nil {
			return err
		}

		capComm.AddResource(InputIO, rp)
	}

	if redirect.LogOutput {
		rp := providers.NewLogProvider(capComm.log, providers.LogInfo)
		capComm.AddResource(OutputIO, rp)

		if redirect.MergeErrorWithOutput {
			capComm.AddResource(ErrorIO, rp)
		}
	}

	if redirect.DirectError {
		capComm.AddResource(ErrorIO, capComm.GetResource(Stderr))
	}

	if redirect.Output != nil {
		mode := providers.IOModeOutput
		if redirect.MergeErrorWithOutput {
			mode |= providers.IOModeError
		}

		rp, err := capComm.createProviderFromOutputSpec(ctx, *redirect.Output, mode)
		if err != nil {
			return err
		}
		capComm.AddResource(OutputIO, rp)

		if redirect.MergeErrorWithOutput {
			capComm.AddResource(ErrorIO, rp)
		}
	}

	if redirect.Error != nil {
		rp, err := capComm.createProviderFromOutputSpec(ctx, *redirect.Error, providers.IOModeError)
		if err != nil {
			return err
		}
		capComm.AddResource(ErrorIO, rp)
	}

	return nil
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
		Option("missingkey=zero").
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

	// Add in variables
	data[VariableTag] = capComm.variables.All()

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

func getParamFromURL(ctx context.Context, url string, optional bool) (string, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, timeOut)
	defer cancel()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", errors.Wrapf(err, "creating request for %s", url)
	}

	resp, err := ctxhttp.Do(ctxTimeout, nil, req)
	if err != nil {
		return "", errors.Wrapf(err, "getting %s", url)
	}
	defer resp.Body.Close()

	// Support optional gets
	if optional && resp.StatusCode == http.StatusNotFound {
		return "", nil
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		// Bad response
		return "", fmt.Errorf("response (%d) %s for %s", resp.StatusCode, resp.Status, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// expandParam carries out template expansion of a parameter.
func (capComm *CapComm) expandParam(ctx context.Context, param Param) (string, error) {
	// Read param
	value := param.Value

	// If param has a file name, open it
	if param.Path != "" {
		fileName, err := capComm.ExpandString(ctx, param.Name, param.Path)
		if err != nil {
			return "", errors.Wrap(err, "rexpanding file name")
		}

		b, err := os.ReadFile(fileName)
		if err != nil {
			if !os.IsNotExist(err) || !param.Optional {
				return "", errors.Wrap(err, "reading value from file")
			}
		} else {
			value += string(b)
		}
	}

	if param.URL != "" {
		// pull the data from the url
		body, err := getParamFromURL(ctx, param.URL, param.Optional)
		if err != nil {
			return "", err
		}

		value += body
	}

	// Skip expanding a param
	if param.SkipExpand {
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

// addMaps appends map data left to right to the template data receiver.
func (td TemplateData) addMaps(maps ...map[string]string) TemplateData {
	for _, m := range maps {
		for k, v := range m {
			td[k] = v
		}
	}

	return td
}

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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/nehemming/cirocket/pkg/buildinfo"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/providers"
	"github.com/nehemming/cirocket/pkg/resource"
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

	// MissionFileParamName is param name of the mission file.
	MissionFileParamName = "missionFile"

	// MissionBaseParamName is the base name of the mission file, without path or extensions.
	MissionBaseParamName = "missionBaseName"

	// MissionDirURLParamName is the full url to the directory containing the mission.
	MissionDirURLParamName = "missionDirURL"

	// MissionDirParamName is the file system path to the file.
	MissionDirParamName = "missionDir"

	// MissionDirAbsParamName is the file system absolute path to the file.
	MissionDirAbsParamName = "missionDirAbs"

	// MissionURLParamName is the full url to the mission file.
	MissionURLParamName = "missionURL"

	// WorkingDirectoryParamName is the name of the working dir.
	WorkingDirectoryParamName = "workingDir"
)

const timeOut = time.Second * 10

type (
	// TemplateData represents the data passed t a template.
	TemplateData map[string]interface{}

	// Runtime holds runtime data accessible to templates.
	Runtime struct {
		// GOOS is the operating system the host application is running against
		GOOS string
		// GOARCH is the architecture the host application is running against
		GOARCH string

		// UserName runtime username
		UserName string
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
		variables             *variableSet
		exportTo              *variableSet
		log                   loggee.Logger
	}
)

// setParamsFromMissionLocation adds config file entries to the environment map.
func setParamsFromMissionLocation(params map[string]string, missionLocation *url.URL) {
	working, _ := os.Getwd()

	missionDirURL := resource.GetURLParentLocation(missionLocation)
	file := path.Base(missionLocation.Path)
	var osMissionAbsDir string
	osMissionDir, err := resource.URLToRelativePath(missionDirURL)
	if err == nil && osMissionDir != "" {
		osMissionAbsDir, _ = filepath.Abs(osMissionDir)
	}

	params[WorkingDirectoryParamName] = working

	params[MissionURLParamName] = missionLocation.String()
	params[MissionDirURLParamName] = missionDirURL.String()
	params[MissionFileParamName] = file
	params[MissionBaseParamName] = strings.TrimSuffix(file, path.Ext(file))
	if osMissionAbsDir != "" {
		params[MissionDirParamName] = osMissionDir
		params[MissionDirAbsParamName] = osMissionAbsDir
	}
}

// NewCapComm returns a cap comm object that is suitable for using for testing.
func NewCapComm(missionLocation string, log loggee.Logger) *CapComm {
	url, err := resource.UltimateURL(missionLocation)
	if err != nil {
		panic(url)
	}
	return newCapCommFromEnvironment(url, log).Copy(false)
}

// newCapCommFromEnvironment creates a new capCom from the environment.
func newCapCommFromEnvironment(missionLocation *url.URL, log loggee.Logger) *CapComm {
	paramKvg := NewKeyValueGetter(nil)
	setParamsFromMissionLocation(paramKvg.kv, missionLocation)

	cc := &CapComm{
		sealed:                true,
		env:                   &osEnvGetter{},
		additionalMissionData: make(TemplateData),
		params:                paramKvg,
		funcMap:               initFuncMap(),
		runtime: Runtime{
			GOOS:     runtime.GOOS,
			GOARCH:   runtime.GOARCH,
			UserName: getUserName(),
		},
		resources: make(providers.ResourceProviderMap),
		variables: newVariableSet(),
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
		runtime:               capComm.runtime,
		resources:             capComm.resources.Copy(),
		variables:             newVariableSet(),
		exportTo:              capComm.variables,
		log:                   capComm.log,
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
	capComm.exportTo.Set(key, value)
	return capComm
}

// ExportVariables exports a list of variables if set in the capComm to aa parent.  If the variable is not found
// the exported name will be checked against the params and exported.
func (capComm *CapComm) ExportVariables(exports Exports) {
	for _, key := range exports {
		if value, ok := capComm.variables.Get(key); ok {
			capComm.exportTo.Set(key, value)
		} else if value := capComm.params.Get(key); value != "" {
			capComm.exportTo.Set(key, value)
		}
	}
}

func (capComm *CapComm) getTemplateVariables() map[string]string {
	m := capComm.variables.All()

	for k, v := range capComm.exportTo.All() {
		if _, ok := m[k]; !ok {
			m[k] = v
		}
	}

	return m
}

// GetVariable gets the value of a variable, first from local variables and then if not found from the exported parent variables.
func (capComm *CapComm) GetVariable(key string) (string, bool) {
	if value, ok := capComm.variables.Get(key); ok {
		return value, true
	}
	if value, ok := capComm.exportTo.Get(key); ok {
		return value, true
	}
	return "", false
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
		v, ok := capComm.GetVariable(inputSpec.Variable)
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
		panic("validation bad input runbook")
	}

	return rp, err
}

// InputSpecToResourceProvider creates a resource provider from the supplied input runbook.
func (capComm *CapComm) InputSpecToResourceProvider(ctx context.Context, inputSpec InputSpec) (providers.ResourceProvider, error) {
	if err := validateInputSpec(&inputSpec); err != nil {
		return nil, err
	}

	return capComm.createProviderFromInputSpec(ctx, inputSpec)
}

// AttachInputSpec adds a named input runbook to the capCom resources.
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

// AttachOutputSpec attaches an output runbook to the capComm.
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
		return errors.New("cannot both redirect to the log and also provide an output runbook")
	}
	if redirect.DirectError && redirect.Error != nil {
		return errors.New("cannot both redirect to stderr and also provide an error runbook")
	}
	if redirect.MergeErrorWithOutput && redirect.Error != nil {
		return errors.New("cannot merge errors with output and specify an error runbook")
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

// AttachRedirect attaches a redirection runbook to the capComm
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

	// safe type conversion as KeyValueGetter used all cases except root thats sealed sp not possible to be here
	kvg := capComm.params.(*KeyValueGetter)

	for index, p := range params {
		// exclude filtered
		if p.Filter.IsFiltered() {
			continue
		}

		if p.Name == "" {
			return fmt.Errorf("parameter %d has no name", index)
		}

		v, err := capComm.expandParam(ctx, p)
		if err != nil {
			return errors.Wrapf(err, "parameter %s", p.Name)
		}
		kvg.kv[p.Name] = v
	}

	return nil
}

// MergeTemplateEnvs adds params into an unsealed CapComm instance.
func (capComm *CapComm) MergeTemplateEnvs(ctx context.Context, env EnvMap) error {
	capComm.mustNotBeSealed()

	// Envs need to be expanded prior to merging
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

// ExpandBool expands a template expression and converts the result to a bool value.
func (capComm *CapComm) ExpandBool(ctx context.Context, name, value string) (bool, error) {
	v, err := capComm.ExpandString(ctx, name, value)
	if err != nil {
		return false, err
	}

	v = strings.Trim(v, " ")

	// default value if empty
	if v == "" {
		return false, nil
	}
	return strconv.ParseBool(v)
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

	// fix <no value> on nil see https://github.com/golang/go/issues/24963
	clean := strings.Replace(buf.String(), "<no value>", "", -1)

	// Finally expand any environment variables in the $VAR format
	return capComm.expandShellEnv(clean), nil
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
	data[VariableTag] = capComm.getTemplateVariables()

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

func (capComm *CapComm) readResource(ctx context.Context, name string,
	location string, optional bool) ([]byte, error) {
	location, err := capComm.ExpandString(ctx, name, location)
	if err != nil {
		return nil, errors.Wrap(err, "expanding")
	}

	b, err := resource.ReadResource(ctx, location)
	if err != nil {
		if resource.IsNotFoundError(err) == nil || !optional {
			return nil, errors.Wrap(err, "reading value")
		}

		// optional is empty resource
		return make([]byte, 0), nil
	}

	return b, nil
}

func (capComm *CapComm) getParamValue(ctx context.Context, param Param) (string, error) {
	// Read param
	value := param.Value

	// If param has a file name, open it
	if param.Path != "" {
		b, err := capComm.readResource(ctx, param.Name, param.Path, param.Optional)
		if err != nil {
			return "", err
		}
		value += string(b)
	}

	return value, nil
}

// expandParam carries out template expansion of a parameter.
func (capComm *CapComm) expandParam(ctx context.Context, param Param) (string, error) {
	value, err := capComm.getParamValue(ctx, param)
	if err != nil {
		return "", err
	}

	if !param.SkipExpand {
		// Expand
		value, err = capComm.ExpandString(ctx, param.Name, value)
		if err != nil {
			return "", err
		}
	}

	// Print?
	if param.Print {
		capComm.log.WithField("value", value).Infof("param: %s", param.Name)
	}

	return value, nil
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

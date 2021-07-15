package rocket

import (
	"context"
	"os"
	"runtime"
	"strings"
	"testing"
)

const testConfigFile = "testdir/file.yml"

func TestNewCapCommFromEnvironment(t *testing.T) {

	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	ctx := context.Background()

	capComm := newCapCommFromEnvironment(ctx, testConfigFile)

	if capComm.data != nil {
		t.Error("Unexpected non nil data")
	}

	if capComm.env == nil {
		t.Error("Unexpected nil env")
	}

	if capComm.env.Get("TEST_ENV_CAPCOMM") != "99" {
		t.Error("env not pulling from os env")
	}

	if capComm.env.Get("UNKNOWN_PARAM_FROM_OS") != "" {
		t.Error("env returning empty string for missing os env")
	}

	if capComm.entrustedParentEnv != nil {
		t.Error("Unexpected non nil entrustedParentEnv")
	}

	if capComm.funcMap == nil {
		t.Error("Unexpected nil funcMap")
	}

	if capComm.additionalMissionData == nil {
		t.Error("Unexpected nil additionalMissionData")
	}

	if len(capComm.additionalMissionData) != 0 {
		t.Error("Unexpected nil additionalMissionData len not 0", len(capComm.additionalMissionData))
	}

	if capComm.params == nil {
		t.Error("Unexpected nil params")
	}

	if capComm.runtime.GOARCH != runtime.GOARCH {
		t.Error("Unexpected GOARCH in runtime", capComm.runtime.GOARCH)
	}

	if capComm.runtime.GOOS != runtime.GOOS {
		t.Error("Unexpected GOOS in runtime", capComm.runtime.GOOS)
	}

	if !capComm.sealed {
		t.Error("Unexpected unsealed root")
	}

	if capComm.mission != nil {
		t.Error("Unexpected non nil mission")
	}

	if capComm.ioSettings == nil {
		t.Error("Unexpected nil ioSettings")
	}

	if capComm.params.Get(ConfigFileFullPath) == "" {
		t.Error("Base params missing ", ConfigFileFullPath)
	}
	if capComm.params.Get(ConfigFile) != "file.yml" {
		t.Error("Base params missing ", ConfigFile, capComm.params.Get(ConfigFile))
	}
	if capComm.params.Get(ConfigDir) != "testdir" {
		t.Error("Base params missing ", ConfigDir, capComm.params.Get(ConfigDir))
	}
	if capComm.params.Get(ConfigBaseName) != "file" {
		t.Error("Base params missing ", ConfigBaseName, capComm.params.Get(ConfigBaseName))
	}
}

func TestCopyTrusted(t *testing.T) {

	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	ctx := context.Background()
	root := newCapCommFromEnvironment(ctx, testConfigFile)
	capComm := root.Copy(false)

	if capComm.data != nil {
		t.Error("Unexpected non nil data")
	}

	if capComm.env == nil {
		t.Error("Unexpected nil env")
	}

	if capComm.env.Get("TEST_ENV_CAPCOMM") != "99" {
		t.Error("env not pulling from os env")
	}

	if capComm.entrustedParentEnv != nil {
		t.Error("Unexpected non nil entrustedParentEnv")
	}

	if capComm.funcMap == nil {
		t.Error("Unexpected nil funcMap")
	}

	if capComm.additionalMissionData == nil {
		t.Error("Unexpected nil additionalMissionData")
	}

	if len(capComm.additionalMissionData) != 0 {
		t.Error("Unexpected nil additionalMissionData len not 0", len(capComm.additionalMissionData))
	}

	if capComm.params == nil {
		t.Error("Unexpected nil params")
	}

	if capComm.runtime.GOARCH != runtime.GOARCH {
		t.Error("Unexpected GOARCH in runtime", capComm.runtime.GOARCH)
	}

	if capComm.runtime.GOOS != runtime.GOOS {
		t.Error("Unexpected GOOS in runtime", capComm.runtime.GOOS)
	}

	if capComm.sealed {
		t.Error("Unexpected sealed copy")
	}

	if capComm.mission != nil {
		t.Error("Unexpected non nil mission")
	}

	if capComm.ioSettings == nil {
		t.Error("Unexpected nil ioSettings")
	}

	capComm.Seal()

	if !capComm.sealed {
		t.Error("Unexpected unsealed post Seal call")
	}
}

func TestCopyNoTrust(t *testing.T) {

	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	ctx := context.Background()
	root := newCapCommFromEnvironment(ctx, testConfigFile)
	capComm := root.Copy(true)

	if capComm.data != nil {
		t.Error("Unexpected non nil data")
	}

	if capComm.env == nil {
		t.Error("Unexpected nil env")
	}

	if capComm.env.Get("TEST_ENV_CAPCOMM") != "" {
		t.Error("env pulling from os env")
	}

	if capComm.entrustedParentEnv == nil {
		t.Error("Unexpected nil entrustedParentEnv")
	}

	if capComm.entrustedParentEnv.Get("TEST_ENV_CAPCOMM") != "99" {
		t.Error("entrustedParentEnv not pulling from os env")
	}

	if capComm.funcMap == nil {
		t.Error("Unexpected nil funcMap")
	}

	if capComm.additionalMissionData == nil {
		t.Error("Unexpected nil additionalMissionData")
	}

	if len(capComm.additionalMissionData) != 0 {
		t.Error("Unexpected nil additionalMissionData len not 0", len(capComm.additionalMissionData))
	}

	if capComm.params == nil {
		t.Error("Unexpected nil params")
	}

	if capComm.runtime.GOARCH != runtime.GOARCH {
		t.Error("Unexpected GOARCH in runtime", capComm.runtime.GOARCH)
	}

	if capComm.runtime.GOOS != runtime.GOOS {
		t.Error("Unexpected GOOS in runtime", capComm.runtime.GOOS)
	}

	if capComm.sealed {
		t.Error("Unexpected sealed copy")
	}

	if capComm.mission != nil {
		t.Error("Unexpected non nil mission")
	}

	if capComm.ioSettings == nil {
		t.Error("Unexpected nil ioSettings")
	}
}

func TestWithMission(t *testing.T) {

	capComm := newCapCommFromEnvironment(context.Background(), testConfigFile).Copy(false)

	if capComm.mission != nil {
		t.Error("Unexpected non nil mission")
	}

	mission := &Mission{}

	capComm.WithMission(mission)

	if capComm.mission != mission {
		t.Error("Unexpected capComm mission mismatch")
	}

	copy := capComm.Copy(true)

	if copy.mission != mission {
		t.Error("Unexpected copy mission mismatch")
	}
}

func TestMergeBasicEnvMap(t *testing.T) {

	capComm := newCapCommFromEnvironment(context.Background(), testConfigFile).Copy(false)

	envMap := make(EnvMap)

	capComm.MergeBasicEnvMap(nil)

	envMap["something"] = "here"

	capComm.MergeBasicEnvMap(envMap)

	if capComm.env.Get("something") != "here" {
		t.Error("env getting something", capComm.env.Get("something"))
	}

	copyTrusted := capComm.Copy(false)

	if copyTrusted.env.Get("something") != "here" {
		t.Error("copyTrusted pulling something", capComm.env.Get("something"))
	}

	copyNoTrust := capComm.Copy(true)

	if copyNoTrust.env.Get("something") != "" {
		t.Error("copyTrusted has something it should nor", copyNoTrust.env.Get("something"))
	}
}

func TestAddAdditionalMissionData(t *testing.T) {

	capComm := newCapCommFromEnvironment(context.Background(), testConfigFile).Copy(false)

	capComm.AddAdditionalMissionData(nil)

	amd := make(TemplateData)
	amd["something"] = int(10)

	capComm.AddAdditionalMissionData(amd)

	x := capComm.additionalMissionData["something"]

	if v, ok := x.(int); !ok {
		t.Error("amd something wrong with type")
	} else if v != 10 {
		t.Error("amd something wrong value", v)
	}

	copyTrusted := capComm.Copy(false)

	x = copyTrusted.additionalMissionData["something"]

	if v, ok := x.(int); !ok {
		t.Error("amd something wrong with type")
	} else if v != 10 {
		t.Error("amd something wrong value", v)
	}

	copyUnTrusted := capComm.Copy(true)

	x = copyUnTrusted.additionalMissionData["something"]

	if v, ok := x.(int); !ok {
		t.Error("amd something wrong with type")
	} else if v != 10 {
		t.Error("amd something wrong value", v)
	}
}

func TestAddFile(t *testing.T) {

	envMap := make(EnvMap)
	envMap["something"] = "here"

	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(false).MergeBasicEnvMap(envMap)

	if err := capComm.AddFile(ctx, OutputIO, "test123", IOModeOutput|IOModeAppend); err != nil {
		t.Error("AddFile error", err)
	}

	if f, ok := capComm.ioSettings.files[OutputIO]; !ok {
		t.Error("No entry")
	} else if f.filePath != "test123" {
		t.Error("Output filePath wrong", f.filePath)
	}

	if err := capComm.AddFile(ctx, OutputIO, "test456", IOModeOutput|IOModeAppend); err != nil {
		t.Error("AddFile(2) error", err)
	}

	if f, ok := capComm.ioSettings.files[OutputIO]; !ok {
		t.Error("No entry")
	} else if f.FilePath() != "test456" {
		t.Error("Output filePath wrong (2)", f.filePath)
	}

	if err := capComm.AddFile(ctx, OutputIO, "test${something}", IOModeOutput|IOModeAppend); err != nil {
		t.Error("AddFile(3) error", err)
	}

	if f, ok := capComm.ioSettings.files[OutputIO]; !ok {
		t.Error("No entry")
	} else if f.filePath != "testhere" {
		t.Error("Output filePath wrong (3)", f.filePath)
	}

	if err := capComm.AddFile(ctx, OutputIO, "test--{{.Env.something}}", IOModeOutput|IOModeAppend); err != nil {
		t.Error("AddFile(4) error", err)
	}

	if f, ok := capComm.ioSettings.files[OutputIO]; !ok {
		t.Error("No entry")
	} else if f.filePath != "test--here" {
		t.Error("Output filePath wrong (4)", f.filePath)
	}

	// Check GetFileDetails
	fd := capComm.GetFileDetails(OutputIO)
	if fd == nil {
		t.Error("No fd entry")
	}
	if !fd.InMode(IOModeOutput | IOModeAppend) {
		t.Error("No fd file mode wrong", fd.fileMode)
	}
	if fd.filePath != "test--here" {
		t.Error("Output fd wrong", fd.filePath)
	}
}

func TestAttachOutputCreateMode(t *testing.T) {

	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(false)

	outSpec := OutputSpec{
		Output: "test1234",
	}

	if err := capComm.AttachOutput(ctx, outSpec); err != nil {
		t.Error("AttachOutput error", err)
	}

	fd := capComm.GetFileDetails(OutputIO)
	if fd == nil {
		t.Error("No fd entry")
	}
	if !fd.InMode(IOModeOutput | IOModeCreate) {
		t.Error("No fd file mode wrong", fd.fileMode)
	}
	if fd.filePath != "test1234" {
		t.Error("Output fd wrong", fd.filePath)
	}
}

func TestAttachOutputAppendMode(t *testing.T) {

	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(false)

	outSpec := OutputSpec{
		Output:       "test1234",
		AppendOutput: true,
	}

	if err := capComm.AttachOutput(ctx, outSpec); err != nil {
		t.Error("AttachOutput error", err)
	}

	fd := capComm.GetFileDetails(OutputIO)
	if fd == nil {
		t.Error("No fd entry")
	}
	if !fd.InMode(IOModeOutput | IOModeAppend) {
		t.Error("No fd file mode wrong", fd.fileMode)
	}
	if fd.filePath != "test1234" {
		t.Error("Output fd wrong", fd.filePath)
	}
}

func TestAttachRedirectNoMerge(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(false)

	redirect := Redirection{
		OutputSpec: OutputSpec{
			Output:       "test1234",
			AppendOutput: true,
		},
		Input: "in2020",
		Error: "sometimes",
	}

	if err := capComm.AttachRedirect(ctx, redirect); err != nil {
		t.Error("AttachRedirect error", err)
	}

	fdOut := capComm.GetFileDetails(OutputIO)
	if fdOut == nil {
		t.Error("No fdOut entry")
	}
	if !fdOut.InMode(IOModeOutput | IOModeAppend) {
		t.Error("No fdOut file mode wrong", fdOut.fileMode)
	}
	if fdOut.filePath != "test1234" {
		t.Error("Output fdOut wrong", fdOut.filePath)
	}

	fdError := capComm.GetFileDetails(ErrorIO)
	if fdError == nil {
		t.Error("No fdError entry")
	}
	if !fdError.InMode(IOModeError | IOModeCreate) {
		t.Error("No fdError file mode wrong", fdError.fileMode)
	}
	if fdError.filePath != "sometimes" {
		t.Error("Output fdError wrong", fdError.filePath)
	}

	fdIn := capComm.GetFileDetails(InputIO)
	if fdIn == nil {
		t.Error("No fdIn entry")
	}
	if !fdIn.InMode(IOModeInput) {
		t.Error("No fdIn file mode wrong", fdIn.fileMode)
	}
	if fdIn.filePath != "in2020" {
		t.Error("Output fdIn wrong", fdIn.filePath)
	}
}

func TestAttachRedirectErrorMerge(t *testing.T) {
	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(false)

	redirect := Redirection{
		OutputSpec: OutputSpec{
			Output:       "test1234",
			AppendOutput: true,
		},
		Input:                "in2020",
		Error:                "sometimes",
		MergeErrorWithOutput: true,
	}

	if err := capComm.AttachRedirect(ctx, redirect); err != nil {
		t.Error("AttachRedirect error", err)
	}

	fdOut := capComm.GetFileDetails(OutputIO)
	if fdOut == nil {
		t.Error("No fdOut entry")
	}
	if !fdOut.InMode(IOModeOutput | IOModeError | IOModeAppend) {
		t.Error("No fdOut file mode wrong", fdOut.fileMode)
	}
	if fdOut.filePath != "test1234" {
		t.Error("Output fdOut wrong", fdOut.filePath)
	}

	fdError := capComm.GetFileDetails(ErrorIO)
	if fdError == nil {
		t.Error("No fdError entry")
	}
	if !fdError.InMode(IOModeOutput | IOModeError | IOModeAppend) {
		t.Error("No fdError file mode wrong", fdError.fileMode)
	}
	if fdError.filePath != "test1234" {
		t.Error("Output fdError wrong", fdError.filePath)
	}

	fdIn := capComm.GetFileDetails(InputIO)
	if fdIn == nil {
		t.Error("No fdIn entry")
	}
	if !fdIn.InMode(IOModeInput) {
		t.Error("No fdIn file mode wrong", fdIn.fileMode)
	}
	if fdIn.filePath != "in2020" {
		t.Error("Output fdIn wrong", fdIn.filePath)
	}
}

func TestAttachRedirectExpand(t *testing.T) {
	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	envMap := make(EnvMap)
	envMap["something"] = "here"

	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(false).MergeBasicEnvMap(envMap)

	redirect := Redirection{
		OutputSpec: OutputSpec{
			Output:       "{{.Env.something}}-test1234",
			AppendOutput: false,
		},
		Input:       "in2020",
		Error:       "sometimes${TEST_ENV_CAPCOMM}",
		AppendError: true,
	}

	if err := capComm.AttachRedirect(ctx, redirect); err != nil {
		t.Error("AttachRedirect error", err)
	}

	fdOut := capComm.GetFileDetails(OutputIO)
	if fdOut == nil {
		t.Error("No fdOut entry")
	}
	if !fdOut.InMode(IOModeOutput | IOModeCreate) {
		t.Error("No fdOut file mode wrong", fdOut.fileMode)
	}
	if fdOut.filePath != "here-test1234" {
		t.Error("Output fdOut wrong", fdOut.filePath)
	}

	fdError := capComm.GetFileDetails(ErrorIO)
	if fdError == nil {
		t.Error("No fdError entry")
	}
	if !fdError.InMode(IOModeError | IOModeAppend) {
		t.Error("No fdError file mode wrong", fdError.fileMode)
	}
	if fdError.filePath != "sometimes99" {
		t.Error("Output fdError wrong", fdError.filePath)
	}

	fdIn := capComm.GetFileDetails(InputIO)
	if fdIn == nil {
		t.Error("No fdIn entry")
	}
	if !fdIn.InMode(IOModeInput) {
		t.Error("No fdIn file mode wrong", fdIn.fileMode)
	}
	if fdIn.filePath != "in2020" {
		t.Error("Output fdIn wrong", fdIn.filePath)
	}
}

func TestMergeParams(t *testing.T) {

	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(false)

	if err := capComm.MergeParams(ctx, nil); err != nil {
		t.Error("nil MergeParams error", err)
	}

	params := make([]Param, 0)

	if err := capComm.MergeParams(ctx, params); err != nil {
		t.Error("params (0) MergeParams error", err)
	}

	params = append(params, Param{
		Name:  "",
		Value: "test",
	})

	if err := capComm.MergeParams(ctx, params); err == nil {
		t.Error("params MergeParams no name no error", err)
	}

	params[0].Name = "test"
	if err := capComm.MergeParams(ctx, params); err != nil {
		t.Error("params name single MergeParams error", err)
	}

	params[0].Value = ""
	if err := capComm.MergeParams(ctx, params); err != nil {
		t.Error("params value empty MergeParams error", err)
	}

	params[0].Value = "{{.Env.TEST_ENV_CAPCOMM}}"
	if err := capComm.MergeParams(ctx, params); err != nil {
		t.Error("params value template expand MergeParams error", err)
	}
}

func TestMergeParamsWithFile(t *testing.T) {

	envMap := make(EnvMap)
	envMap["name"] = "config"

	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(false).MergeBasicEnvMap(envMap)

	params := make([]Param, 0)
	params = append(params, Param{
		Name: "test",
		File: "{{.Env.name}}.go",
	})

	if err := capComm.MergeParams(ctx, params); err != nil {
		t.Error("params MergeParams error", err)
	}

	data := capComm.params.Get("test")
	if !strings.HasPrefix(data, "package rocket") {
		t.Error("has the package been renamed?", data)
	}
}

func TestMergeParamsWithOptionalFile(t *testing.T) {

	envMap := make(EnvMap)
	envMap["name"] = "notaconfig"

	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(false).MergeBasicEnvMap(envMap)

	params := make([]Param, 0)
	params = append(params, Param{
		Name:         "test",
		File:         "{{.Env.name}}.go",
		FileOptional: true,
	})

	if err := capComm.MergeParams(ctx, params); err != nil {
		t.Error("params MergeParams error", err)
	}

	data := capComm.params.Get("test")
	if len(data) != 0 {
		t.Error("found an unexpected file", data)
	}
}

func TestMergeParamsWithSkipTemplate(t *testing.T) {

	envMap := make(EnvMap)
	envMap["name"] = "notaconfig"

	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(false).MergeBasicEnvMap(envMap)

	params := make([]Param, 0)
	params = append(params, Param{
		Name:         "test",
		Value:        "{{.Env.name}}.go",
		SkipTemplate: true,
	})

	if err := capComm.MergeParams(ctx, params); err != nil {
		t.Error("params MergeParams error", err)
	}

	data := capComm.params.Get("test")
	if data != "{{.Env.name}}.go" {
		t.Error("remplate expanded", data)
	}
}

func TestMergeParamsWithParam(t *testing.T) {

	envMap := make(EnvMap)
	envMap["name"] = "config"

	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(false).MergeBasicEnvMap(envMap)

	params := make([]Param, 0)
	params = append(params, Param{
		Name:  "test",
		Value: "{{.Env.name}}.go",
	})

	if err := capComm.MergeParams(ctx, params); err != nil {
		t.Error("params MergeParams error", err)
	}

	data := capComm.params.Get("test")
	if data != "config.go" {
		t.Error("data is unexpected", data)
	}

	//new params
	params = make([]Param, 0)
	params = append(params, Param{
		Name:  "gogo",
		Value: "{{.test}}.go",
	})

	if err := capComm.MergeParams(ctx, params); err != nil {
		t.Error("params MergeParams (2) error", err)
	}

	data = capComm.params.Get("gogo")
	if data != "config.go.go" {
		t.Error("expanded param data is unexpected", data)
	}
}

func TestMergeTemplateEnvs(t *testing.T) {

	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(false)

	envMap := make(EnvMap)

	capComm.MergeBasicEnvMap(nil)

	envMap["something"] = "here"
	envMap["outher"] = "{{.Env.TEST_ENV_CAPCOMM}}"

	if err := capComm.MergeTemplateEnvs(ctx, envMap); err != nil {
		t.Error("MergeTemplateEnvs error", err)
	}

	if capComm.env.Get("something") != "here" {
		t.Error("env getting something", capComm.env.Get("something"))
	}
	if capComm.env.Get("outher") != "99" {
		t.Error("env getting outher", capComm.env.Get("outher"))
	}
}

func TestFuncMap(t *testing.T) {

	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(false)

	fm := capComm.FuncMap()
	if len(fm) != 1 {
		t.Error("Functions in func map", len(fm))
	}
}

func TestFGetExecEnvNoOSInherit(t *testing.T) {

	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	ctx := context.Background()
	execEnv := newCapCommFromEnvironment(ctx, testConfigFile).Copy(true).GetExecEnv()

	if len(execEnv) != 0 {
		t.Error("execEnv len not 0", len(execEnv))
	}
}

func TestGetExecEnv(t *testing.T) {

	envMap := make(EnvMap)
	envMap["TEST_ENV_CAPCOMM"] = "99"
	ctx := context.Background()
	execEnv := newCapCommFromEnvironment(ctx, testConfigFile).Copy(true).MergeBasicEnvMap(envMap).GetExecEnv()

	if len(execEnv) != 1 {
		t.Error("execEnv len not 1", len(execEnv))
	} else if execEnv[0] != "TEST_ENV_CAPCOMM=99" {
		t.Error("Sing env wrong", execEnv[0])
	}
}

func TestIsFiltered(t *testing.T) {

	ctx := context.Background()
	capComm := newCapCommFromEnvironment(ctx, testConfigFile).Copy(true)

	if capComm.isFiltered(nil) != false {
		t.Error("Nil should not filter")
	}

	// Included testing
	f := &Filter{}
	if capComm.isFiltered(f) != false {
		t.Error("Empty should not filter")
	}
	f.IncludeOS = []string{runtime.GOOS}
	if capComm.isFiltered(f) != false {
		t.Error("Same Os should not filter")
	}
	f.IncludeArch = []string{runtime.GOARCH}
	if capComm.isFiltered(f) != false {
		t.Error("Same Arch should not filter")
	}

	// Not included testing
	f = &Filter{}
	f.IncludeOS = []string{runtime.GOOS + "nope"}
	if capComm.isFiltered(f) != true {
		t.Error("Diff Os should filter")
	}
	f.IncludeArch = []string{runtime.GOARCH + "nope"}
	if capComm.isFiltered(f) != true {
		t.Error("Diff Arch should filter")
	}

	// Excluded test
	f = &Filter{}
	f.ExcludeOS = []string{runtime.GOOS}
	if capComm.isFiltered(f) != true {
		t.Error("Same Os should exclude filter")
	}
	f = &Filter{}
	f.ExcludeArch = []string{runtime.GOARCH}
	if capComm.isFiltered(f) != true {
		t.Error("Same Arch should exclude filter")
	}

	// Non exclude test
	f = &Filter{}
	f.ExcludeOS = []string{runtime.GOOS + "nope"}
	if capComm.isFiltered(f) != false {
		t.Error("Diff Os should exclude filter")
	}
	f = &Filter{}
	f.ExcludeArch = []string{runtime.GOARCH + "nope"}
	if capComm.isFiltered(f) != false {
		t.Error("Diff Arch should exclude filter")
	}
}

func TestMustNotBeSealed(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Error("No panic")
		}
	}()

	newCapCommFromEnvironment(context.Background(), testConfigFile).mustNotBeSealed()

}

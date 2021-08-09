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
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
	"github.com/nehemming/cirocket/pkg/providers"
	"github.com/nehemming/cirocket/pkg/resource"
)

const testMissionFile = "testdir/file.yml"

func getTestMissionFile() *url.URL {
	// helper to get a url for the mission file
	u, e := resource.UltimateURL(testMissionFile)
	if e != nil {
		panic(e)
	}
	return u
}

func TestCapCommConfigFile(t *testing.T) {
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())

	if capComm.params.Get(MissionURLParamName) == "" {
		t.Error("Base params missing ", MissionURLParamName)
	}
	if capComm.params.Get(MissionFileParamName) != "file.yml" {
		t.Error("Base params missing ", MissionFileParamName, capComm.params.Get(MissionFileParamName))
	}
	if capComm.params.Get(MissionDirParamName) != "testdir" {
		t.Error("Base params missing ", MissionDirParamName, capComm.params.Get(MissionDirParamName))
	}
	if capComm.params.Get(MissionBaseParamName) != "file" {
		t.Error("Base params missing ", MissionBaseParamName, capComm.params.Get(MissionBaseParamName))
	}
}

func TestCapCommRuntime(t *testing.T) {
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())

	if capComm.runtime.GOARCH != runtime.GOARCH {
		t.Error("Unexpected GOARCH in runtime", capComm.runtime.GOARCH)
	}

	if capComm.runtime.GOOS != runtime.GOOS {
		t.Error("Unexpected GOOS in runtime", capComm.runtime.GOOS)
	}
}

func TestCapCommEnv(t *testing.T) {
	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())

	if capComm.env == nil {
		t.Error("Unexpected nil env")
	}

	if capComm.env.Get("TEST_ENV_CAPCOMM") != "99" {
		t.Error("env not pulling from os env")
	}

	if capComm.env.Get("UNKNOWN_PARAM_FROM_OS") != "" {
		t.Error("env returning empty string for missing os env")
	}
}

func TestNewCapCommFromTestLog(t *testing.T) {
	l := stdlog.New()
	capComm := newCapCommFromEnvironment(getTestMissionFile(), l)

	if capComm.Log() != l {
		t.Error("log issue")
	}
}

func TestNewCapCommFromEnvironment(t *testing.T) {
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())

	if capComm.data != nil {
		t.Error("Unexpected non nil data")
	}

	if len(capComm.additionalMissionData) != 0 {
		t.Error("Unexpected nil additionalMissionData len not 0", len(capComm.additionalMissionData))
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

	if capComm.params == nil {
		t.Error("Unexpected nil params")
	}

	if !capComm.sealed {
		t.Error("Unexpected unsealed root")
	}

	if capComm.mission != nil {
		t.Error("Unexpected non nil mission")
	}

	if capComm.resources == nil {
		t.Error("Unexpected nil resources")
	}
}

func TestNewCapCommFromEnvironmentResources(t *testing.T) {
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())

	if capComm.resources == nil {
		t.Error("Unexpected nil resources")
	}

	if len(capComm.resources) != 5 {
		t.Error("should have 6 resources", len(capComm.resources))
	}
}

func TestNewCapCommFromEnvironmentVariables(t *testing.T) {
	capComm := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())

	if capComm.exportTo != nil {
		t.Error("exportTo should be nil")
	}

	if capComm.variables == nil {
		t.Error("Unexpected nil variables")
	}
}

func TestCapCommCopyEnvTrusted(t *testing.T) {
	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	root := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	capComm := root.Copy(false)

	if capComm.env == nil {
		t.Error("Unexpected nil env")
	}

	if capComm.env.Get("TEST_ENV_CAPCOMM") != "99" {
		t.Error("env not pulling from os env")
	}

	if capComm.env.Get("UNKNOWN_PARAM_FROM_OS") != "" {
		t.Error("env returning empty string for missing os env")
	}

	if capComm.exportTo == nil {
		t.Error("exportTo should not be nil")
	}
}

func TestCapCommCopyEnvNotTrusted(t *testing.T) {
	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	root := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	capComm := root.Copy(true)

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

	if capComm.exportTo == nil {
		t.Error("exportTo should not be nil")
	}
}

func TestCapCommRuntimeCopyTrusted(t *testing.T) {
	root := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	capComm := root.Copy(false)

	if capComm.runtime.GOARCH != runtime.GOARCH {
		t.Error("Unexpected GOARCH in runtime", capComm.runtime.GOARCH)
	}

	if capComm.runtime.GOOS != runtime.GOOS {
		t.Error("Unexpected GOOS in runtime", capComm.runtime.GOOS)
	}
}

func TestCapCommRuntimeCopyNotTrusted(t *testing.T) {
	root := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	capComm := root.Copy(true)

	if capComm.runtime.GOARCH != runtime.GOARCH {
		t.Error("Unexpected GOARCH in runtime", capComm.runtime.GOARCH)
	}

	if capComm.runtime.GOOS != runtime.GOOS {
		t.Error("Unexpected GOOS in runtime", capComm.runtime.GOOS)
	}
}

func TestCapCommCopyUnSealed(t *testing.T) {
	root := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	capComm := root.Copy(false)

	if capComm.sealed {
		t.Error("Unexpected sealed copy")
	}

	capComm.Seal()

	if !capComm.sealed {
		t.Error("Unexpected unsealed post Seal call")
	}
}

func TestCopyTrusted(t *testing.T) {
	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	root := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	capComm := root.Copy(false)

	if capComm.data != nil {
		t.Error("Unexpected non nil data")
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

	if capComm.mission != nil {
		t.Error("Unexpected non nil mission")
	}

	if len(capComm.resources) != 5 {
		t.Error("Unexpected resources", len(capComm.resources))
	}
}

func TestCopyNoTrust(t *testing.T) {
	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	root := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New())
	capComm := root.Copy(true)

	if capComm.data != nil {
		t.Error("Unexpected non nil data")
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

	if capComm.sealed {
		t.Error("Unexpected sealed copy")
	}

	if capComm.mission != nil {
		t.Error("Unexpected non nil mission")
	}

	if len(capComm.resources) != 5 {
		t.Error("Unexpected resources", len(capComm.resources))
	}
}

func TestWithMission(t *testing.T) {
	capComm := NewCapComm(testMissionFile, stdlog.New())

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
	capComm := NewCapComm(testMissionFile, stdlog.New())

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
	capComm := NewCapComm(testMissionFile, stdlog.New())

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

func TestAddFileResource(t *testing.T) { //nolint -- keep as one test

	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	if err := capComm.AddFileResource(ctx, OutputIO, "test123", providers.IOModeOutput|providers.IOModeAppend); err != nil {
		t.Error("AddFileResource error", err)
	}

	var outputProvider providers.ResourceProvider
	if f, ok := capComm.resources[OutputIO]; !ok {
		t.Error("No entry OutputIO")
	} else {
		outputProvider = f
	}

	if err := capComm.AddFileResource(ctx, OutputIO, "test456", providers.IOModeOutput|providers.IOModeAppend); err != nil {
		t.Error("AddFileResource(2) error", err)
	}

	if f, ok := capComm.resources[OutputIO]; !ok {
		t.Error("No entry OutputIO (2)")
	} else if outputProvider == f {
		t.Error("Resource not replaced")
	}
}

func TestAddFileResourceExpansion(t *testing.T) {
	envMap := make(EnvMap)
	envMap["something"] = "io"
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New()).MergeBasicEnvMap(envMap)

	if err := capComm.AddFileResource(ctx, InputIO, "${something}.go", providers.IOModeInput); err != nil {
		t.Error("AddFileResource error", err)
	}

	if f, ok := capComm.resources[InputIO]; !ok {
		t.Error("No entry")
	} else {
		r, err := f.OpenRead(ctx)
		if err != nil {
			t.Error("OpenRead error, missing file expansion?", err)
			return
		}

		defer r.Close()

		b, err := io.ReadAll(r)

		if b == nil || err != nil {
			t.Error("OpenRead error, data?", err)
		}
	}
}

func TestGetResource(t *testing.T) {
	capComm := NewCapComm(testMissionFile, stdlog.New())

	for _, r := range []providers.ResourceID{
		OutputIO, ErrorIO, Stdin, Stdout, Stderr,
	} {
		res := capComm.GetResource(r)
		if res == nil {
			t.Error("No entry", r)
		}
	}

	if inRes := capComm.GetResource(InputIO); inRes != nil {
		t.Error("Input defined")
	}
}

func TestAttachOutputSpecFileCreateMode(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	outSpec := OutputSpec{
		Path: "test1234",
	}

	if err := capComm.AttachOutputSpec(ctx, OutputIO, outSpec); err != nil {
		t.Error("AttachOutput error", err)
	}

	res := capComm.GetResource(OutputIO)
	if res == nil {
		t.Error("No res entry")
	}

	if fd, ok := res.(providers.FileDetailer); !ok {
		t.Error("No res is not a file")
	} else {
		if !fd.InMode(providers.IOModeOutput | providers.IOModeTruncate) {
			t.Error("No fd file mode wrong", fd.IOMode())
		}
		path, _ := resource.URLToRelativePath(fd.URL())
		if path != "test1234" {
			t.Error("Output fd wrong", fd.FilePath(), fd.URL())
		}
	}
}

func TestAttachOutputSpecFileAppendMode(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	outSpec := OutputSpec{
		Path:   "test1234",
		Append: true,
	}

	if err := capComm.AttachOutputSpec(ctx, OutputIO, outSpec); err != nil {
		t.Error("AttachOutput error", err)
	}

	res := capComm.GetResource(OutputIO)
	if res == nil {
		t.Error("No res entry")
		return
	}

	fd := res.(providers.FileDetailer)

	if !fd.InMode(providers.IOModeOutput | providers.IOModeAppend) {
		t.Error("No fd io mode wrong", fd.IOMode())
	}
	path, _ := resource.URLToRelativePath(fd.URL())
	if path != "test1234" {
		t.Error("Output fd wrong", fd.FilePath(), fd.URL())
	}
}

func TestAttachRedirectNoMergeStdOut(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	redirect := Redirection{
		Output: &OutputSpec{
			Path:   "test1234",
			Append: true,
		},
		Input: &InputSpec{
			Path: "in2020",
		},
		Error: &OutputSpec{
			Path: "sometimes",
		},
	}

	if err := capComm.AttachRedirect(ctx, redirect); err != nil {
		t.Error("AttachRedirect error", err)
	}

	resOut := capComm.GetResource(OutputIO)
	if resOut == nil {
		t.Error("No resOut")
	}

	fdOut := resOut.(providers.FileDetailer)
	if !fdOut.InMode(providers.IOModeOutput | providers.IOModeAppend) {
		t.Error("No fdOut io mode wrong", fdOut.IOMode())
	}
	path, _ := resource.URLToRelativePath(fdOut.URL())
	if path != "test1234" {
		t.Error("Output fdOut wrong", fdOut.FilePath(), fdOut.URL())
	}
}

func TestAttachRedirectNoMergeStdErr(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	redirect := Redirection{
		Output: &OutputSpec{
			Path:   "test1234",
			Append: true,
		},
		Input: &InputSpec{
			Path: "in2020",
		},
		Error: &OutputSpec{
			Path: "sometimes",
		},
	}

	if err := capComm.AttachRedirect(ctx, redirect); err != nil {
		t.Error("AttachRedirect error", err)
	}

	resErr := capComm.GetResource(ErrorIO)
	if resErr == nil {
		t.Error("No resErr")
	}

	fdError := resErr.(providers.FileDetailer)
	if !fdError.InMode(providers.IOModeError | providers.IOModeTruncate) {
		t.Error("No fdError io mode wrong", fdError.IOMode())
	}
	path, _ := resource.URLToRelativePath(fdError.URL())
	if path != "sometimes" {
		t.Error("Output fdError wrong", fdError.FilePath(), fdError.URL())
	}
}

func TestAttachRedirectNoMergeStdIn(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	redirect := Redirection{
		Output: &OutputSpec{
			Path:   "test1234",
			Append: true,
		},
		Input: &InputSpec{
			Path: "in2020",
		},
		Error: &OutputSpec{
			Path: "sometimes",
		},
	}

	if err := capComm.AttachRedirect(ctx, redirect); err != nil {
		t.Error("AttachRedirect error", err)
	}

	resIn := capComm.GetResource(InputIO)
	if resIn == nil {
		t.Error("No resIn")
	}

	fdIn := resIn.(providers.FileDetailer)
	if !fdIn.InMode(providers.IOModeInput) {
		t.Error("No fdIn io mode wrong", fdIn.IOMode())
	}
	path, _ := resource.URLToRelativePath(fdIn.URL())
	if path != "in2020" {
		t.Error("Output fdIn wrong", fdIn.FilePath(), fdIn.URL())
	}
}

func TestAttachRedirectErrorMergeStdOut(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	redirect := Redirection{
		Output: &OutputSpec{
			Path:   "test1234",
			Append: true,
		},
		Input: &InputSpec{
			Path: "in2020",
		},
		MergeErrorWithOutput: true,
	}

	if err := capComm.AttachRedirect(ctx, redirect); err != nil {
		t.Error("AttachRedirect error", err)
	}

	resOut := capComm.GetResource(OutputIO)
	if resOut == nil {
		t.Error("No resOut")
	}
	fdOut := resOut.(providers.FileDetailer)
	if !fdOut.InMode(providers.IOModeOutput | providers.IOModeError | providers.IOModeAppend) {
		t.Error("No fdOut error io mode wrong", fdOut.IOMode())
	}

	resErr := capComm.GetResource(ErrorIO)
	if resErr != resOut {
		t.Error("resErr different from resOut")
	}
}

func TestAttachRedirectErrorMergeStdErr(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	redirect := Redirection{
		DirectError: true,
	}

	if err := capComm.AttachRedirect(ctx, redirect); err != nil {
		t.Error("AttachRedirect error", err)
	}

	resErr := capComm.GetResource(ErrorIO)
	if resErr == nil {
		t.Error("No resErr")
	}
	resStdErr := capComm.GetResource(Stderr)
	if resStdErr == nil {
		t.Error("No resStdErr")
	}

	if resErr != resStdErr {
		t.Error("resErr different from resStdErr")
	}
}

func TestAttachRedirectOutputExpand(t *testing.T) {
	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	envMap := make(EnvMap)
	envMap["something"] = "here"

	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New()).MergeBasicEnvMap(envMap)

	redirect := Redirection{
		Output: &OutputSpec{
			Path: "{{.Env.something}}-test1234",
		},
	}

	if err := capComm.AttachRedirect(ctx, redirect); err != nil {
		t.Error("AttachRedirect error", err)
	}

	resOut := capComm.GetResource(OutputIO)
	if resOut == nil {
		t.Error("No resOut")
	}
	fdOut := resOut.(providers.FileDetailer)
	if !fdOut.InMode(providers.IOModeOutput | providers.IOModeTruncate) {
		t.Error("No fdOut io mode wrong", fdOut.IOMode())
	}

	path, _ := resource.URLToRelativePath(fdOut.URL())
	if path != "here-test1234" {
		t.Error("Output fdOut wrong", fdOut.FilePath(), fdOut.URL())
	}
}

func TestAttachRedirectErrorExpand(t *testing.T) {
	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	envMap := make(EnvMap)
	envMap["something"] = "here"

	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New()).MergeBasicEnvMap(envMap)

	redirect := Redirection{
		Error: &OutputSpec{
			Path:   "sometimes${TEST_ENV_CAPCOMM}",
			Append: true,
		},
	}

	if err := capComm.AttachRedirect(ctx, redirect); err != nil {
		t.Error("AttachRedirect error", err)
	}

	resErr := capComm.GetResource(ErrorIO)
	if resErr == nil {
		t.Error("No resErr")
	}

	fdError := resErr.(providers.FileDetailer)
	if !fdError.InMode(providers.IOModeError | providers.IOModeAppend) {
		t.Error("No fdError io mode wrong", fdError.IOMode())
	}

	path, _ := resource.URLToRelativePath(fdError.URL())
	if path != "sometimes99" {
		t.Error("Output fdError wrong", fdError.FilePath(), fdError.URL())
	}
}

func TestAttachRedirectInExpand(t *testing.T) {
	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	redirect := Redirection{
		Input: &InputSpec{
			Inline: "{{.Env.TEST_ENV_CAPCOMM}}-test1234",
		},
	}

	if err := capComm.AttachRedirect(ctx, redirect); err != nil {
		t.Error("AttachRedirect error", err)
	}

	resIn := capComm.GetResource(InputIO)
	if resIn == nil {
		t.Error("No resIn")
	}

	r, err := resIn.OpenRead(ctx)
	if err != nil {
		t.Error("open inline", err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		t.Error("read inline", err)
		return
	}

	if string(b) != "99-test1234" {
		t.Error("string mismatch", string(b))
	}
}

func TestAttachRedirectInExpandNotSet(t *testing.T) {
	envMap := make(EnvMap)
	envMap["something"] = "here"

	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New()).MergeBasicEnvMap(envMap)

	redirect := Redirection{
		Input: &InputSpec{
			Inline: "{{.Env.TEST_ENV_CAPCOMM}}-test1234",
		},
	}

	if err := capComm.AttachRedirect(ctx, redirect); err != nil {
		t.Error("AttachRedirect error", err)
	}

	resIn := capComm.GetResource(InputIO)
	if resIn == nil {
		t.Error("No resIn")
	}

	r, err := resIn.OpenRead(ctx)
	if err != nil {
		t.Error("open inline", err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		t.Error("read inline", err)
		return
	}

	if string(b) != "-test1234" {
		t.Error("string mismatch", string(b))
	}
}

func TestMergeParams(t *testing.T) {
	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

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
	envMap["name"] = "io"

	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New()).MergeBasicEnvMap(envMap)

	params := make([]Param, 0)
	params = append(params, Param{
		Name: "test",
		Path: "{{.Env.name}}.go",
	})

	if err := capComm.MergeParams(ctx, params); err != nil {
		t.Error("params MergeParams error", err)
	}

	data := capComm.params.Get("test")
	if !strings.HasPrefix(data, "/*") {
		t.Error("has the package been renamed?", data)
	}
}

func TestMergeParamsWithOptionalFile(t *testing.T) {
	envMap := make(EnvMap)
	envMap["name"] = "notaconfig"

	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New()).MergeBasicEnvMap(envMap)

	params := make([]Param, 0)
	params = append(params, Param{
		Name:     "test",
		Path:     "{{.Env.name}}.go",
		Optional: true,
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
	capComm := NewCapComm(testMissionFile, stdlog.New()).MergeBasicEnvMap(envMap)

	params := make([]Param, 0)
	params = append(params, Param{
		Name:       "test",
		Value:      "{{.Env.name}}.go",
		SkipExpand: true,
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
	capComm := NewCapComm(testMissionFile, stdlog.New()).MergeBasicEnvMap(envMap)

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

	// new params
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
	capComm := NewCapComm(testMissionFile, stdlog.New())

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
	capComm := NewCapComm(testMissionFile, stdlog.New())

	fm := capComm.FuncMap()
	if len(fm) != 10 {
		t.Error("Functions in func map", len(fm))
	}
}

func TestGetExecEnvNoOSInherit(t *testing.T) {
	os.Setenv("TEST_ENV_CAPCOMM", "99")
	defer os.Unsetenv("TEST_ENV_CAPCOMM")

	execEnv := NewCapComm(testMissionFile, stdlog.New()).Copy(true).GetExecEnv()

	if len(execEnv) != 0 {
		t.Error("execEnv len not 0", len(execEnv))
	}
}

func TestGetExecEnv(t *testing.T) {
	envMap := make(EnvMap)
	envMap["TEST_ENV_CAPCOMM"] = "99"

	execEnv := newCapCommFromEnvironment(getTestMissionFile(), stdlog.New()).
		Copy(true).
		MergeBasicEnvMap(envMap).GetExecEnv()

	if len(execEnv) != 1 {
		t.Error("execEnv len not 1", len(execEnv))
	} else if execEnv[0] != "TEST_ENV_CAPCOMM=99" {
		t.Error("Sing env wrong", execEnv[0])
	}
}

func TestMustNotBeSealed(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("No panic")
		}
	}()

	newCapCommFromEnvironment(getTestMissionFile(), stdlog.New()).mustNotBeSealed()
}

func TestIndentTemplateFunc(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	s, err := capComm.ExpandString(ctx, "spacing", `{{"\nhello"|indent 6}}
	`)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if !strings.HasPrefix(s, "\n      hello") {
		t.Error("indent missing", s)
	}
}

func TestIndentFirstLineTemplateFunc(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	s, err := capComm.ExpandString(ctx, "spacing", `{{"hello"|indent 6}}
	`)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if !strings.HasPrefix(s, "hello") {
		t.Error("indent present", s)
	}
}

func TestExportVariable(t *testing.T) {
	capComm := NewCapComm(testMissionFile, stdlog.New())

	if len(capComm.variables) != 0 {
		t.Error("pre existing vars", len(capComm.variables))
	}

	taskCapCom := capComm.Copy(true)

	taskCapCom.ExportVariable("hello", "there")

	if v, ok := capComm.variables["hello"]; !ok || v != "there" {
		t.Error("var not exported", ok, v)
	}
}

func TestValidateInputSpecEmpty(t *testing.T) {
	inputSpec := &InputSpec{}

	if err := validateInputSpec(inputSpec); err == nil || err.Error() != "no input source was specified" {
		t.Error("input runbook empty check fails")
	}
}

func TestValidateInputSpecMultiple(t *testing.T) {
	for i, r := range []InputSpec{
		{
			Inline: "-",
			Path:   "-",
		},
		{
			Inline:   "-",
			Variable: "-",
		},
		{
			Inline: "-",
			URL:    "-",
		},
		{
			Path:     "-",
			Variable: "-",
		},
		{
			Path: "-",
			URL:  "-",
		},
		{
			Variable: "-",
			URL:      "-",
		},
	} {
		if err := validateInputSpec(&r); err == nil || err.Error() != "more than one input source was specified, only one is permitted" {
			t.Error("input runbook multi check fails", i, r)
		}
	}
}

func TestCreateProviderFromInputSpec(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	capComm.ExportVariable("test_it", "hello")

	for i, r := range []InputSpec{
		{
			Inline: "12345",
		},
		{
			Path: "testdata/six.yml",
		},
		{
			URL: "https://raw.githubusercontent.com/nehemming/cirocket/master/README.md",
		},
		{
			Variable: "test_it",
		},
	} {
		rp, err := capComm.createProviderFromInputSpec(ctx, r)
		if err != nil {
			t.Error("unexpected error", i, err)
		}

		_, err = rp.OpenWrite(ctx)
		if err == nil {
			t.Error("open write", i)
			return
		}

		r, err := rp.OpenRead(ctx)
		if err != nil {
			t.Error("error open read", i, err)
			return
		}
		r.Close()
	}
}

func TestAttachInputSpec(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	capComm.ExportVariable("test_it", "hello")

	var rp providers.ResourceProvider
	for i, r := range []InputSpec{
		{
			Inline:     "12345",
			SkipExpand: true,
		},
		{
			Inline: "12345",
		},
		{
			Path: "testdata/six.yml",
		},
		{
			Path:       "testdata/six.yml",
			SkipExpand: true,
		},
		{
			URL: "https://raw.githubusercontent.com/nehemming/cirocket/master/README.md",
		},
		{
			Variable: "test_it",
		},
	} {
		if err := capComm.AttachInputSpec(ctx, "test", r); err != nil {
			t.Error("unexpected error", i, err)
		}

		rpNext := capComm.GetResource("test")

		if rpNext == rp {
			t.Error("resource update issue", i)
		}

		rp = rpNext
	}
}

func TestValidateOutputSpecEmpty(t *testing.T) {
	outputSpec := &OutputSpec{}

	if err := validateOutputSpec(outputSpec); err == nil || err.Error() != "no output source was specified" {
		t.Error("output runbook empty check fails")
	}
}

func TestValidateOutputSpecMultiple(t *testing.T) {
	for i, r := range []OutputSpec{
		{
			Path:     "-",
			Variable: "-",
		},
	} {
		if err := validateOutputSpec(&r); err == nil || err.Error() != "more than one output source was specified, only one is permitted" {
			t.Error("output runbook multi check fails", i, r)
		}
	}
}

func TestCreateProviderFromoutputSpec(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())
	capComm.ExportVariable("test_it", "hello")

	defer func() {
		_ = os.Remove(filepath.FromSlash("testdata/dummy.tmp"))
	}()

	for i, r := range []OutputSpec{
		{
			Path: "testdata/dummy.tmp",
		},
		{
			Path:       "testdata/dummy.tmp",
			SkipExpand: true,
		},
		{
			Variable: "test_it",
		},
	} {
		rp, err := capComm.createProviderFromOutputSpec(ctx, r, providers.IOModeOutput)
		if err != nil {
			t.Error("unexpected error", i, err)
		}
		_, err = rp.OpenRead(ctx)
		if err == nil {
			t.Error("open read", i)
			return
		}

		w, err := rp.OpenWrite(ctx)
		if err != nil {
			t.Error("error open write", i, err)
			return
		}

		w.Close()
	}

	if capComm.exportTo["test_it"] != "" {
		t.Error("Variable not set")
	}
}

func TestValidateRedirection(t *testing.T) {
	for i, r := range []Redirection{
		{
			LogOutput: true,
			Output: &OutputSpec{
				Path: "testdata/dummy.tmp",
			},
		},
		{
			DirectError: true,
			Error: &OutputSpec{
				Path: "testdata/dummy.tmp",
			},
		},
		{
			MergeErrorWithOutput: true,
			Error: &OutputSpec{
				Path: "testdata/dummy.tmp",
			},
		},
		{
			Error: &OutputSpec{},
		},
		{
			Output: &OutputSpec{},
		},
		{
			Input: &InputSpec{},
		},
	} {
		err := validateRedirection(&r)
		if err == nil {
			t.Error("should fail validation", i)
			return
		}
	}
}

func TestGetParamFromURLSuccess(t *testing.T) {
	url := "https://raw.githubusercontent.com/nehemming/cirocket/master/CREDITS"

	data, err := getParamFromURL(context.Background(), url, false)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if len(data) == 0 {
		t.Error("no data")
	}
}

func TestGetParamFromURLMissingError(t *testing.T) {
	url := "https://raw.githubusercontent.com/nehemming/cirocket/master/notknown"

	_, err := getParamFromURL(context.Background(), url, false)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetParamFromURLOptionalSuccess(t *testing.T) {
	url := "https://raw.githubusercontent.com/nehemming/cirocket/master/notknown"

	data, err := getParamFromURL(context.Background(), url, true)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if len(data) != 0 {
		t.Error("data")
	}
}

func TestExpandParam(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())
	capComm.ExportVariable("test_it", "hello")

	for i, r := range []Param{
		{
			Name: "fileTest",
			Path: "testdata/six.yml",
		},
		{
			Name:  "valueTest",
			Value: "1234",
			Print: true,
		},
		{
			Name: "valueTest",
			Path: "https://raw.githubusercontent.com/nehemming/cirocket/master/CREDITS",
		},
	} {
		v, err := capComm.expandParam(ctx, r)
		if err != nil {
			t.Error("unexpected", i, r.Name, err)
		}

		if len(v) == 0 {
			t.Error("zero data", i, r.Name)
		}
	}
}

func TestAttachRedirectLogOutput(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	redirect := Redirection{
		LogOutput:            true,
		MergeErrorWithOutput: true,
	}

	if err := capComm.AttachRedirect(ctx, redirect); err != nil {
		t.Error("AttachRedirect error", err)
	}
}

func TestStringExpandNoValueIssue(t *testing.T) {
	capComm := NewCapComm(testMissionFile, stdlog.New())
	ctx := context.Background()

	s, err := capComm.ExpandString(ctx, "string", "{{.notfound}}test")
	if s != "test" || err != nil {
		t.Error("unexpected", s, err)
	}
}

func TestExpandAdjacentMergeParamsExpandsInOrder(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	params := []Param{
		{
			Name:  "one",
			Value: "here",
		},
		{
			Name:  "two",
			Value: "{{.one}}",
			Print: true,
		},
	}

	err := capComm.MergeParams(ctx, params)
	if err != nil {
		t.Error("unexpected", err)
	}

	v := capComm.params.Get("two")

	if v != "here" {
		t.Error("unexpected v", v)
	}
}

func TestParamsFilter(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	params := []Param{
		{
			Name:  "one",
			Value: "here",
		},
		{
			Name:   "two",
			Value:  "two",
			Print:  true,
			Filter: &Filter{Skip: true},
		},
	}

	err := capComm.MergeParams(ctx, params)
	if err != nil {
		t.Error("unexpected", err)
	}

	v := capComm.params.Get("two")

	if v != "" {
		t.Error("unexpected v filtered", v)
	}
}

func TestParamsFilterDuplicated(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	params := []Param{
		{
			Name:  "one",
			Value: "here",
		},
		{
			Name:  "two",
			Value: "this one",
			Print: true,
		},
		{
			Name:   "two",
			Value:  "two",
			Print:  true,
			Filter: &Filter{Skip: true},
		},
	}

	err := capComm.MergeParams(ctx, params)
	if err != nil {
		t.Error("unexpected", err)
	}

	v := capComm.params.Get("two")

	if v != "this one" {
		t.Error("unexpected v picked wrong one", v)
	}
}

func TestDirName(t *testing.T) {
	if d := dirname("a/b/c"); d != "a/b" {
		t.Error("unexpected", d)
	}
}

func TestDirNameURL(t *testing.T) {
	if d := dirname("file:///a/b/c"); d != "file:/a/b" {
		t.Error("unexpected", d)
	}
}

func TestDirNameFilePath(t *testing.T) {
	if d := dirname(filepath.Join("a", "b", "c")); d != "a/b" {
		t.Error("unexpected", d)
	}
}

func TestBaseName(t *testing.T) {
	if d := basedname("a/b/c"); d != "c" {
		t.Error("unexpected", d)
	}
}

func TestBaseNameURL(t *testing.T) {
	if d := basedname("file:///a/b/c"); d != "c" {
		t.Error("unexpected", d)
	}
}

func TestBaseNameFilePath(t *testing.T) {
	if d := basedname(filepath.Join("a", "b", "c.txt")); d != "c.txt" {
		t.Error("unexpected", d)
	}
}

func TestUltimate(t *testing.T) {
	if runtime.GOOS == "windows" {
		return // skip as covered in resources
	}

	d, e := ultimate("/a/b/c")

	if d != "file:///a/b/c" || e != nil {
		t.Error("unexpected", d, e)
	}
}

func TestBUltimateURL(t *testing.T) {
	if d, e := ultimate("file:///a/b/c"); d != "file:///a/b/c" || e != nil {
		t.Error("unexpected", d, e)
	}
}

func TestUltimateFilePath(t *testing.T) {
	if d, e := ultimate(filepath.Join("a", "b", "c.txt")); !strings.HasSuffix(d, "/a/b/c.txt") || e != nil {
		t.Error("unexpected", d, e)
	}
}

/*


func ultimate(p ...string) (string, error) {
	u, e := resource.UltimateURL(p...)
	if e != nil {
		return "", e
	}
	return u.String(), nil
}*/

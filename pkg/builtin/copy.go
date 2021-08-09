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

package builtin

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/pkg/errors"
)

type (
	AbsRel struct {
		Abs string
		Rel string
	}

	DestSpec struct {
		Path  string
		IsDir bool
	}

	// Copy task is used to run a specific task.
	Copy struct {
		Sources     []string `mapstructure:"sources"`
		Destination string   `mapstructure:"destination"`
		Log         bool     `mapstructure:"log"`
	}

	copyType struct{}
)

func (copyType) Type() string {
	return "copy"
}

// Prepare loads the tasks configuration and returns the operation function or an error.
func (copyType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {
	copyCfg := &Copy{}

	if err := mapstructure.Decode(task.Definition, copyCfg); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	fn := func(execCtx context.Context) error {
		// get the destination
		dest, err := capComm.ExpandString(ctx, "dir", copyCfg.Destination)
		if err != nil {
			return errors.Wrapf(err, "expanding dest %s", copyCfg.Destination)
		}

		// get the destination spec from the raw path entered
		destSpec, err := getDestSpec(dest)
		if err != nil {
			return errors.Wrapf(err, "dest %s", copyCfg.Destination)
		}

		// Expand files
		rawSpecs := make([]string, 0, len(copyCfg.Sources))
		for index, f := range copyCfg.Sources {
			rawSpec, err := capComm.ExpandString(ctx, "source", f)
			if err != nil {
				return errors.Wrapf(err, "expanding file %d", index)
			}
			rawSpecs = append(rawSpecs, rawSpec)
		}

		// glob the files
		files, err := globFileAbsRel(rawSpecs...)
		if err != nil {
			return err
		}

		// copy
		return copyFiles(execCtx, files, destSpec, copyCfg.Log)
	}

	return fn, nil
}

func getDestSpec(rawDest string) (destSpec DestSpec, err error) {
	if rawDest == "" {
		return destSpec, errors.New("destination cannot be blank")
	}

	var isDir bool
	clean, err := filepath.Abs(filepath.FromSlash(rawDest))
	if err != nil {
		return destSpec, err
	}
	stat, err := os.Stat(clean)

	if err == nil && stat.IsDir() {
		isDir = true
	} else if clean == "." || strings.HasSuffix(filepath.ToSlash(rawDest), "/") {
		isDir = true
	}

	return DestSpec{Path: clean, IsDir: isDir}, nil
}

func getBaseDir(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	for dir != "" {

		dir = filepath.Dir(dir)
		if !strings.ContainsAny(dir, "*?") {
			break
		}
	}

	return dir, nil
}

func appendAbsRelList(files []AbsRel, spec string, list ...string) ([]AbsRel, error) {
	dir, err := getBaseDir(spec)
	if err != nil {
		return nil, errors.Wrapf(err, "dir walk %s", spec)
	}

	for _, l := range list {

		stat, err := os.Stat(l)
		if err != nil {
			return nil, errors.Wrapf(err, "stat %s", l)
		}
		if stat.IsDir() {
			// skip dirs
			continue
		}

		abs, err := filepath.Abs(l)
		if err != nil {
			return nil, errors.Wrapf(err, "abs %s", l)
		}

		rel, err := filepath.Rel(dir, abs)
		if err != nil {
			return nil, errors.Wrapf(err, "rel dir of %s", l)
		}

		files = append(files, AbsRel{Abs: abs, Rel: rel})
	}

	return files, nil
}

func globFileAbsRel(rawSpecs ...string) (files []AbsRel, err error) {
	for _, rawSpec := range rawSpecs {

		clean := filepath.Clean(filepath.FromSlash(rawSpec))
		list, err := recursiveGlob(clean)
		if err != nil {
			return nil, errors.Wrapf(err, "globbing %s", clean)
		}

		files, err = appendAbsRelList(files, clean, list...)
		if err != nil {
			return nil, err
		}
	}

	return toDistinctAbsRelSlice(files...), nil
}

func toDistinctAbsRelSlice(files ...AbsRel) []AbsRel {
	// Make list distinct, as more than one ref may match
	m := make(map[string]bool)
	res := make([]AbsRel, 0, len(files))
	for _, v := range files {
		if m[v.Rel] {
			continue
		}
		m[v.Rel] = true
		res = append(res, v)
	}
	return res
}

func copyFiles(ctx context.Context, sources []AbsRel, dest DestSpec, log bool) error {
	for _, source := range sources {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err := copyFile(source, dest, log); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(source AbsRel, dest DestSpec, log bool) error {
	// Get the source files permission
	stat, err := os.Stat(source.Abs)
	if err != nil {
		return errors.Wrapf(err, "stat %s:", source.Abs)
	}

	// open reader
	srcFile, err := os.Open(source.Abs)
	if err != nil {
		return errors.Wrapf(err, "open %s:", source.Abs)
	}
	defer srcFile.Close()

	var finalPath string
	if dest.IsDir {
		finalPath = filepath.Join(dest.Path, source.Rel)
	} else {
		finalPath = dest.Path
	}
	dir := filepath.Dir(finalPath)
	err = os.MkdirAll(dir, 0777)

	if err != nil {
		return errors.Wrapf(err, "dir %s:", dir)
	}

	destFile, err := os.OpenFile(finalPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, stat.Mode())
	if err != nil {
		return errors.Wrapf(err, "dest %s:", finalPath)
	}
	defer destFile.Close()

	destRel, err := filepath.Rel(dest.Path, finalPath)
	if err != nil {
		return errors.Wrapf(err, "dest %s, rel %s:", dest.Path, finalPath)
	}

	// Do copy
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return errors.Wrapf(err, "copy %s => %s:", source.Rel, destRel)
	}

	// log
	if log {
		loggee.Infof("copy %s => %s", source.Rel, destRel)
	}

	return nil
}

func init() {
	rocket.Default().RegisterTaskTypes(copyType{})
}

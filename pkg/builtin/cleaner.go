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
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/pkg/errors"
)

type (

	// Remove lists files or directories to remove
	Remove struct {
		Files []string `mapstructure:"files"`
		Log   bool     `mapstructure:"log"`
	}

	removeType  struct{}
	cleanerType struct{}
)

func (removeType) Type() string {
	return "remove"
}

func (removeType) Description() string {
	return "deletes files matching on of the file glob specs."
}

func (cleanerType) Type() string {
	return "cleaner"
}

func (cleanerType) Description() string {
	return "cleans up files matching on of the file glob specs."
}

// Prepare prepares the cleaner task and returns an operation or an error.
func (cleanerType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {
	removeType := &Remove{}

	if err := mapstructure.WeakDecode(task.Definition, removeType); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	return buildRemoveOp(capComm, removeType)
}

// Prepare loads the tasks configuration and returns the operation function or an error.
func (removeType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {
	removeType := &Remove{}

	if err := mapstructure.WeakDecode(task.Definition, removeType); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	return buildRemoveOp(capComm, removeType)
}

func buildRemoveOp(capComm *rocket.CapComm, config *Remove) (rocket.ExecuteFunc, error) {
	fn := func(execCtx context.Context) error {
		// Expand files
		specs := make([]string, 0, len(config.Files))
		for index, f := range config.Files {
			fileSpec, err := capComm.ExpandString(execCtx, "file", f)
			if err != nil {
				return errors.Wrapf(err, "expanding file %d", index)
			}
			specs = append(specs, fileSpec)
		}

		// glob the files
		files, err := globFile(specs)
		if err != nil {
			return err
		}

		// clean
		return deleteFiles(files, config.Log)
	}

	return fn, nil
}

func globFile(specs []string) ([]string, error) {
	files := make([]string, 0, len(specs))
	for _, spec := range specs {
		list, err := recursiveGlob(spec)
		if err != nil {
			return nil, errors.Wrapf(err, "globbing %s", spec)
		}
		files = append(files, list...)
	}

	return toDistinctStrings(files...), nil
}

func deleteFiles(files []string, log bool) error {
	for _, file := range files {
		stat, err := os.Stat(filepath.FromSlash(file))
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrapf(err, "stat %s:", file)
		}

		if stat == nil {
			continue
		}

		if stat.IsDir() {
			err = os.RemoveAll(filepath.FromSlash(file))
		} else {
			err = os.Remove(filepath.FromSlash(file))
		}

		if err != nil {
			return errors.Wrapf(err, "rm %s:", file)
		}

		// log
		if log {
			loggee.Infof("removed %s", friendlyRelativePath(file))
		}
	}

	return nil
}

func toDistinctStrings(files ...string) []string {
	// Make list distinct, as more than one ref may match
	m := make(map[string]bool)
	res := make([]string, 0, len(files))
	for _, v := range files {
		if m[v] {
			continue
		}
		m[v] = true
		res = append(res, v)
	}
	return res
}

func init() {
	rocket.Default().RegisterTaskTypes(cleanerType{}, removeType{})
}

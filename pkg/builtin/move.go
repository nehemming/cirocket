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
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/pkg/errors"
)

type (

	// Move task is used to move files.  The tool can recursively move files.
	// Directories cannot be specified as a source, the wild card ** pattern can
	// be used to recursively move the contents of the directory.
	// The directory structure is moved, but the source sub directories are not deleted.
	Move struct {
		Sources     []string `mapstructure:"sources"`
		Destination string   `mapstructure:"destination"`
		Overwrite   string   `mapstructure:"overwrite"`
		Log         bool     `mapstructure:"log"`
	}

	moveType struct{}
)

func (moveType) Type() string {
	return "move"
}

func (moveType) Description() string {
	return "moves files matching the source glob specs to the destination folder."
}

// Prepare loads the tasks configuration and returns the operation function or an error.
func (moveType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {
	moveCfg := &Move{}

	if err := mapstructure.WeakDecode(task.Definition, moveCfg); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	fn := func(execCtx context.Context) error {
		// get the destination
		dest, err := capComm.ExpandString(ctx, "dir", moveCfg.Destination)
		if err != nil {
			return errors.Wrapf(err, "expanding dest %s", moveCfg.Destination)
		}

		// get the overrite value
		overwrite, err := capComm.ExpandBool(ctx, "overrite", moveCfg.Overwrite)
		if err != nil {
			return errors.Wrapf(err, "expanding dest %s", moveCfg.Destination)
		}

		// get the destination spec from the raw path entered
		destSpec, err := getDestSpec(dest)
		if err != nil {
			return errors.Wrapf(err, "dest %s", moveCfg.Destination)
		}

		// Expand files
		rawSpecs := make([]string, 0, len(moveCfg.Sources))
		for index, f := range moveCfg.Sources {
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
		return moveFiles(execCtx, files, destSpec, overwrite, getLogFromCapComm(capComm, moveCfg.Log))
	}

	return fn, nil
}

func moveFiles(ctx context.Context, sources []AbsRel, dest DestSpec, allowOverwrite bool, log loggee.Logger) error {
	for _, source := range sources {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err := moveFile(source, dest, allowOverwrite, log); err != nil {
			return err
		}
	}
	return nil
}

func moveFile(source AbsRel, dest DestSpec, allowOverwrite bool, log loggee.Logger) error {
	// Get the source files permission
	stat, err := os.Stat(source.Abs)
	if err != nil {
		return errors.Wrapf(err, "stat %s:", source.Abs)
	}

	if stat.IsDir() {
		// cannot move dirs, move as dir/**
		return fmt.Errorf("%s is a directory and cannot be moved, replace source with %[1]s/**", source.Rel)
	}

	destAbsRel, err := prepDestination(source, dest, allowOverwrite)
	if err != nil {
		return err
	}
	if destAbsRel == nil || source.Abs == destAbsRel.Abs {
		// skipping
		if log != nil {
			log.Infof("skipping %s", source.Rel)
		}
		return nil
	}

	// Do the move
	err = os.Rename(source.Abs, destAbsRel.Abs)
	if err != nil {
		return errors.Wrapf(err, "move %s => %s:", source.Rel, destAbsRel.Rel)
	}

	// log
	if log != nil {
		log.Infof("move %s => %s", source.Rel, destAbsRel.Rel)
	}

	return nil
}

func init() {
	rocket.Default().RegisterTaskTypes(moveType{})
}

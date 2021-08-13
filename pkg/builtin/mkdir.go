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
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/pkg/errors"
)

type (
	// MkDir task is used create directories.
	MkDir struct {
		Dirs []string `mapstructure:"dirs"`
		Log  bool     `mapstructure:"log"`
	}

	mkDirType struct{}
)

func (mkDirType) Type() string {
	return "mkdir"
}

func (mkDirType) Description() string {
	return "creates directories as needed from the dirs list."
}

func (mkDirType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {
	mkDirCfg := &MkDir{}

	if err := mapstructure.Decode(task.Definition, mkDirCfg); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	fn := func(execCtx context.Context) error {
		// Expand directories
		directories := make([]string, 0, len(mkDirCfg.Dirs))
		for index, f := range mkDirCfg.Dirs {
			dir, err := capComm.ExpandString(ctx, "dir", f)
			if err != nil {
				return errors.Wrapf(err, "expanding dir %d", index)
			}

			directories = append(directories, dir)
		}
		// create
		for _, dir := range directories {
			err := os.MkdirAll(filepath.FromSlash(dir), 0777)
			if err != nil {
				return errors.Wrapf(err, "mkdir -p %s:", dir)
			}

			// log
			if mkDirCfg.Log {
				capComm.Log().Infof("created %s", dir)
			}
		}

		return nil
	}

	return fn, nil
}

func init() {
	rocket.Default().RegisterTaskTypes(mkDirType{})
}

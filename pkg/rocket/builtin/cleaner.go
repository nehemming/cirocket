package builtin

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/pkg/errors"
)

type (
	// Cleaner task is used to run a specific task
	Cleaner struct {
		Files []string `mapstructure:"files"`
		Log   bool     `mapstructure:"log"`
	}

	cleanerType struct{}
)

func (cleanerType) Type() string {
	return "cleaner"
}

func (cleanerType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {

	cleanCfg := &Cleaner{}

	if err := mapstructure.Decode(task.Definition, cleanCfg); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	// Expand files
	specs := make([]string, 0, len(cleanCfg.Files))
	for index, f := range cleanCfg.Files {
		if fileSpec, err := capComm.ExpandString(ctx, "file", f); err != nil {
			return nil, errors.Wrapf(err, "expanding file %d", index)
		} else {
			specs = append(specs, fileSpec)
		}
	}

	fn := func(execCtx context.Context) error {

		// glob the files
		files := make([]string, 0, len(specs))
		for _, spec := range specs {
			if strings.ContainsAny(spec, "*?") {
				if list, err := filepath.Glob(spec); err != nil {
					return errors.Wrapf(err, "globbing %s", spec)
				} else {
					files = append(files, list...)
				}
			} else {
				files = append(files, spec)
			}
		}

		// clean
		for _, file := range files {
			if stat, err := os.Stat(file); err != nil && !os.IsNotExist(err) {
				return errors.Wrapf(err, "stat %s:", file)
			} else {
				var err error
				if stat.IsDir() {
					err = os.RemoveAll(file)
				} else {
					err = os.Remove(file)
				}

				if err != nil {
					return errors.Wrapf(err, "rm %s:", file)
				}

				// log
				if cleanCfg.Log {
					loggee.Infof("removed %s", file)
				}
			}
		}

		return nil
	}

	return fn, nil

}

func init() {
	rocket.Default().RegisterTaskTypes(cleanerType{})
}

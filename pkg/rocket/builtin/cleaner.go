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
	// Cleaner task is used to run a specific task.
	Cleaner struct {
		Files []string `mapstructure:"files"`
		Log   bool     `mapstructure:"log"`
	}

	cleanerType struct{}
)

func (cleanerType) Type() string {
	return "cleaner"
}

// Prepare loads the tasks configuration and returns the operation function or an error.
func (cleanerType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {
	cleanCfg := &Cleaner{}

	if err := mapstructure.Decode(task.Definition, cleanCfg); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	fn := func(execCtx context.Context) error {
		// Expand files
		specs := make([]string, 0, len(cleanCfg.Files))
		for index, f := range cleanCfg.Files {
			fileSpec, err := capComm.ExpandString(ctx, "file", f)
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
		return deleteFiles(files, cleanCfg.Log)
	}

	return fn, nil
}

func globFile(specs []string) ([]string, error) {
	files := make([]string, 0, len(specs))
	for _, spec := range specs {
		if strings.ContainsAny(spec, "*?") {
			list, err := filepath.Glob(spec)
			if err != nil {
				return nil, errors.Wrapf(err, "globbing %s", spec)
			}
			files = append(files, list...)
		} else {
			files = append(files, spec)
		}
	}

	return files, nil
}

func deleteFiles(files []string, log bool) error {
	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrapf(err, "stat %s:", file)
		}

		if stat == nil {
			continue
		}

		if stat.IsDir() {
			err = os.RemoveAll(file)
		} else {
			err = os.Remove(file)
		}

		if err != nil {
			return errors.Wrapf(err, "rm %s:", file)
		}

		// log
		if log {
			loggee.Infof("removed %s", file)
		}
	}

	return nil
}

func init() {
	rocket.Default().RegisterTaskTypes(cleanerType{})
}

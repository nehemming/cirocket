package builtin

import (
	"context"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/pkg/errors"
)

type (
	// Cleaner task is used to run a specific task
	MkDir struct {
		Dirs []string `mapstructure:"dirs"`
		Log  bool     `mapstructure:"log"`
	}

	mkDirType struct{}
)

func (mkDirType) Type() string {
	return "mkdir"
}

func (mkDirType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {

	mkDirCfg := &MkDir{}

	if err := mapstructure.Decode(task.Definition, mkDirCfg); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	// Expand directories
	directories := make([]string, 0, len(mkDirCfg.Dirs))
	for index, f := range mkDirCfg.Dirs {
		if dir, err := capComm.ExpandString(ctx, "dir", f); err != nil {
			return nil, errors.Wrapf(err, "expanding dir %d", index)
		} else {
			directories = append(directories, dir)
		}
	}

	fn := func(execCtx context.Context) error {

		// create
		for _, dir := range directories {
			err := os.MkdirAll(dir, 0777)
			if err != nil {
				return errors.Wrapf(err, "mkdir -p %s:", dir)
			}

			// log
			if mkDirCfg.Log {
				loggee.Infof("created %s", dir)
			}
		}

		return nil
	}

	return fn, nil
}

func init() {
	rocket.Default().RegisterTaskTypes(mkDirType{})
}

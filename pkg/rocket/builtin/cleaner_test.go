package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
	"github.com/nehemming/cirocket/pkg/rocket"
)

func TestCleanerType(t *testing.T) {
	var rt cleanerType

	if rt.Type() != "cleaner" {
		t.Error("Wrong cleaner type", rt.Type())
	}
}

func TestCleanerRun(t *testing.T) {

	loggee.SetLogger(stdlog.New())

	mc := rocket.NewMissionControl()
	RegisterAll(mc)

	// create some data
	for i := 0; i < 5; i++ {
		dir := fmt.Sprintf("%s-%d", filepath.Join("testdata", "clean"), i)
		// check clean
		_ = os.RemoveAll(dir)

		if err := os.MkdirAll(dir, 0777); err != nil {
			panic(err)
		}

		for f := 0; f < 5; f++ {
			name := fmt.Sprintf("file-%d", f)
			fn := filepath.Join(dir, name)
			if err := os.WriteFile(fn, []byte("hello"), 0666); err != nil {
				panic(err)
			}
		}
	}

	mission, cfgFile := loadMission("cleaner")

	if err := mc.FlyMission(context.Background(), cfgFile, mission); err != nil {
		t.Error("failure", err)
	}

	// Check and clean
	for i := 0; i < 5; i++ {
		dir := fmt.Sprintf("%s-%d", filepath.Join("testdata", "clean"), i)

		spec := filepath.Join(dir, "*")

		if i != 3 {
			// need 5 files
			if _, err := os.Stat(dir); err != nil {
				t.Error("dir missing", dir, err)
			}

			files, err := filepath.Glob(spec)
			if err != nil {
				t.Error("error listing", dir, err)
			}

			count := len(files)

			if i == 0 {
				if count != 5 {
					t.Error("Dir mismatch", dir, count)
				}
			} else if i == 1 {
				if count != 4 {
					t.Error("Dir mismatch", dir, count)
				}
			} else if count != 0 {
				t.Error("Dir mismatch", dir, count)
			}
		} else if _, err := os.Stat(dir); err == nil {
			t.Error("dir present", dir)
		}

		_ = os.RemoveAll(dir)
	}
}

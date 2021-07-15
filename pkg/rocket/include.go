package rocket

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"golang.org/x/net/context/ctxhttp"
	"gopkg.in/yaml.v2"
)

const timeOut = time.Duration(time.Second * 10)

func loadMapFromUrl(ctx context.Context, url string) (map[string]interface{}, error) {
	if url == "" {
		return nil, errors.New("url empty")
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, timeOut)
	defer cancel()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "creating request for %s", url)
	}

	resp, err := ctxhttp.Do(ctxTimeout, nil, req)
	if err != nil {
		return nil, errors.Wrapf(err, "getting %s", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		// Bad response
		return nil, fmt.Errorf("response (%d) %s for %s", resp.StatusCode, resp.Status, url)
	}

	m := make(map[string]interface{})
	err = yaml.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func getConfigFileName(configFile string) (string, error) {

	if configFile == "" {
		if dir, err := os.Getwd(); err != nil {
			return "", err
		} else {
			return filepath.Join(dir, "default"), nil
		}
	}

	return configFile, nil
}

func mergePaths(base, rel string) string {

	if path.IsAbs(rel) {
		return rel
	}

	return path.Join(base, rel)
}

func loadMapFromPath(path, basePath string) (map[string]interface{}, string, error) {
	if path == "" {
		return nil, path, errors.New("path empty")
	}

	path = mergePaths(basePath, path)

	fh, err := os.Open(path)
	if err != nil {
		return nil, path, err
	}
	defer fh.Close()

	m := make(map[string]interface{})
	err = yaml.NewDecoder(fh).Decode(&m)
	if err != nil {
		return nil, path, err
	}

	return m, path, nil
}

func loadPreMissionMaps(ctx context.Context, spaceDust map[string]interface{}, configFile string) ([]map[string]interface{}, error) {
	// Premission
	preMission := &PreMission{}

	cfgMaps := make([]map[string]interface{}, 0)

	// Load in the mission from the spaceDust
	if d, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			Result:           preMission,
		}); err != nil {
		return nil, errors.Wrap(err, "setting up premission decoder")
	} else if err := d.Decode(spaceDust); err != nil {
		return nil, errors.Wrap(err, "parsing mission premission profile")
	}

	cfgMaps = append(cfgMaps, preMission.Mission)

	// No includes, exit here
	if len(preMission.Includes) == 0 {
		return cfgMaps, nil
	}

	// Get the base path of the config file so includes are relative to it
	basePath := filepath.Dir(configFile)

	for index, include := range preMission.Includes {

		if include.Path != "" && include.Url != "" {
			return nil, fmt.Errorf("include[%d] has both url and path specified", index)
		}

		if include.Path != "" {
			path := os.ExpandEnv(include.Path)

			if m, cfgFile, err := loadMapFromPath(path, basePath); err != nil {
				return nil, errors.Wrapf(err, "include[%d]", index)
			} else {
				// if m has its own includes need to load its includes too
				if _, ok := m["includes"]; ok {
					// need too get its
					if pm, err := loadPreMissionMaps(ctx, m, cfgFile); err != nil {
						return nil, errors.Wrapf(err, "include[%d]", index)
					} else {
						cfgMaps = append(cfgMaps, pm...)
					}
				} else {
					cfgMaps = append(cfgMaps, m)
				}
			}
		} else if include.Url != "" {
			url := os.ExpandEnv(include.Url)
			if m, err := loadMapFromUrl(ctx, url); err != nil {
				return nil, errors.Wrapf(err, "include[%d]", index)
			} else {
				cfgMaps = append(cfgMaps, m)
			}
		} else {
			return nil, fmt.Errorf("include[%d] has neither url nor path specified", index)
		}
	}

	return cfgMaps, nil

}

func mergeMissions(mission, addition *Mission) {

	if addition.Name != "" {
		mission.Name = addition.Name
	}
	if addition.Version != "" {
		mission.Version = addition.Version
	}
	if len(addition.BasicEnv) > 0 {
		if mission.BasicEnv == nil {
			mission.BasicEnv = make(EnvMap)
		}
		for k, v := range addition.BasicEnv {
			mission.BasicEnv[k] = v
		}
	}
	if len(addition.Env) > 0 {
		if mission.Env == nil {
			mission.Env = make(EnvMap)
		}
		for k, v := range addition.Env {
			mission.Env[k] = v
		}
	}
	if len(addition.Params) > 0 {
		if mission.Params == nil {
			mission.Params = make([]Param, 0, len(addition.Params))
		}
		mission.Params = append(mission.Params, addition.Params...)
	}
	if len(addition.Stages) > 0 {
		if mission.Stages == nil {
			mission.Stages = make([]Stage, 0, len(addition.Stages))
		}
		mission.Stages = append(mission.Stages, addition.Stages...)
	}

	if len(addition.Sequences) > 0 {
		if mission.Sequences == nil {
			mission.Sequences = make(map[string][]string)
		}
		for k, v := range addition.Sequences {
			mission.Sequences[k] = v
		}
	}
}

func loadPreMission(ctx context.Context, spaceDust map[string]interface{}, configFile string) (*Mission, error) {

	// Load all the includes as config maps
	cfgMaps, err := loadPreMissionMaps(ctx, spaceDust, configFile)
	if err != nil {
		return nil, err
	}

	mission := &Mission{}

	// iterate through config maps loading their missions, merging as we go
	for _, cfgMap := range cfgMaps {

		partialMission := &Mission{}

		// Load in the mission from the spaceDust
		if d, err := mapstructure.NewDecoder(
			&mapstructure.DecoderConfig{
				WeaklyTypedInput: true,
				Result:           partialMission,
			}); err != nil {
			return nil, errors.Wrap(err, "setting up decoder")
		} else if err := d.Decode(cfgMap); err != nil {
			return nil, errors.Wrap(err, "parsing mission profile")
		}

		mergeMissions(mission, partialMission)
	}

	// Setup name
	if mission.Name == "" {
		mission.Name = filepath.Base(configFile)
	}

	return mission, nil
}

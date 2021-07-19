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

const timeOut = time.Second * 10

func loadMapFromURL(ctx context.Context, url string) (map[string]interface{}, error) {
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
		dir, err := os.Getwd()
		if err != nil {
			return "", err
		}

		return filepath.Join(dir, "default"), nil
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

func decodeSpaceDust(spaceDust map[string]interface{}) ([]map[string]interface{}, []Include, error) {
	preMission := &PreMission{}
	cfgMaps := make([]map[string]interface{}, 0)

	// Load in the mission from the spaceDust
	if d, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			Result:           preMission,
		}); err != nil {
		return nil, nil, errors.Wrap(err, "setting up pre-mission decoder")
	} else if err := d.Decode(spaceDust); err != nil {
		return nil, nil, errors.Wrap(err, "parsing mission pre-mission profile")
	}

	cfgMaps = append(cfgMaps, preMission.Mission)

	return cfgMaps, preMission.Includes, nil
}

func importURLInclude(ctx context.Context, index int, include Include, cfgMaps []map[string]interface{}) ([]map[string]interface{}, error) {
	url := os.ExpandEnv(include.URL)
	m, err := loadMapFromURL(ctx, url)
	if err != nil {
		return nil, errors.Wrapf(err, "include[%d]", index)
	}
	cfgMaps = append(cfgMaps, m)

	return cfgMaps, nil
}

func importPathInclude(ctx context.Context, index int, include Include, basePath string, cfgMaps []map[string]interface{}) ([]map[string]interface{}, error) {
	path := os.ExpandEnv(include.Path)
	m, cfgFile, err := loadMapFromPath(path, basePath)
	if err != nil {
		return nil, errors.Wrapf(err, "include[%d]", index)
	}

	// if m has its own includes need to load its includes too
	if _, ok := m["includes"]; ok {
		// nested includes
		pm, err := loadPreMissionMaps(ctx, m, cfgFile)
		if err != nil {
			return nil, errors.Wrapf(err, "include[%d]", index)
		}

		cfgMaps = append(cfgMaps, pm...)
	} else {
		cfgMaps = append(cfgMaps, m)
	}

	return cfgMaps, nil
}

func loadPreMissionMaps(ctx context.Context, spaceDust map[string]interface{}, configFile string) ([]map[string]interface{}, error) {
	cfgMaps, includes, err := decodeSpaceDust(spaceDust)
	if err != nil {
		return nil, err
	}

	// No includes, exit here
	if len(includes) == 0 {
		return cfgMaps, nil
	}

	// Get the base path of the config file so includes are relative to it
	basePath := filepath.Dir(configFile)

	for index, include := range includes {
		if include.Path != "" && include.URL != "" {
			return nil, fmt.Errorf("include[%d] has both url and path specified", index)
		}

		if include.Path != "" {
			cfgMaps, err = importPathInclude(ctx, index, include, basePath, cfgMaps)
			if err != nil {
				return nil, err
			}
		} else if include.URL != "" {
			cfgMaps, err = importURLInclude(ctx, index, include, cfgMaps)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("include[%d] has neither url nor path specified", index)
		}
	}

	return cfgMaps, nil
}

func missionMergeParams(mission, addition *Mission) {
	m := make(map[string]bool)
	for _, p := range mission.Params {
		m[p.Name] = true
	}

	params := make([]Param, 0, len(addition.Params))
	for _, p := range addition.Params {
		if _, ok := m[p.Name]; !ok || p.Name == "" {
			params = append(params, p)
		}
	}

	mission.Params = append(mission.Params, params...)
}

func missionMergeStages(mission, addition *Mission) {
	m := make(map[string]bool)
	for _, st := range mission.Stages {
		m[st.Name] = true
	}

	stages := make([]Stage, 0, len(addition.Stages))
	for _, st := range addition.Stages {
		if _, ok := m[st.Name]; !ok || st.Name == "" {
			stages = append(stages, st)
		}
	}

	mission.Stages = append(mission.Stages, stages...)
}

func missionMergeSequences(mission, addition *Mission) {
	if mission.Sequences == nil {
		mission.Sequences = make(map[string][]string)
	}

	for k, seq := range addition.Sequences {
		if _, ok := mission.Sequences[k]; !ok {
			mission.Sequences[k] = seq
		}
	}
}

func missionMergeEnv(mission, addition *Mission) {
	if len(addition.BasicEnv) > 0 {
		if mission.BasicEnv == nil {
			mission.BasicEnv = make(EnvMap)
		}
		for k, v := range addition.BasicEnv {
			if _, ok := mission.BasicEnv[k]; !ok {
				mission.BasicEnv[k] = v
			}
		}
	}
	if len(addition.Env) > 0 {
		if mission.Env == nil {
			mission.Env = make(EnvMap)
		}
		for k, v := range addition.Env {
			if _, ok := mission.Env[k]; !ok {
				mission.Env[k] = v
			}
		}
	}
}

func mergeMissions(mission, addition *Mission) {
	if mission.Name != "" {
		mission.Name = addition.Name
	}
	if mission.Version != "" {
		mission.Version = addition.Version
	}

	missionMergeEnv(mission, addition)

	if len(addition.Params) > 0 {
		missionMergeParams(mission, addition)
	}

	if len(addition.Stages) > 0 {
		missionMergeStages(mission, addition)
	}

	if len(addition.Sequences) > 0 {
		missionMergeSequences(mission, addition)
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

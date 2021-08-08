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

package rocket

import (
	"bytes"
	"context"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/resource"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func getStartingMissionURL(location string) (*url.URL, error) {
	if location == "" {
		location = "default"
	}

	return resource.UltimateURL(location)
}

func loadPreMission(ctx context.Context, spaceDust map[string]interface{}, missionURL *url.URL) (*Mission, error) {
	// loadPreMissionMaps returns a slice of maps
	// each map is then read in sequence and merged into the mission
	// includes do no override their parent's settings
	missionMaps, err := loadPreMissionMaps(ctx, spaceDust, missionURL)
	if err != nil {
		return nil, err
	}

	// With maps in place now build the mission
	return buildMissionFromMaps(missionMaps, missionURL)
}

func buildMissionFromMaps(missionMaps []map[string]interface{}, missionURL *url.URL) (*Mission, error) {
	// final mission
	mission := &Mission{}

	// iterate through config maps loading their missions, merging as we go
	for _, missionMap := range missionMaps {
		partialMission := &Mission{}

		// Load in the mission from the spaceDust
		if d, err := mapstructure.NewDecoder(
			&mapstructure.DecoderConfig{
				WeaklyTypedInput: true,
				Result:           partialMission,
			}); err != nil {
			return nil, errors.Wrap(err, "prepare")
		} else if err := d.Decode(missionMap); err != nil {
			return nil, errors.Wrap(loggee.BindMultiErrorFormatting(err), "decode")
		}

		mergeMissions(mission, partialMission)
	}

	// Setup the mission's name
	if mission.Name == "" {
		mission.Name = nameFromURL(missionURL)
	}

	return mission, nil
}

func nameFromURL(missionURL *url.URL) string {
	return strings.TrimSuffix(path.Base(missionURL.Path), path.Ext(missionURL.Path))
}

func loadPreMissionMaps(ctx context.Context, spaceDust map[string]interface{}, missionURL *url.URL) ([]map[string]interface{}, error) {
	// spaceDust contains a map of data, load this as a pre-mission to extract the includes
	missionMaps, includes, err := decodePreMissionSpaceDust(spaceDust, missionURL)
	if err != nil {
		return nil, err
	}

	// No includes, exit here
	if len(includes) == 0 {
		return missionMaps, nil
	}

	// Get the base path of the config file so includes are relative to it
	baseLocation := resource.GetURLParentLocation(missionURL).String()

	// Load the include maps
	for index, include := range includes {
		err := include.Validate()
		if err != nil {
			return nil, errors.Wrapf(err, "include[%d]", index)
		}

		// Convert the include struct to a URL
		url, err := convertIncludeToURL(baseLocation, include)
		if err != nil {
			return nil, errors.Wrapf(err, "include[%d]", index)
		}

		includeMap, err := loadMapFromURL(ctx, url)
		if err != nil {
			return nil, errors.Wrapf(err, "include[%d]", index)
		}

		// if m has its own includes need to load its includes too
		if _, ok := includeMap["includes"]; ok {
			// nested includes, recurse this function again
			pm, err := loadPreMissionMaps(ctx, includeMap, url)
			if err != nil {
				return nil, errors.Wrapf(err, "include[%d]", index)
			}

			missionMaps = append(missionMaps, pm...)
		} else {
			missionMaps = append(missionMaps, includeMap)
		}
	}

	return missionMaps, nil
}

func convertIncludeToURL(baseLocation string, include Include) (*url.URL, error) {
	if include.Path != "" && include.URL != "" {
		return nil, errors.New("both url and path specified")
	}
	var src string
	if include.URL != "" {
		src = include.URL
	} else {
		src = include.Path
	}

	// Expand environment variables
	src = os.ExpandEnv(src)

	return resource.UltimateURL(baseLocation, src)
}

func decodePreMissionSpaceDust(spaceDust map[string]interface{}, missionURL *url.URL) ([]map[string]interface{}, []Include, error) {
	preMission := &PreMission{}
	missionMaps := make([]map[string]interface{}, 0)

	// Load in the mission from the spaceDust
	if d, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			Result:           preMission,
		}); err != nil {
		return nil, nil, errors.Wrap(err, "setting up pre-mission decoder")
	} else if err := d.Decode(spaceDust); err != nil {
		return nil, nil, errors.Wrapf(loggee.BindMultiErrorFormatting(err), "parsing mission pre-mission %s ", missionURL)
	}

	missionMaps = append(missionMaps, preMission.Mission)

	return missionMaps, preMission.Includes, nil
}

func loadMapFromURL(ctx context.Context, url *url.URL) (map[string]interface{}, error) {
	b, err := resource.ReadURL(ctx, url)
	if err != nil {
		return nil, err
	}

	// Reads
	m := make(map[string]interface{})
	err = yaml.NewDecoder(bytes.NewBuffer(b)).Decode(&m)
	if err != nil {
		return nil, err
	}

	return m, err
}

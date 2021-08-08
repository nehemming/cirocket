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
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/resource"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Assemble locates a blueprint from the assembly sources, loads the runbook and builds the assembly following the blueprint.
func (mc *missionControl) Assemble(ctx context.Context, blueprintName string, sources []string, runbook string, params []Param) error {
	// blueprint, if not abs need to search sources to locate
	// once blueprint found extract it
	blueprint, blueprintLocation, err := mc.searchSources(ctx, blueprintName, sources, mc.missionLog())
	if err != nil {
		return err
	}

	// use specLocation to find any runbook
	var flightSequence []string
	if runbook != "" {
		rb, err := mc.loadRunbook(ctx, runbook)
		if err != nil {
			return errors.Wrapf(err, "loading runbook %s", resource.Relative(runbook))
		}
		// merge runbook and params, params take precedence
		params = mergeParamSets(params, rb.Params...)
		flightSequence = rb.FlightSequence
	}

	// load the mission
	spaceDust, location, err := loadMapFromLocation(ctx, blueprint.Mission, blueprintLocation)
	if err != nil {
		return errors.Wrapf(err, "loading mission for blueprint %s (%s)", blueprintName, resource.Relative(blueprintLocation))
	}

	// form a mission based on the blueprint mission, combned with runbook merged params
	// all following the runbook flight sequence
	// run the mission
	return mc.LaunchMissionWithParams(ctx, location, spaceDust, params, flightSequence...)
}

func loadMapFromLocation(ctx context.Context,
	location Location, parentLocation string) (map[string]interface{}, string, error) {
	// Check the location struct is valid
	err := location.Validate()
	if err != nil {
		return nil, "", err
	}

	// read location
	var missionMap map[string]interface{}
	if location.Inline != "" {
		// inline, mission text
		b := bytes.NewBufferString(location.Inline)
		err = yaml.NewDecoder(b).Decode(&missionMap)
		if err != nil {
			return nil, "", err
		}

		return missionMap, parentLocation, nil
	}

	if location.Path == "" {
		panic("location Validate failed")
	}

	//	Read the map
	u, err := resource.UltimateURL(parentLocation, location.Path)
	if err != nil {
		return nil, "", err
	}

	m, err := loadMapFromURL(ctx, u)
	if err != nil {
		return nil, "", err
	}

	return m, u.String(), nil
}

func buildSpecFromMap(specDust map[string]interface{}) (*Runbook, error) {
	runbook := new(Runbook)

	if d, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			Result:           runbook,
		}); err != nil {
		return nil, errors.Wrap(err, "runbook decoder")
	} else if err := d.Decode(specDust); err != nil {
		return nil, errors.Wrap(loggee.BindMultiErrorFormatting(err), "parsing runbook")
	}

	return runbook, nil
}

func (mc *missionControl) loadRunbook(ctx context.Context, runbookLocation string) (*Runbook, error) {
	url, err := resource.UltimateURL(runbookLocation)
	if err != nil {
		return nil, err
	}

	specDust, err := loadMapFromURL(ctx, url)
	if err != nil {
		return nil, err
	}

	return buildSpecFromMap(specDust)
}

func mergeParamSets(sourceParams []Param, additional ...Param) []Param {
	m := make(map[string]bool)
	for _, p := range sourceParams {
		m[p.Name] = true
	}

	params := make([]Param, 0, len(additional))
	for _, p := range additional {
		if _, ok := m[p.Name]; !ok || p.Name == "" {
			params = append(params, p)
		}
	}

	return append(sourceParams, params...)
}

const manifestFileName = "/blueprint.yml"

func (mc *missionControl) searchSources(ctx context.Context, blueprintName string,
	sources []string, log loggee.Logger) (*Blueprint, string, error) {
	blueprintName = filepath.ToSlash(blueprintName)

	if strings.HasSuffix(blueprintName, manifestFileName) {
		// name includes blueprint runbook file, strip as added later
		blueprintName = path.Dir(blueprintName)
	}

	if isBlueprintNameAbs(blueprintName) {
		// get the path to the names blueprint
		u, err := resource.GetParentLocation(blueprintName)
		if err != nil {
			return nil, "", err
		}
		blueprintName = path.Base(blueprintName)
		sources = make([]string, 1)
		sources[0] = u.String()
	}

	if blueprintName == "" {
		// Empty name, malformed
		return nil, "", fmt.Errorf("name %s is malformed", blueprintName)
	}

	return loadBlueprint(ctx, blueprintName, log, sources)
}

func loadBlueprint(ctx context.Context, blueprintName string,
	log loggee.Logger, sources []string) (*Blueprint, string, error) {
	path := path.Join(blueprintName, manifestFileName[1:])

	b, u, err := resource.Search(ctx, path, reportProgress(blueprintName, log), sources...)
	if err != nil {
		// relate back to blue print rather than path
		if nfe := resource.IsNotFoundError(err); nfe != nil {
			return nil, "", errors.Wrapf(nfe.Unwrap(), "blueprint %s", blueprintName)
		}

		return nil, "", err
	}

	// get back to the
	u, err = resource.GetParentLocation(u.String())
	if err != nil {
		return nil, "", err
	}
	location := u.String()
	blueprint, err := decodeBluerint(b)
	if err != nil {
		return nil, "", errors.Wrapf(err, "blueprint %s", blueprintName)
	}

	// finally init files
	if blueprint.Name == "" {
		blueprint.Name = blueprintName
	}
	blueprint.Location = location

	return blueprint, location, nil
}

func decodeBluerint(b []byte) (*Blueprint, error) {
	blueStuff := make(map[string]interface{})
	err := yaml.NewDecoder(bytes.NewBuffer(b)).Decode(&blueStuff)
	if err != nil {
		return nil, err
	}

	blueprint := new(Blueprint)

	if d, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			Result:           blueprint,
		}); err != nil {
		return nil, errors.Wrap(err, "blueprint decoder")
	} else if err := d.Decode(blueStuff); err != nil {
		return nil, errors.Wrapf(loggee.BindMultiErrorFormatting(err), "parsing blueprint")
	}

	// Remove trailing new lines from description
	blueprint.Description = strings.Trim(blueprint.Description, " \n")

	return blueprint, nil
}

func reportProgress(name string, log loggee.Logger) resource.Progress {
	return func(source string, u *url.URL, nfe *resource.NotFoundError) {
		if nfe != nil {
			log.WithField("source", source).
				WithField("blueprint", name).
				Debug("Tried")
		} else {
			log.WithField("source", source).
				WithField("blueprint", name).
				Debug("Found")
		}
	}
}

func isBlueprintNameAbs(name string) bool {
	// has to be a url or a filepath
	if strings.HasPrefix(name, "https://") ||
		strings.HasPrefix(name, "http://") ||
		strings.HasPrefix(name, "file:/") ||
		filepath.IsAbs(name) {
		return true
	}

	return false
}

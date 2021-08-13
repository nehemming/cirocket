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
	"context"

	"github.com/nehemming/cirocket/pkg/resource"
)

type (
	// Runbook is a list of parameters required to complete a blueprint assembly.
	Runbook struct {
		// Name of the runbook.
		Name string `mapstructure:"name"`

		// Blueprint is the name of the blue print
		Blueprint string `mapstructure:"blueprint"`

		// Params is a list of parameters needed for the runbook.
		Params []Param `mapstructure:"params"`

		FlightSequence []string `mapstructure:"sequence"`
	}
)

// GetRunbook gets the runbook for a blueprint.
func (mc *missionControl) GetRunbook(ctx context.Context, blueprintName string, sources []string) (string, error) {
	// blueprint, if not abs need to search sources to locate
	// once blueprint found extract it
	blueprint, blueprintLocation, err := mc.searchSources(ctx, blueprintName, sources)
	if err != nil {
		return "", err
	}

	if blueprint.Runbook == nil {
		return "", nil
	}

	// Get the runbook from the blueprint
	return getRunbookFromLocattion(ctx, *blueprint.Runbook, blueprintLocation)
}

func getRunbookFromLocattion(ctx context.Context, location Location, blueprintLocation string) (string, error) {
	// Check the location struct is valid
	err := location.Validate()
	if err != nil {
		return "", err
	}

	if location.Inline != "" {
		return location.Inline, nil
	}

	// Get from path
	b, err := resource.ReadResource(ctx, blueprintLocation, location.Path)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

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
	"sort"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/resource"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type (
	// TaskTypeInfo describes a task type.
	TaskTypeInfo struct {
		// Type of task.
		Type string `mapstructure:"name"`

		// Description is a free text description of the task
		Description string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description"`
	}

	// BlueprintInfo contains information about a blueprint.
	BlueprintInfo struct {
		// Name of the blueprint.
		Name string `mapstructure:"name"`

		// Version of the blueprint
		Version string `json:"version,omitempty" yaml:"version,omitempty" mapstructure:"version"`

		// Location is the location of the blueprint
		Location string `mapstructure:"location"`

		// Description is a free text description of the blueprint
		Description string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description"`
	}

	// Inventory is a list of blueprints.
	Inventory struct {
		Items []string `mapstructure:"items"`
	}
)

// ListTaskTypes list the types of task registered withmission control.
func (mc *missionControl) ListTaskTypes(ctx context.Context) (TaskTypeInfoList, error) {
	var list TaskTypeInfoList
	for _, tt := range mc.types {
		list = append(list, TaskTypeInfo{
			Type:        tt.Type(),
			Description: tt.Description(),
		})
	}

	sort.Sort(list)

	return list, nil
}

// ListBlueprints builds a list of all blueprints found in the passed sources.
func (mc *missionControl) ListBlueprints(ctx context.Context, sources []string) ([]BlueprintInfo, error) {
	var list []BlueprintInfo
	for _, source := range sources {
		err := ctx.Err()
		if err != nil {
			return list, err
		}

		list, err = listSource(ctx, source, list)
		if err != nil {
			return list, err
		}
	}

	return list, nil
}

func listSource(ctx context.Context, source string, list []BlueprintInfo) ([]BlueprintInfo, error) {
	b, err := resource.ReadResource(ctx, source, "inventory.yml")
	if err != nil {
		// ignore no inventories
		if resource.IsNotFoundError(err) != nil {
			return list, nil
		}

		return list, err
	}

	// get resource
	inventory, err := decodeInventory(b)
	if err != nil {
		return list, errors.Wrapf(err, "source %s", source)
	}

	// Get inventory items
	info, err := getBlueprintInfoFromInventory(ctx, source, inventory)
	if err != nil {
		return list, errors.Wrapf(err, "source %s", source)
	}

	return append(list, info...), nil
}

func decodeInventory(b []byte) (*Inventory, error) {
	blueStuff := make(map[string]interface{})
	err := yaml.NewDecoder(bytes.NewBuffer(b)).Decode(&blueStuff)
	if err != nil {
		return nil, err
	}

	inventory := new(Inventory)

	if d, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			Result:           inventory,
		}); err != nil {
		return nil, errors.Wrap(err, "inventory decoder")
	} else if err := d.Decode(blueStuff); err != nil {
		return nil, errors.Wrapf(loggee.BindMultiErrorFormatting(err), "parsing inventory")
	}

	return inventory, nil
}

func getBlueprintInfoFromInventory(ctx context.Context, source string, inventory *Inventory) ([]BlueprintInfo, error) {
	var info []BlueprintInfo

	// Get each blueprint and extract info
	for _, item := range inventory.Items {
		// get manifest
		b, err := resource.ReadResource(ctx, source, item, manifestFileName[1:])
		if err != nil {
			return nil, err
		}

		// get the blueprint
		blueprint, err := decodeBluerint(b)
		if err != nil {
			return nil, errors.Wrapf(err, "blueprint %s", item)
		}

		// get back to the location
		location, _ := resource.UltimateURL(source, item)

		if blueprint.Name == "" {
			blueprint.Name = item
		}

		info = append(info, BlueprintInfo{
			Name:        blueprint.Name,
			Location:    location.String(),
			Description: blueprint.Description,
		})
	}

	return info, nil
}

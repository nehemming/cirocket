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

import "github.com/pkg/errors"

type (
	// Blueprint is a ready to assemble mission.
	Blueprint struct {
		// Name of the blueprint.
		Name string `mapstructure:"name"`

		// Location is the location of the blueprint, set at read time
		Location string `mapstructure:"-"`

		// Description is a free text description of the blueprint
		Description string `mapstructure:"description"`

		// Location of runbook pro forma
		Runbook *Location `mapstructure:"runbook"`

		// Location of the mission
		Mission Location `mapstructure:"mission"`

		// Params is a list of parameters needed for the runbook.
		Params []Param `mapstructure:"params"`
	}

	// Location is a location of input data.
	Location struct {
		// Inline detail
		Inline string `mapstructure:"inline"`

		// Path provides the path to the item, relative to the blueprints location
		Path string `mapstructure:"path"`
	}
)

// Validate checks the object meets the validation requirements to
// have one and only one of the source properties defined.
func (l *Location) Validate() error {
	count := 0
	if l.Inline != "" {
		count++
	}
	if l.Path != "" {
		count++
	}

	if count > 1 {
		return errors.New("more than one source was specified, only one is permitted")
	}
	if count == 0 {
		return errors.New("no source was specified")
	}

	return nil
}

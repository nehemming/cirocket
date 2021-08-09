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

import "runtime"

// IsFiltered returns true if the filter should be applied to exclude the item.
func (filter *Filter) IsFiltered() bool { //nolint:cyclop
	if filter == nil {
		return false
	}

	if filter.Skip {
		return true
	}

	if len(filter.ExcludeArch) > 0 {
		for _, a := range filter.ExcludeArch {
			if a == runtime.GOARCH {
				return true
			}
		}
	}

	if len(filter.ExcludeOS) > 0 {
		for _, o := range filter.ExcludeOS {
			if o == runtime.GOOS {
				return true
			}
		}
	}

	if len(filter.IncludeArch) > 0 {
		included := false
		for _, a := range filter.IncludeArch {
			if a == runtime.GOARCH {
				included = true
				break
			}
		}

		if !included {
			return true
		}
	}

	if len(filter.IncludeOS) > 0 {
		included := false
		for _, o := range filter.IncludeOS {
			if o == runtime.GOOS {
				included = true
				break
			}
		}

		if !included {
			return true
		}
	}

	return false
}

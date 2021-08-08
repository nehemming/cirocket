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

package resource

import (
	"context"
	"errors"
	"net/url"
)

// Progress functions are called by Search with the result of each source search.
type Progress func(string, *url.URL, *NotFoundError)

// Search searches for a resource against a list of sources.
// If the resource is found tts contents are returned along with the url on which itt was located.
// If the resource could not be found an error is returned.
// The search will stop on any error other than a not found error.
// The relLocation and absSources are merged using UltimateURL.
func Search(ctx context.Context, relLocation string, progress Progress, absSources ...string) ([]byte, *url.URL, error) {
	for _, source := range absSources {
		// Cancelled
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}

		// Get rhe url
		url, err := UltimateURL(source, relLocation)
		if err != nil {
			return nil, nil, err
		}

		// Read the url
		b, err := ReadURL(ctx, url)
		if err != nil {
			// failed, try more?
			if nfe, ok := err.(*NotFoundError); ok {
				// try next source
				if progress != nil {
					progress(source, url, nfe)
				}
				continue
			}

			// Failure
			return nil, nil, err
		}

		// Found
		if progress != nil {
			progress(source, url, nil)
		}
		return b, url, nil
	}

	return nil, nil, NewNotFoundError(relLocation,
		errors.New("cannot find in sources"))
}

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

package builtin

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	globber "github.com/bmatcuk/doublestar/v4"
	"github.com/nehemming/cirocket/pkg/resource"
)

// recursiveGlob supports ** patterns in windows and mac/linux.
// In windows escaping is disabled.
func recursiveGlob(path string) ([]string, error) {
	// handle path in url format
	u, err := resource.UltimateURL(path)
	if err != nil {
		return nil, err
	}
	path, err = resource.URLToPath(u)
	if err != nil {
		return nil, err
	}

	// convert os paths to slash format
	dir, pattern := globber.SplitPattern(filepath.ToSlash(path))
	files, err := globber.Glob(slashingFS(dir), pattern)
	if err != nil {
		return nil, err
	}

	osDir := filepath.FromSlash(dir)
	for i, f := range files {
		osName := filepath.FromSlash(f)
		if filepath.IsAbs(osName) {
			files[i] = osName
		} else {
			files[i] = filepath.Join(osDir, osName)
		}
	}

	return files, nil
}

// slashingFS implements a file system supporting the glob library, which opperated in slash mode only
type slashingFS string

// Open opens a file expressed in slash format in the os specific manner.
// If an issue occurs an error is returned.
func (dir slashingFS) Open(name string) (fs.File, error) {
	osName := filepath.FromSlash(name)
	if filepath.IsAbs(osName) {
		return os.Open(osName)
	}

	osPath := filepath.Join(filepath.FromSlash(string(dir)), osName)

	return os.Open(osPath)
}

// friendlyRelativePath is a helper function to convert a path to be abs or relative based on the cwd and how it will appear
func friendlyRelativePath(path string, relTo ...string) string {
	var cwd string
	var err error

	// if no relTo assume cwd
	if len(relTo) > 0 {
		cwd, err = filepath.Abs(filepath.Join(relTo...))
		if err != nil {
			return path
		}
	} else {
		cwd, err = os.Getwd()
		if err != nil {
			return path
		}
	}

	wdBase := cwd
	if !strings.HasSuffix(cwd, string(filepath.Separator)) {
		wdBase += string(filepath.Separator)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	if !strings.HasPrefix(abs, wdBase) {
		// not rel.
		return path
	}

	rel, err := filepath.Rel(cwd, abs)
	if err != nil {
		return path
	}

	return rel
}

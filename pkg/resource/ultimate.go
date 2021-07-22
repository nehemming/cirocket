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
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
)

// Relative returns the best relative path to the current working directory
// If its file this will be a OS path, if its not it will be the the URL.
func Relative(any interface{}) string {
	switch url := any.(type) {
	case *url.URL:
		return relativeURL(url)
	case fmt.Stringer:
		return relativeString(url.String())
	case string:
		return relativeString(url)
	default:
		panic(fmt.Sprintf("type %T not supported", any))
	}
}

func relativeString(url string) string {
	u, e := UltimateURL(url)
	if e != nil {
		return url
	}
	return relativeURL(u)
}

func relativeURL(url *url.URL) string {
	if url.Scheme != "file" {
		return url.String()
	}
	p, err := URLToRelativePath(url)
	if err != nil {
		return url.String()
	}
	return p
}

// URLToPath returns a file system path too the resource or an error.
func URLToPath(url *url.URL) (string, error) {
	if url.Scheme != "file" {
		return "", fmt.Errorf("scheme %s is not supported", url.Scheme)
	}

	if url.Host != "" {
		return filepath.FromSlash("//" + url.Host + url.Path), nil
	}

	return filepath.FromSlash(url.Path), nil
}

// URLToRelativePath returns a file system path too the resource or an error.
func URLToRelativePath(url *url.URL, baseParts ...string) (string, error) {
	path, err := URLToPath(url)
	if err != nil {
		return "", err
	}

	var base string
	if len(baseParts) == 0 {
		base, err = os.Getwd()
		if err != nil {
			return "", err
		}
	} else {
		rel, err := UltimateURL(baseParts...)
		if err != nil {
			return "", err
		}
		base, err = URLToPath(rel)
		if err != nil {
			return "", err
		}
	}

	return filepath.Rel(base, path)
}

// GetURLParentLocation returns the parent location of a URL.
func GetURLParentLocation(url *url.URL) *url.URL {
	cpy := *url
	cpy.Path = path.Dir(path.Clean(cpy.Path))
	return &cpy
}

// GetParentLocation takes a filesystem or url path and returns a url representing the partent
// of the path.
// If there is no parent, the url will contain a path of ".".
func GetParentLocation(location string) (*url.URL, error) {
	parts := standardizeParts(location)
	if len(parts) == 0 {
		return nil, errors.New("no url provided")
	}
	location = parts[0]

	if isURL(location) {
		u, err := url.Parse(location)
		if err != nil {
			return nil, err
		}
		u.Path = path.Dir(path.Clean(u.Path))
		return u, nil
	}

	// Non url,
	return ultimateToFileURL(path.Dir(path.Clean(location)))
}

// UltimateURL returns a url or an error.
// The path parts are merged left to right to create a resource path which is then encoded as a UrL
// The parts may be either a file system path or a url in the file or http(s) schemes.
// pathParts are relative to the merger of the preceding parts.
// If no absolute path is provided the current working directory is used as the root directory.
//
// Examples of supported part formats
//
// windows: c:\\windows\\something or .\\windows\\file spaced.doc" or c:/something
//     unc: \\\\server\\share\\path
//    unix: /root/url/bin/somthing or ~/.home/thing or ../bin/go
//   https: https://server/resource
//    file: file:/somedata or file:///somedata or file:///c:/windows/clock.avi
//
// Support for both os windows and unix styles.
//
func UltimateURL(pathParts ...string) (*url.URL, error) {
	parts := standardizeParts(pathParts...)

	if len(parts) == 0 {
		return nil, errors.New("no url parts provided")
	}

	ultimate, err := mergeUltimate(parts)
	if err != nil {
		return nil, err
	}

	// ultimate is now concatenated
	// is it a url?
	if isURL(ultimate) {
		u, err := url.Parse(ultimate)
		if err != nil {
			return nil, err
		}
		u.Path = path.Clean(u.Path)
		return u, nil
	}
	return ultimateToFileURL(ultimate)
}

func ultimateToFileURL(ultimate string) (*url.URL, error) {
	if isUNC(ultimate) {
		// have ////server// ...
		server, rest := splitUnc(ultimate)
		ultimate = fmt.Sprintf("file://%s/%s", server, path.Clean(rest))

		return url.Parse(ultimate)
	}

	if isHome(ultimate) {
		// have ~/
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		ultimate = fmt.Sprintf("file://%s", expandHome(ultimate, home))

		return url.Parse(ultimate)
	}

	osPath, err := filepath.Abs(filepath.FromSlash(ultimate))
	if err != nil {
		return nil, err
	}

	ultimate = filepath.ToSlash(osPath)
	if isWindows(ultimate) {
		ultimate = "/" + ultimate
	}

	ultimate = fmt.Sprintf("file://%s", ultimate)

	return url.Parse(ultimate)
}

func mergeUltimate(parts []string) (string, error) {
	var ultimate string

	insideURL := false

	for _, part := range parts {
		// windows or "~" or '//unc/" paths always reset the any scheme
		if isWindows(part) || isUNC(part) || isHome(part) {
			ultimate = part
			insideURL = false
			continue
		}

		if isURL(part) {
			// reset scheme
			ultimate = part
			insideURL = true
			continue
		}

		if filepath.IsAbs(part) {
			if insideURL {
				// have abs path, but as extending an existing url it
				// replaces only the path part
				u, err := preserveURL(ultimate, part)
				if err != nil {
					return "", err
				}
				ultimate = u
			} else {
				// not ina url so this isa new abs file path
				ultimate = part
			}
			continue
		}

		// basic join - pat cannot start with a / or it is root
		ultimate = join(ultimate, part)
	}
	return ultimate, nil
}

func preserveURL(ultimate, rootPath string) (string, error) {
	u, err := url.Parse(ultimate)
	if err != nil {
		// bad url, return
		return "", err
	}

	// Kep the rest of the url intact
	u.Path = rootPath

	return u.String(), nil
}

func join(root, addition string) string {
	add := []rune(addition)
	if len(add) == 0 {
		return root
	}

	l := len(add) - 1
	if add[l] == '/' {
		addition = string(add[:l])
	}

	if root == "" {
		return addition
	}

	return root + "/" + addition
}

func expandHome(relPath, home string) string {
	if len(relPath) <= 1 {
		return home
	}

	return path.Join(home, relPath[1:])
}

func splitUnc(uncPath string) (string, string) {
	// guard
	if len(uncPath) < 3 {
		panic(fmt.Sprintf("unc bug, code should ensure longer than 3 chars, %s", uncPath))
	}

	parts := strings.SplitN(uncPath[2:], "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return parts[0], ""
}

func isHome(part string) bool {
	return part == "~" || strings.HasPrefix(part, "~/")
}

func isWindows(part string) bool {
	if len(part) > 2 && part[1] == ':' {
		return true
	}
	return false
}

func isURL(part string) bool {
	urlParts := strings.SplitN(part, ":/", 2)
	if len(urlParts) == 2 && len(urlParts[0]) > 1 {
		return true
	}
	return false
}

func isUNC(part string) bool {
	return strings.HasPrefix(part, "//")
}

func standardizeParts(rawParts ...string) []string {
	stdParts := make([]string, 0, len(rawParts))

	for _, raw := range rawParts {
		part := strings.Trim(filepath.ToSlash(raw), " ")

		if part == "" {
			continue
		}

		stdParts = append(stdParts, part)
	}

	return stdParts
}

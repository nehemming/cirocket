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
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/nehemming/cirocket/pkg/resource"
)

func initFuncMap() template.FuncMap {
	fm := template.FuncMap{

		"Indent":   indent,
		"indent":   indent,
		"username": getUserName,
		"now":      time.Now,
		"pwd":      os.Getwd,
		"dirname":  dirname,
		"basename": basedname,
		"ultimate": ultimate,
		"relative": resource.Relative,
		"home":     homedir.Dir,
	}

	return fm
}

func indent(indent int, text string) string {
	// indent indents all lines, except the first line by indent spaces
	// use in templates as a pipeline '|'
	lines := strings.Split(text, "\n")
	sb := strings.Builder{}
	spaces := strings.Repeat(" ", indent)
	for i, line := range lines {
		if i > 0 {
			sb.WriteString(spaces)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}

func dirname(p string) string {
	return path.Dir(filepath.ToSlash(p))
}

func basedname(p string) string {
	return path.Base(filepath.ToSlash(p))
}

func ultimate(p ...string) (string, error) {
	u, e := resource.UltimateURL(p...)
	if e != nil {
		return "", e
	}
	return u.String(), nil
}

func getUserName() string {
	if u, err := user.Current(); err == nil {
		if u.Name != "" {
			return u.Name
		}
		if u.Username != "" {
			return u.Username
		}
		if u.Uid != "" {
			return u.Uid
		}
	}

	return "unknown"
}

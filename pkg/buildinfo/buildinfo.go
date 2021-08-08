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

// Package buildinfo stores information about an applications current build.
package buildinfo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// Info is information about the build.
type Info struct {
	// BuiltBy who/what build this version.
	BuiltBy string
	// Date compiled.
	Date string
	// Commit used for build.
	Commit string
	// CompiledName compile name used for build.
	CompiledName string
	// RunName is the name of the program obtained from arg[0].
	RunName string
	// Version is the version of the program.
	Version string
	// Start dir is the working directory when the program starts.
	StartDir string
}

// ctxKeyType private context key type.
type ctxKeyType string

// ctxKey is the context key.
const ctxKey = ctxKeyType("info")

// GetBuildInfo returns the run context.
func GetBuildInfo(ctx context.Context) Info {
	v := ctx.Value(ctxKey)

	info, ok := v.(Info)
	if !ok {
		return Info{}
	}

	if info.RunName == "" {
		info.RunName = GetRunNameForProgram()
	}
	return info
}

// NewInfo creates a new info from the passed arguments.
// Use NewInfo to pass in the information captured from link flags.
//
// Eg --ldflags "-s -w -X main.version={{ .Version }} -X main.commit={{ .Commit }} -X main.date={{ .CommitDate }} -X main.builtBy={{ .Env.BUILTBY }}".
func NewInfo(version, commit, date, builtBy, compiledName string) Info {
	sd, _ := os.Getwd()

	return Info{
		CompiledName: compiledName,
		RunName:      GetRunNameForProgram(),
		Version:      version,
		Commit:       commit,
		Date:         date,
		BuiltBy:      builtBy,
		StartDir:     sd,
	}
}

// GetRunNameForProgram returns the base name of the running program.
func GetRunNameForProgram() string {
	return strings.ToLower(filepath.Base(os.Args[0]))
}

// String converts the build info to a string.
func (info Info) String() string {
	builtBy := info.BuiltBy

	if builtBy != "" {
		builtBy = "[" + builtBy + "]"
	}

	v := info.Version
	for _, p := range []string{info.Commit, info.Date, info.CompiledName, builtBy} {
		if p != "" {
			v = v + " " + p
		}
	}

	return v
}

// NewContext creates a new context containing the build information.
func (info Info) NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey, info)
}

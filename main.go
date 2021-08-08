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

// Application cirocket is a command line task runner tool aimed at support automated development activities.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nehemming/cirocket/internal/cmd"
	"github.com/nehemming/cirocket/pkg/buildinfo"
	_ "github.com/nehemming/cirocket/pkg/builtin"
)

var (
	// version is the built version of the software.
	version                         = "dev build"
	commit, date, builtBy, compName string
)

func main() {
	// Create main app context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Attach signal handler
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()

	// Save the captured build information to the context
	ctx = buildinfo.NewInfo(version, commit, date, builtBy, compName).NewContext(ctx)

	// Main service entrypoint
	exitCode := cmd.Run(ctx)

	// Exit with the returned exit code
	os.Exit(exitCode)
}

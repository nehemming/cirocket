package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nehemming/cirocket/internal/cmd"
	"github.com/nehemming/cirocket/pkg/buildinfo"
	_ "github.com/nehemming/cirocket/pkg/rocket/builtin"
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

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

// Package cmd provides the command line interface to cirocket.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/nehemming/cirocket/pkg/buildinfo"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/loggee/apexlog"
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// ExitCodeSuccess indicates a successful exit.
	ExitCodeSuccess = 0

	// ExitCodeError indicates a non successful process exit.
	ExitCodeError = 1

	configAssemblySources = "assembly.sources"

	configFileType = "yml"
)

type (
	cli struct {
		appName          string
		rootCmd          *cobra.Command
		missionFile      string
		configFile       string
		workingDir       string
		ctx              context.Context
		missionFileError error
		configError      error
		mission          *viper.Viper
		config           *viper.Viper
		debug            bool
		silent           bool
		logger           loggee.Logger
		homeDir          string
	}
)

// Run executes the command line interface to the app.  The passed ctx is used to cancel long running tasks.
// appName is the name of the application and forms the suffix of the dot config file.
func Run(ctx context.Context) int {
	// Setup cli clogging handler
	loggee.SetLogger(apexlog.NewCli(2))
	loggee.Default().SetLevel(loggee.InfoLevel)

	return runWithArgs(ctx, os.Args[1:], loggee.Default())
}

func stdMultiErrorLogger(log loggee.Logger) func([]error) string {
	return func(es []error) string {
		count := len(es)
		if count == 1 {
			return es[0].Error()
		}

		for _, err := range es {
			log.Warn(err.Error())
		}

		return fmt.Sprintf("%d errors occurred", count)
	}
}

func runWithArgs(ctx context.Context, args []string, logger loggee.Logger) int {
	loggee.SetMultiErrorFormatting(stdMultiErrorLogger(logger))

	// Create the cli and root command
	cli := newCli(ctx, logger)

	if err := cli.setHomeDir(); err != nil {
		loggee.Error(err.Error())
		return ExitCodeError
	}

	// Bind the args
	cli.rootCmd.SetArgs(args)

	// Add the commands
	cli.rootCmd.AddCommand(cli.newAssemblyCommand())
	cli.rootCmd.AddCommand(cli.newLaunchCommand())

	initCmd := cli.newInitCommand()

	initCmd.AddCommand(cli.newInitMissionCommand())
	initCmd.AddCommand(cli.newInitRunbookCommand())
	cli.rootCmd.AddCommand(initCmd)

	// Register the config hook until processed the flags will not have been read.
	cobra.OnInitialize(cli.loadMissionAndConfig)

	// Execute the root command
	if err := cli.rootCmd.Execute(); err != nil {
		loggee.Error(err.Error())
		return ExitCodeError
	}

	// Exit with success
	return ExitCodeSuccess
}

func newCli(ctx context.Context, logger loggee.Logger) *cli {
	info := buildinfo.GetBuildInfo(ctx)

	cli := &cli{
		appName: info.RunName,
		rootCmd: &cobra.Command{
			Use:           info.RunName,
			Short:         "launch the rocket powered task runner \U0001F680 ",
			Long:          "rocket powered task runner to support delivering ci build missions",
			Args:          cobra.NoArgs,
			SilenceErrors: true,
			Version:       info.String(),
		},
		ctx:     ctx,
		mission: viper.New(),
		config:  viper.New(),
		logger:  logger,
	}

	// Handle profile loading errors
	cli.rootCmd.PersistentPreRunE = cli.preRunCheckInitErrors

	cli.rootCmd.SetVersionTemplate(`{{printf "%s:%s\n" .Name .Version}}`)

	// express home in a familiar way
	home := "~"
	if runtime.GOOS == "windows" {
		home = "%HOME%"
	}

	cli.rootCmd.PersistentFlags().StringVar(&cli.configFile, flagConfig, "",
		filepath.FromSlash(
			fmt.Sprintf("specify a configuration file (default is %s)",
				cli.homeConfigName(home))))

	cli.rootCmd.PersistentFlags().StringVar(&cli.workingDir, flagWorkingDir, "",
		"specify a working directory (default is the starting directory)")

	cli.rootCmd.PersistentFlags().BoolVar(&cli.debug, flagDebug, false,
		"include debug info")

	cli.rootCmd.PersistentFlags().BoolVar(&cli.silent, flagSilent, false,
		"silence output (ignored if debug is specified too)")

	return cli
}

func (cli *cli) setHomeDir() error {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	cli.homeDir = home
	return nil
}

// preRunCheckInitErrors checks for config errors.
func (cli *cli) preRunCheckInitErrors(cmd *cobra.Command, args []string) error {
	if cli.configError != nil {
		return cli.configError
	}

	return rocket.Default().SetOptions(rocket.LoggerOption(cli.logger))
}

// loadMissionAndConfig is called during the cobra start up process to init the config/profile settings.
func (cli *cli) loadMissionAndConfig() {
	if cli.debug {
		cli.logger.SetLevel(loggee.DebugLevel)
	} else if cli.silent {
		cli.logger.SetLevel(loggee.ErrorLevel)
	}

	// Switch dir if necessary
	if cli.workingDir != "" {
		if err := os.Chdir(filepath.FromSlash(cli.workingDir)); err != nil {
			cli.configError = err
			return
		}
	}

	// check app has a name
	if cli.appName == "" {
		cli.configError = errors.New("No app name")
		return
	}

	// load config
	if err := cli.loadConfig(cli.homeDir); err != nil {
		cli.configError = err
		return
	}

	// load mission
	if err := cli.loadMission(); err != nil {
		cli.missionFileError = err
	}
}

func (cli *cli) defaultLocalConfigName() string {
	return fmt.Sprintf(".%scfg.yml", cli.appName)
}

func (cli *cli) configDir() string {
	return fmt.Sprintf(".%s", cli.appName)
}

func (cli *cli) homeConfigName(home string) string {
	return filepath.Join(home, cli.configDir(), "config.yml")
}

// loadConfig loads the users config.
func (cli *cli) loadConfig(home string) error {
	config := cli.config

	// Set the file type we expect
	config.SetConfigType(configFileType)

	config.SetEnvPrefix(strings.ToUpper(cli.appName))
	config.AutomaticEnv()

	if cli.configFile != "" {
		// Use profile file from the flag.
		config.SetConfigFile(cli.configFile)
		return config.ReadInConfig()
	}

	// Try in local file and if that is unsuccessful try in home folder
	cli.configFile = cli.defaultLocalConfigName()
	config.SetConfigName(cli.configFile)
	config.SetConfigFile(cli.configFile)
	err := config.ReadInConfig()
	if err == nil {
		return nil
	}

	// See if not found
	if _, ok := err.(*fs.PathError); !ok {
		// was not a file not found error
		return err
	}

	// using home folder
	cli.configFile = cli.homeConfigName(home)
	config.SetConfigFile(cli.configFile)
	err = config.ReadInConfig()
	if err == nil {
		return nil
	}

	// See if not found
	if _, ok := err.(*fs.PathError); !ok {
		// was not a file not found error
		return err
	}

	return cli.createDefaultConfig()
}

// createDefaultConfig creates a default user config.
func (cli *cli) createDefaultConfig() error {
	dir, name := filepath.Split(cli.configFile)

	// Set the assembly locations
	cli.config.Set(configAssemblySources, []string{
		"https://raw.githubusercontent.com/nehemming/cirocket-config/master/blueprints",
	})

	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	err = cli.config.SafeWriteConfigAs(cli.configFile)
	if err != nil {
		return err
	}

	// Save values down
	cli.config.SetConfigName(name)
	cli.config.SetConfigFile(cli.configFile)

	return nil
}

func (cli *cli) loadMission() error {
	mission := cli.mission

	// Establish logging
	isCustomConfig := false
	mission.SetConfigType("yml")

	if cli.missionFile != "" {
		// Use mission file from the flag.
		mission.SetConfigFile(cli.missionFile)
		isCustomConfig = true
	} else {
		// Search for mission in current dir ".(appName).yml".
		mission.AddConfigPath(".")
		mission.SetConfigName("." + cli.appName)
	}

	// If a config file is found, read it in.
	err := mission.ReadInConfig()
	cfgName := mission.ConfigFileUsed()

	// Save the config actually used
	if !isCustomConfig {
		cli.missionFile = cfgName
	}

	return err
}

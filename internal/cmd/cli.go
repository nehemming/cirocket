/*
Copyright © 2018-2021 Neil Hemming
*/

//Package cmd provides the command line interface to cirocket
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/nehemming/cirocket/pkg/buildinfo"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/loggee/apexlog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// ExitCodeSuccess indicates a successful exit
	ExitCodeSuccess = 0

	// ExitCodeError indicates a non successful process exit
	ExitCodeError = 1

	flagConfig     = "config"
	flagWorkingDir = "dir"
)

type (
	cli struct {
		appName    string
		rootCmd    *cobra.Command
		configFile string
		workingDir string
		ctx        context.Context
		initError  error
	}
)

// Run executes the command line interface to the app.  The passed ctx is used to cancel long running tasks.
// appName is the name of the application and forms the suffix of the dot config file
func Run(ctx context.Context) int {

	// Setup cli clogging handler
	loggee.SetLogger(apexlog.NewCli(2))

	info := buildinfo.GetBuildInfo(ctx)

	cli := &cli{
		appName: info.RunName,
		rootCmd: &cobra.Command{
			Use:           info.RunName,
			Short:         "\U0001F680 Rocket powered task runner",
			Long:          "Rocket powered task runner to assist delivering ci build missions",
			Args:          cobra.NoArgs,
			SilenceErrors: true,
			Version:       info.String(),
		},
		ctx: ctx,
	}

	cli.rootCmd.SetVersionTemplate(`{{printf "%s:%s\n" .Name .Version}}`)

	initCmd := &cobra.Command{
		Use:           "init",
		Short:         "Initialize a new configuration file",
		Long:          "Creates a new configuration file",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          cli.runInitCmd,
	}

	launchCmd := &cobra.Command{
		Use:           "launch [{flightSequence}]",
		Short:         "Launch the CI rocket \U0001F680",
		Long:          "Runs the CI config, if the config uses sequences, one or more can be specified as additional args",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          cli.runFireCmd,
	}

	cli.rootCmd.PersistentFlags().StringVar(&cli.configFile, flagConfig, "",
		fmt.Sprintf("specify a configuration file (default is ./%s)", cli.appName))

	cli.rootCmd.PersistentFlags().StringVar(&cli.workingDir, flagWorkingDir, "",
		"specify a working directory (default is the starting directory)")

	cli.rootCmd.AddCommand(initCmd)
	cli.rootCmd.AddCommand(launchCmd)
	cli.bindLaunchFlagsAndConfig(launchCmd)

	// Register the config hook, until svr.rootCmd.Execute() is in progress
	// the flags will not have been read.
	cobra.OnInitialize(cli.initConfig)

	// Execute the root command
	if err := cli.rootCmd.Execute(); err != nil {
		loggee.Error(err.Error())
		return ExitCodeError
	}

	// Exit with success
	return ExitCodeSuccess
}

// initConfig is called during the cobra start up process to init the config settings
func (cli *cli) initConfig() {

	// Switch dir if necessary
	if cli.workingDir != "" {
		if err := os.Chdir(cli.workingDir); err != nil {
			cli.initError = err
			return
		}
	}

	// Establish logging
	isCustomConfig := false
	viper.SetConfigType("yaml")

	if cli.configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cli.configFile)
		isCustomConfig = true
	} else {
		// Search config in current dir ".(appName).yml".
		viper.AddConfigPath(".")
		viper.SetConfigName("." + cli.appName + ".yml")
	}

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	cfgName := viper.ConfigFileUsed()

	//Save the config actually used
	if !isCustomConfig {
		cli.configFile = cfgName
	}

	// Error opening file, log issue
	if err != nil {
		cli.initError = err
	}
}

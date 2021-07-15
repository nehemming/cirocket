package cmd

import (
	"embed"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

//go:embed initcoonfig.yml
var initConfig embed.FS

func (cli *cli) runInitCmd(cmd *cobra.Command, args []string) error {

	if cli.initError == nil {
		// Opened an already existing file, error
		return fmt.Errorf("config file %s already exists", cli.configFile)
	}

	// Create a default config file
	f, err := initConfig.Open("initcoonfig.yml")
	if err != nil {
		// Developer build issue
		panic(err)
	}
	defer f.Close()

	if b, err := ioutil.ReadAll(f); err != nil {
		// Bad as incorrect config, developer issue
		panic(err)
	} else {
		if err := os.WriteFile(cli.configFile, b, 0666); err != nil {
			return errors.Wrap(err, " write config file")
		}
	}

	return nil
}

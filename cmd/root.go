/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/iansinnott/browser-gopher/pkg/logging"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// overwrite with:
// go build -ldflags "-X github.com/iansinnott/browser-gopher/cmd.Version=$(git describe --tags)"
var Version string = "v0.0.0-dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "browser-gopher",
	Short: "A tool aggregate your browsing history",
	Long: `browser-gopher will aggregate and backup your browsing history. Use the
populate command to populate all URLs from currently supported browsers.

Example:

	browser-gopher populate

`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		v, err := cmd.Flags().GetBool("version")
		if err != nil {
			fmt.Println(errors.Wrap(err, "failed to get version flag"))
			os.Exit(1)
		}

		if v {
			fmt.Println(Version)
		} else {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	debug := os.Getenv("DEBUG")
	if debug != "" && debug != "0" && debug != "false" {
		logging.SetLogLevel(logging.DEBUG)
		logging.Debug().Println("running in debug mode")
	}

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Display the version number")
}

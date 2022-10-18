/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/spf13/cobra"
)

// dbPathCmd represents the dbPath command
var dbPathCmd = &cobra.Command{
	Use:   "db-path",
	Short: "Print the path to the database",
	Long: `
Print the path to the database. Useful if you want to use SQL on your database
directly.

Example:
	# Print the path
	browser-gopher db-path
	
	# Use the path to connect via sqlite3
	sqlite3 $(browser-gopher db-path) 'SELECT * FROM urls LIMIT 3;'

	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.Config.DBPath)
	},
}

func init() {
	rootCmd.AddCommand(dbPathCmd)
}

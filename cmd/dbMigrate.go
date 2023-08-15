/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/spf13/cobra"
)

// dbMigrateCmd represents the dbMigrate command
var dbMigrateCmd = &cobra.Command{
	Use:   "db-migrate",
	Short: "Migrate the database and do nothing else.",
	Long:  `Migrate the database and do nothing else. This is useful in development.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := persistence.InitDb(cmd.Context(), config.Config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer db.Close()
		fmt.Println("Database migrated successfully.")
	},
}

func init() {
	rootCmd.AddCommand(dbMigrateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dbMigrateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dbMigrateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

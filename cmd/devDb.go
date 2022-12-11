package cmd

import (
	"fmt"
	"os"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database related commands",
}

var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	Run: func(cmd *cobra.Command, args []string) {
		// init the db
		_, err := persistence.InitDb(cmd.Context(), config.Config)
		if err != nil {
			fmt.Println("error initializing db", err)
			os.Exit(1)
		}

		fmt.Println("db initialized: " + config.Config.DBPath)
	},
}

func init() {
	dbCmd.AddCommand(dbInitCmd)
	devCmd.AddCommand(dbCmd)
}

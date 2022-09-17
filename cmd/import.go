package cmd

import (
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import from various data sources. ",
	Long:  `Importing is used for pulling data from non-browser sources into the URL database.`,
}

func init() {
	rootCmd.AddCommand(importCmd)
}

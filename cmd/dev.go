package cmd

import (
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Dev tools",
}

func init() {
	rootCmd.AddCommand(devCmd)
}

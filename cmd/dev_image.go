package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/trashhalo/imgcat"
)

var devimageCmd = &cobra.Command{
	Use:   "image",
	Short: "Render a test image",
	Run: func(cmd *cobra.Command, args []string) {
		url, err := cmd.Flags().GetString("url")
		if err != nil {
			fmt.Println("could not parse --url:", err)
			os.Exit(1)
		}
		if url == "" {
			fmt.Println("--url is required")
			os.Exit(1)
		}

		w, err := cmd.Flags().GetInt("width")
		if err != nil {
			fmt.Println("could not parse --width:", err)
			os.Exit(1)
		}
		h, err := cmd.Flags().GetInt("height")
		if err != nil {
			fmt.Println("could not parse --height:", err)
			os.Exit(1)
		}

		fmt.Println("todo: implement w, h", w, h)

		m := imgcat.NewModel([]string{url})
		p := tea.NewProgram(m)
		if err := p.Start(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}

func init() {
	devCmd.AddCommand(devimageCmd)
	devimageCmd.Flags().String("url", "https://avatars.githubusercontent.com/u/3154865?s=40&v=4", "The image to load")
	devimageCmd.Flags().Int("width", 128, "width")
	devimageCmd.Flags().Int("height", 128, "height")
}

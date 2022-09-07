/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/iansinnott/browser-gopher/pkg/safari"
	"github.com/iansinnott/browser-gopher/pkg/types"
)

// @note Not using cobra for now, but the project was initialized with it for
// later use once managing cli commands becomes tedious.

// import "github.com/iansinnott/browser-gopher/cmd"
// func main() {
// 	cmd.Execute()
// }

func PopulateAll(extractor types.Extractor) error {
	urls, err := extractor.GetAllUrls()
	if err != nil {
		return err
	}

	visits, err := extractor.GetAllVisits()

	log.Println("[safari] found urls", len(urls))
	log.Println("[safari] found visits", len(visits))

	return nil
}

func expanduser(path string) string {
	userHome, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("could not get user home", err)
		os.Exit(1)
	}

	return strings.Replace(path, "~", userHome, 1)
}

func main() {
	var err error

	browserName := flag.String("browser", "", "Specify which browser")

	flag.Parse()

	if browserName == nil || len(*browserName) == 0 {
		fmt.Println("No browser name specified. Use the -browser flag")
		os.Exit(1)
	}

	switch *browserName {
	case "safari":
		safari := safari.SafariExtractor{
			Name:          "safari",
			HistoryDBPath: expanduser("~/Library/Safari/History.db"),
		}
		err = PopulateAll(&safari)
	default:
		fmt.Printf(`Browser not supported "%s"\n`, *browserName)
		os.Exit(1)
	}

	if err != nil {
		fmt.Println("Encountered an error", err)
		os.Exit(1)
	}
}

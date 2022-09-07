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

	"github.com/iansinnott/browser-gopher/pkg/extractors"
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
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

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

	extractors := []types.Extractor{
		&extractors.SafariExtractor{
			Name:          "safari",
			HistoryDBPath: expanduser("~/Library/Safari/History.db"),
		},
	}

	switch *browserName {
	case "safari":
		var extractor types.Extractor

		for _, x := range extractors {
			if x.GetName() == *browserName {
				extractor = x
			}
		}

		if extractor == nil {
			fmt.Println("Could not find extractor for", *browserName)
			os.Exit(1)
		}

		err = PopulateAll(extractor)
	case "":
		errs := []error{}
		// Given the empty string populate all browsers
		for _, x := range extractors {
			e := PopulateAll(x)
			if e != nil {
				errs = append(errs, e)
			}
		}

		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Println(e)
			}
			err = fmt.Errorf("one or more browsers failed")
		}
	default:
		fmt.Printf(`Browser not supported "%s"\n`, *browserName)
		os.Exit(1)
	}

	if err != nil {
		fmt.Println("Encountered an error", err)
		os.Exit(1)
	}
}

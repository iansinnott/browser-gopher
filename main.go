/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"

	ex "github.com/iansinnott/browser-gopher/pkg/extractors"
	"github.com/iansinnott/browser-gopher/pkg/types"
)

// @note Not using cobra for now, but the project was initialized with it for
// later use once managing cli commands becomes tedious.

// import "github.com/iansinnott/browser-gopher/cmd"
// func main() {
// 	cmd.Execute()
// }

func PopulateAll(extractor types.Extractor) error {
	log.Println("["+extractor.GetName()+"] reading", extractor.GetDBPath())
	conn, err := sql.Open("sqlite", extractor.GetDBPath())
	ctx := context.TODO()

	if err != nil {
		fmt.Println("could not connect to db at", extractor.GetDBPath(), err)
		return err
	}
	defer conn.Close()

	urls, err := extractor.GetAllUrls(ctx, conn)
	if err != nil {
		return err
	}

	visits, err := extractor.GetAllVisits(ctx, conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	log.Println("["+extractor.GetName()+"]\tfound urls", len(urls))
	log.Println("["+extractor.GetName()+"]\tfound visits", len(visits))

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
		&ex.SafariExtractor{
			Name:          "safari",
			HistoryDBPath: expanduser("~/Library/Safari/History.db"),
		},
	}

	ffdbs, err := ex.FindFirefoxDBs(expanduser("~/Library/Application Support/Firefox/Profiles/"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, ffdb := range ffdbs {
		extractors = append(extractors, &ex.ChromiumExtractor{
			Name:          "firefox",
			HistoryDBPath: ffdb,
		})
	}

	chdbs, err := ex.FindChromiumDBs(expanduser("~/Library/Application Support/Google/Chrome/"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, chdb := range chdbs {
		extractors = append(extractors, &ex.ChromiumExtractor{
			Name:          "chrome",
			HistoryDBPath: chdb,
		})
	}

	switch *browserName {
	case "safari", "firefox", "chrome":
		for _, x := range extractors {
			if x.GetName() == *browserName {
				err = PopulateAll(x)
				if err != nil {
					log.Printf("Error with extractor: %+v\n", x)
				}
			}
		}
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

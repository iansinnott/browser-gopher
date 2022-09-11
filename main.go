/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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

func copyPath(frm, to string) error {
	dest, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer dest.Close()

	src, err := os.Open(frm)
	if err != nil {
		return err
	}
	defer src.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		return err
	}

	return nil
}

func PopulateAll(extractor types.Extractor) error {
	log.Println("["+extractor.GetName()+"] reading", extractor.GetDBPath())
	conn, err := sql.Open("sqlite", extractor.GetDBPath())
	ctx := context.TODO()

	if err != nil {
		log.Println("could not connect to db at", extractor.GetDBPath(), err)
		return err
	}
	defer conn.Close()

	// Handle the case where the database is in use, or return if the database cannot be read or copied.
	_, err = extractor.VerifyConnection(ctx, conn)
	if err != nil {
		if !strings.Contains(err.Error(), "SQLITE_BUSY") {
			log.Println("[err] Could read from DB", extractor.GetDBPath())
			return err
		}

		log.Println("[" + extractor.GetName() + "] database is locked. copying for read access.")

		tmpPath := filepath.Join(os.TempDir(), extractor.GetName()+"_backup.sqlite")

		err := copyPath(extractor.GetDBPath(), tmpPath)
		if err != nil {
			fmt.Println("could not copy:", tmpPath)
			return err
		}

		// Update extractor to use the tmp path
		extractor.SetDBPath(tmpPath)

		defer func() {
			if os.Remove(tmpPath) != nil {
				log.Println("could not remove tmp file:", tmpPath)
			}
		}()

		return PopulateAll(extractor)
	}

	urls, err := extractor.GetAllUrls(ctx, conn)
	if err != nil {
		return err
	}

	visits, err := extractor.GetAllVisits(ctx, conn)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("["+extractor.GetName()+"]\tfound urls", len(urls))
	log.Println("["+extractor.GetName()+"]\tfound visits", len(visits))

	return nil
}

func main() {
	var err error

	browserName := flag.String("browser", "", "Specify which browser")

	flag.Parse()

	extractors, err := ex.BuildExtractorList()
	if err != nil {
		log.Println("error getting extractors", err)
		os.Exit(1)
	}

	if *browserName != "" {
		for _, x := range extractors {
			if x.GetName() == *browserName {
				err = PopulateAll(x)
				if err != nil {
					log.Printf("Error with extractor: %+v\n", x)
				}
			}
		}
	} else {
		errs := []error{}

		// Without a browser name, populate everything
		for _, x := range extractors {
			e := PopulateAll(x)
			if e != nil {
				errs = append(errs, e)
			}
		}

		if len(errs) > 0 {
			for _, e := range errs {
				log.Println(e)
			}
			err = fmt.Errorf("one or more browsers failed")
		}
	}

	if err != nil {
		fmt.Println("Encountered an error", err)
		os.Exit(1)
	}
}

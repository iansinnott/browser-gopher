package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/iansinnott/browser-gopher/pkg/util"
)

type AppConfig struct {
	AppDataPath     string
	BackupDir       string
	DBPath          string
	SearchIndexPath string
}

// initialize the config object and perform setup tasks.
func newConfig() *AppConfig {
	conf := &AppConfig{
		AppDataPath: util.Expanduser(filepath.Join("~", ".config", "browser-gopher")),
		BackupDir:   util.Expanduser(filepath.Join("~", ".cache", "browser-gopher")),
	}

	err := os.MkdirAll(conf.AppDataPath, 0755)
	if err != nil {
		log.Fatal("could not create app data path: "+conf.AppDataPath, err)
	}

	err = os.MkdirAll(conf.BackupDir, 0755)
	if err != nil {
		log.Fatal("could not create app data path: "+conf.AppDataPath, err)
	}

	conf.DBPath = filepath.Join(conf.AppDataPath, "db.sqlite")
	conf.SearchIndexPath = filepath.Join(conf.AppDataPath, "searchindex.bleve")

	return conf
}

var Config *AppConfig = newConfig()

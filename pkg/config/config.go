package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/iansinnott/browser-gopher/pkg/util"
)

type config struct {
	AppDataPath string
}

// initialize the config object and perform setup tasks.
func newConfig() *config {
	conf := &config{
		AppDataPath: util.Expanduser(filepath.Join("~", ".config", "browser-gopher")),
	}

	err := os.MkdirAll(conf.AppDataPath, 0755)
	if err != nil {
		log.Fatal("could not create app data path: "+conf.AppDataPath, err)
	}

	return conf
}

var Config *config = newConfig()

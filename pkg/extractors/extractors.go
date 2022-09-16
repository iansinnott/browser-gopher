package extractors

import (
	"log"
	"os"

	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
)

type pathSpec struct {
	name            string
	path            string
	findDBs         func(string) ([]string, error)
	createExtractor func(name string, dbPath string) types.Extractor
}

// Build a list of relevant extractors for this system
// @todo If we want to go multi platform this is currently the place to specify
// the logic to determine paths on a per-platform basis. The extractors should
// all Just Work if they are pointed to an appropriate sqlite db.
func BuildExtractorList() ([]types.Extractor, error) {
	result := []types.Extractor{}

	pathsToTry := []pathSpec{
		// Chrome-like
		{
			name:    "chrome",
			path:    util.Expanduser("~/Library/Application Support/Google/Chrome/"),
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},
		{
			name:    "brave",
			path:    util.Expanduser("~/Library/Application Support/BraveSoftware/Brave-Browser"),
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},
		{
			name:    "brave-beta",
			path:    util.Expanduser("~/Library/Application Support/BraveSoftware/Brave-Browser-Beta"),
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},
		{
			name:    "arc",
			path:    util.Expanduser("~/Library/Application Support/Arc/User Data"),
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},
		{
			name:    "vivaldi",
			path:    util.Expanduser("~/Library/Application Support/Vivaldi"),
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},
		{
			name:    "sidekick",
			path:    util.Expanduser("~/Library/Application Support/Sidekick"),
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},
		{
			name:    "edge",
			path:    util.Expanduser("~/Library/Application Support/Microsoft Edge"),
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},

		// Firefox-like
		// @todo What is the path for FF dev edition?
		{
			name:    "firefox",
			path:    util.Expanduser("~/Library/Application Support/Firefox/Profiles/"),
			findDBs: FindFirefoxDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &FirefoxExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},

		// Firefox-like
		// @todo What is the path for safari preview edition?
		{
			name: "safari",
			path: util.Expanduser("~/Library/Safari/"),
			findDBs: func(s string) ([]string, error) {
				dbPath := s + "History.db"
				if _, err := os.Stat(dbPath); err != nil {
					return nil, err
				}
				return []string{dbPath}, nil
			},
			createExtractor: func(name, dbPath string) types.Extractor {
				return &SafariExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},

		// Orion
		{
			name: "orion",
			path: util.Expanduser("~/Library/Application Support/Orion/Defaults/"),
			findDBs: func(s string) ([]string, error) {
				dbPath := s + "history"
				if _, err := os.Stat(dbPath); err != nil {
					return nil, err
				}
				return []string{dbPath}, nil
			},
			createExtractor: func(name, dbPath string) types.Extractor {
				return &OrionExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},

		// @note Of all the browsers listed above sigmaos seems to be the most
		// actively changing with the most novel data model. So this may well break
		// with some future update.
		{
			name: "sigmaos",
			path: util.Expanduser("~/Library/Containers/com.sigmaos.sigmaos.macos/Data/Library/Application Support/SigmaOS/"),
			findDBs: func(s string) ([]string, error) {
				dbPath := s + "Model.sqlite"
				if _, err := os.Stat(dbPath); err != nil {
					return nil, err
				}
				return []string{dbPath}, nil
			},
			createExtractor: func(name, dbPath string) types.Extractor {
				return &SigmaOSExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},
	}

	for _, x := range pathsToTry {
		stat, err := os.Stat(x.path)
		if err != nil || !stat.IsDir() {
			// @todo Put this into a debug logger to avoid noise
			log.Println("["+x.name+"] not found. skipping:", x.path)
			continue
		}

		dbs, err := x.findDBs(x.path)
		if err != nil {
			return nil, err
		}
		for _, dbPath := range dbs {
			result = append(result, x.createExtractor(x.name, dbPath))
		}
	}

	return result, nil
}

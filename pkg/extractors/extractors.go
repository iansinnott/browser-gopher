package extractors

import (
	"errors"
	"os"

	"github.com/iansinnott/browser-gopher/pkg/logging"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
)

type browserDataSource struct {
	name string
	// @todo Hrm, so make this []string as a way to support multiple OSs? is there a case where any of the other logic would not be platform agnostic?
	paths           []string
	findDBs         func(string) ([]string, error)
	createExtractor func(name string, dbPath string) types.Extractor
}

// Build a list of relevant extractors for this system
// @todo If we want to go multi platform this is currently the place to specify
// the logic to determine paths on a per-platform basis. The extractors should
// all Just Work if they are pointed to an appropriate sqlite db.
func BuildExtractorList() ([]types.Extractor, error) {
	result := []types.Extractor{}

	candidateBrowsers := []browserDataSource{
		// Chrome-like
		{
			name:    "chrome",
			paths:   []string{util.Expanduser("~/Library/Application Support/Google/Chrome/")},
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},
		{
			name:    "brave",
			paths:   []string{util.Expanduser("~/Library/Application Support/BraveSoftware/Brave-Browser")},
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},
		{
			name:    "brave-beta",
			paths:   []string{util.Expanduser("~/Library/Application Support/BraveSoftware/Brave-Browser-Beta")},
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},
		{
			name:    "arc",
			paths:   []string{util.Expanduser("~/Library/Application Support/Arc/User Data")},
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},
		{
			name:    "vivaldi",
			paths:   []string{util.Expanduser("~/Library/Application Support/Vivaldi")},
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},
		{
			name:    "sidekick",
			paths:   []string{util.Expanduser("~/Library/Application Support/Sidekick")},
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},
		{
			name:    "edge",
			paths:   []string{util.Expanduser("~/Library/Application Support/Microsoft Edge")},
			findDBs: FindChromiumDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &ChromiumExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},

		// Firefox-like
		// @todo What is the path for FF dev edition?
		{
			name: "firefox",
			paths: []string{
				util.Expanduser("~/Library/Application Support/Firefox/Profiles/"), // osx
				util.Expanduser("~/.mozilla/firefox/"),                             // lin
			},
			findDBs: FindFirefoxDBs,
			createExtractor: func(name, dbPath string) types.Extractor {
				return &FirefoxExtractor{Name: name, HistoryDBPath: dbPath}
			},
		},

		// Firefox-like
		// @todo What is the path for safari preview edition?
		{
			name:  "safari",
			paths: []string{util.Expanduser("~/Library/Safari/")},
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
			name:  "orion",
			paths: []string{util.Expanduser("~/Library/Application Support/Orion/Defaults/")},
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
			name:  "sigmaos",
			paths: []string{util.Expanduser("~/Library/Containers/com.sigmaos.sigmaos.macos/Data/Library/Application Support/SigmaOS/")},
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

	for _, browser := range candidateBrowsers {
		for _, p := range browser.paths {
			_, err := os.Stat(p)
			if errors.Is(err, os.ErrNotExist) {
				// @todo Put this into a debug logger to avoid noise
				logging.Debug().Println("["+browser.name+"] not found. skipping:", browser.paths)
				continue
			}

			dbs, err := browser.findDBs(p)
			if err != nil {
				return nil, err
			}
			for _, dbPath := range dbs {
				result = append(result, browser.createExtractor(browser.name, dbPath))
			}
		}
	}

	return result, nil
}

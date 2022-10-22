package search

import (
	"strings"

	bs "github.com/blevesearch/bleve/v2/search"
)

type HighlightMatchFn func(match string) string

func HighlightAll(locations bs.TermLocationMap, text string, render HighlightMatchFn) string {
	var lastEnd uint64 = 0
	xs := []string{}

	for _, locs := range locations {
		for _, loc := range locs {
			if lastEnd > loc.Start {
				// locations out of order? if you slice something like [12:3] it will fail. first must be less than second
				continue
			}

			xs = append(xs, text[lastEnd:loc.Start])
			xs = append(xs, render(text[loc.Start:loc.End]))
			lastEnd = loc.End
		}
	}

	xs = append(xs, text[lastEnd:])

	return strings.Join(xs, "")
}

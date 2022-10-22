package search_test

import (
	"testing"

	bs "github.com/blevesearch/bleve/v2/search"
	"github.com/iansinnott/browser-gopher/pkg/search"
)

func testHighlighter(s string) string {
	return `<match>` + s + `</match>`
}

func TestHighlight(t *testing.T) {
	table := []struct {
		name     string
		a        bs.TermLocationMap
		b        string
		expected string
	}{
		{
			name: "individual term at start of string, sub-word match",
			a: bs.TermLocationMap{
				"test": bs.Locations{
					{
						Start: 0,
						End:   4,
					},
				},
			},
			b:        "testing how this works",
			expected: `<match>test</match>ing how this works`,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			actual := search.HighlightAll(tt.a, tt.b, testHighlighter)
			if actual != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, actual)
			}
		})
	}
}

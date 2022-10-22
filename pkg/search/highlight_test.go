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
					{Start: 0, End: 4},
				},
			},
			b:        "testing how this works",
			expected: `<match>test</match>ing how this works`,
		},
		{
			name: "multiple matches",
			a: bs.TermLocationMap{
				"heyo": bs.Locations{
					{Start: 0, End: 4},
					{Start: 5, End: 9},
				},
			},
			b:        "heyo heyo",
			expected: `<match>heyo</match> <match>heyo</match>`,
		},
		{
			a: bs.TermLocationMap{
				"hey": bs.Locations{
					{Start: 0, End: 3},
				},
				"you": bs.Locations{
					{Start: 4, End: 7},
				},
			},
			b:        "hey you",
			expected: `<match>hey</match> <match>you</match>`,
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

package util_test

import (
	"testing"

	"github.com/iansinnott/browser-gopher/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestReverseSlice(t *testing.T) {

	table := []struct {
		name     string
		a        []string
		expected []string
	}{
		{
			name:     "empty slice",
			a:        []string{},
			expected: []string{},
		},
		{"single item slice", []string{"a"}, []string{"a"}},
		{"two item slice", []string{"a", "b"}, []string{"b", "a"}},
		{"three item slice", []string{"a", "b", "c"}, []string{"c", "b", "a"}},
		{"three item slice", []string{"a", "bb", "c"}, []string{"c", "bb", "a"}},
		{"multi-slice", []string{"abc", "bb", "heyo"}, []string{"heyo", "bb", "abc"}},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			actual := util.ReverseSlice(tt.a)
			require.ElementsMatch(t, tt.expected, actual)
		})
	}

	intsTable := []struct {
		name     string
		a        []int
		expected []int
	}{
		{"empty slice", []int{}, []int{}},
		{"single item slice", []int{12, 3, 58, 2}, []int{2, 58, 3, 12}},
	}

	for _, tt := range intsTable {
		t.Run(tt.name, func(t *testing.T) {
			actual := util.ReverseSlice(tt.a)
			require.ElementsMatch(t, tt.expected, actual)
		})
	}
}

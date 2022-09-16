package util_test

import (
	"testing"

	"github.com/iansinnott/browser-gopher/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestReverseSlice(t *testing.T) {
	reversed := util.ReverseSlice([]string{"a", "b", "c"})
	expected := []string{"c", "b", "a"}
	require.ElementsMatch(t, reversed, expected)
}

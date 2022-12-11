package persistence_test

import (
	"context"
	"testing"

	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/iansinnott/browser-gopher/pkg/persistence/testutils"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestInitDb(t *testing.T) {
	dbConn, err := testutils.GetTestDBConn(t)
	require.NoError(t, err)
	defer dbConn.Close()
}

func TestInsertUrlMetadata(t *testing.T) {
	ctx := context.Background()
	dbConn, err := testutils.GetTestDBConn(t)
	require.NoError(t, err)
	defer dbConn.Close()

	table := []struct {
		name     string
		metas    []types.UrlMetaRow
		expected []string
	}{
		{
			name: "single item slice",
			metas: []types.UrlMetaRow{
				{Url: "http://www.google.com"},
			},
			expected: []string{util.HashMd5String("http://www.google.com")},
		},
		{
			name: "single item slice",
			metas: []types.UrlMetaRow{
				{Url: "http://abc"},
				{Url: "http://123"},
			},
			expected: []string{
				util.HashMd5String("http://abc"),
				util.HashMd5String("http://123"),
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			err = persistence.InsertUrlMeta(ctx, dbConn, tt.metas...)
			require.NoError(t, err)

			for i, expected := range tt.expected {
				var result string
				hash := util.HashMd5String(tt.metas[i].Url)
				dbConn.QueryRow("SELECT url_md5 FROM urls_meta where url_md5 = ?", hash).Scan(&result)
				require.Equal(t, expected, result)
			}
		})
	}

}

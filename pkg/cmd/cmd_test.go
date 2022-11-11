package cmd_test

import (
	"context"
	"database/sql"
	"net/url"
	"testing"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/save"
)

func testdb(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	db, err := save.Open(":memory:", url.Values{"cache": []string{"shared"}})
	assert.Nil(t, err)
	t.Cleanup(func() { db.Close() })
	_, err = db.ExecContext(ctx, save.Schema)
	assert.Nil(t, err)
	return db
}

package save_test

import (
	"context"
	"database/sql"
	"net/url"
	"testing"
	"time"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/save"
)

func TestAddPerson(t *testing.T) {
	ctx := context.Background()
	q := save.New(testdb(t, ctx))
	assert.Nil(t, q.AddPerson(ctx, "alice"))
	assert.Error(t, "UNIQUE constraint failed", q.AddPerson(ctx, "alice"))
}

func TestAddShift(t *testing.T) {
	ctx := context.Background()
	q := save.New(testdb(t, ctx))
	ps := []save.Person{{Handle: "alice"}, {Handle: "bob"}}
	for _, p := range ps {
		assert.Nil(t, q.AddPerson(ctx, p.Handle))
	}
	t0 := time.Date(2022, 10, 31, 13, 25, 42, 12345, time.UTC)
	t1 := time.Date(2022, 11, 3, 13, 25, 42, 12345, time.UTC)

	assert.Nil(t, q.AddShift(ctx, save.AddShiftParams{
		Person:    ps[0].Handle,
		StartAt:   t0,
		EndBefore: t1,
	}))

	// No uniqueness constraint
	assert.Nil(t, q.AddShift(ctx, save.AddShiftParams{
		Person:    ps[0].Handle,
		StartAt:   t0,
		EndBefore: t1,
	}))

	// Backwards doesn't work
	assert.Error(t, "CHECK constraint failed", q.AddShift(ctx, save.AddShiftParams{
		Person:    ps[0].Handle,
		StartAt:   t1,
		EndBefore: t0,
	}))
}

func testdb(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	db, err := save.Open(":memory:", url.Values{"cache": []string{"shared"}})
	assert.Nil(t, err)
	t.Cleanup(func() { db.Close() })
	_, err = db.ExecContext(ctx, save.Schema)
	assert.Nil(t, err)
	return db
}

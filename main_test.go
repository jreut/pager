package main

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/save"
)

func TestTimeflag(t *testing.T) {
	edt, err := time.LoadLocation("America/New_York")
	assert.Nil(t, err)
	for _, tt := range []struct {
		arg  string
		want time.Time
	}{
		{
			arg:  "2022-10-30T00:06:00-04:00",
			want: time.Date(2022, 10, 30, 0, 6, 0, 0, edt),
		},
		{
			arg: "2022-10-30T00:06-04:00",
		},
		{
			arg: "2022-10-30T00:06:00EDT",
		},
	} {
		var val time.Time
		f := timeflag{&val}
		err := f.Set(tt.arg)
		var empty time.Time
		if tt.want != empty {
			assert.Nil(t, err)
			assert.Cmp(t, tt.want, *f.Time)
		} else {
			assert.Error(t, "cannot parse", err)
		}
	}

	dt := time.Date(2022, 10, 30, 0, 22, 0, 0, edt)
	got := (timeflag{&dt}).String()
	assert.Cmp(t, "2022-10-30T00:22:00-04:00", got)
}

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
	schema, err := os.ReadFile("schema.sql")
	assert.Nil(t, err)
	_, err = db.ExecContext(ctx, string(schema))
	assert.Nil(t, err)
	return db
}

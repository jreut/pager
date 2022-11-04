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

func TestAddInterval(t *testing.T) {
	ctx := context.Background()
	q := save.New(testdb(t, ctx))

	const alice = "alice"
	assert.Nil(t, q.AddPerson(ctx, alice))

	t0 := time.Date(2022, 10, 31, 13, 25, 42, 12345, time.UTC)
	t1 := time.Date(2022, 11, 3, 13, 25, 42, 12345, time.UTC)

	assert.Error(t, "CHECK constraint failed.*kind", q.AddInterval(ctx, save.AddIntervalParams{
		Person:    alice,
		StartAt:   t0,
		EndBefore: t1,
	}))

	assert.Error(t, "CHECK constraint failed.*kind", q.AddInterval(ctx, save.AddIntervalParams{
		Person:    alice,
		StartAt:   t0,
		EndBefore: t1,
		Kind:      "unknown",
	}))

	assert.Error(t, "^FOREIGN KEY constraint failed$", q.AddInterval(ctx, save.AddIntervalParams{
		Person:    "unknown",
		StartAt:   t0,
		EndBefore: t1,
		Kind:      save.IntervalKindShift,
	}))

	assert.Nil(t, q.AddInterval(ctx, save.AddIntervalParams{
		Person:    alice,
		StartAt:   t0,
		EndBefore: t1,
		Kind:      save.IntervalKindShift,
	}))

	assert.Nil(t, q.AddInterval(ctx, save.AddIntervalParams{
		Person:    alice,
		StartAt:   t0,
		EndBefore: t1,
		Kind:      save.IntervalKindExclusion,
	}))

	// No uniqueness constraint
	assert.Nil(t, q.AddInterval(ctx, save.AddIntervalParams{
		Person:    alice,
		StartAt:   t0,
		EndBefore: t1,
		Kind:      save.IntervalKindExclusion,
	}))

	// Backwards doesn't work
	assert.Error(t, "CHECK constraint failed", q.AddInterval(ctx, save.AddIntervalParams{
		Person:    alice,
		StartAt:   t1,
		EndBefore: t0,
		Kind:      save.IntervalKindExclusion,
	}))
}

func TestListIntervals(t *testing.T) {
	ctx := context.Background()
	q := save.New(testdb(t, ctx))

	var (
		t0 = time.Unix(0, 0).In(time.UTC)
		t1 = t0.Add(1 * time.Minute)
		t2 = t0.Add(2 * time.Minute)
		t4 = t0.Add(4 * time.Minute)
	)

	const alice = "alice"

	before := []save.AddIntervalParams{
		{Person: alice, StartAt: t0, EndBefore: t1, Kind: save.IntervalKindShift},
		{Person: alice, StartAt: t2, EndBefore: t4, Kind: save.IntervalKindShift},
	}
	assert.Nil(t, q.AddPerson(ctx, alice))

	for _, i := range before {
		assert.Nil(t, q.AddInterval(ctx, i))
	}

	got, err := q.ListIntervals(ctx, save.IntervalKindExclusion)
	assert.Nil(t, err)
	assert.Cmp(t, []save.Interval(nil), got)

	got, err = q.ListIntervals(ctx, save.IntervalKindShift)
	assert.Nil(t, err)
	assert.Cmp(t, []save.Interval{
		{Person: alice, StartAt: t0, EndBefore: t1, Kind: save.IntervalKindShift},
		{Person: alice, StartAt: t2, EndBefore: t4, Kind: save.IntervalKindShift},
	}, got)
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

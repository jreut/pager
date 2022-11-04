package cmd_test

import (
	"context"
	"database/sql"
	"net/url"
	"testing"
	"time"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/cmd"
	"github.com/jreut/pager/v2/pkg/save"
)

func TestAddInterval(t *testing.T) {
	var (
		t0 = time.Unix(0, 0).In(time.UTC)
		t1 = t0.Add(1 * time.Minute)
		t2 = t0.Add(2 * time.Minute)
	)

	const (
		alice = "alice"
		bob   = "bob"
	)

	ctx := context.Background()
	for _, tt := range []struct {
		label     string
		intervals []save.Interval
		arg       save.AddIntervalParams
		err       string
	}{
		{
			label:     "no existing intervals",
			intervals: nil,
			arg: save.AddIntervalParams{
				Person:    alice,
				StartAt:   t0,
				EndBefore: t1,
				Kind:      save.IntervalKindShift,
			},
			err: "",
		},
		{
			label: "overlapping interval with different person",
			intervals: []save.Interval{
				{Person: bob, StartAt: t0, EndBefore: t1, Kind: save.IntervalKindShift},
			},
			arg: save.AddIntervalParams{
				Person:    alice,
				StartAt:   t0,
				EndBefore: t1,
				Kind:      save.IntervalKindExclusion,
			},
			err: "",
		},
		{
			label: "exclusion over an exclusion is ok",
			intervals: []save.Interval{
				{Person: alice, StartAt: t0, EndBefore: t1, Kind: save.IntervalKindExclusion},
			},
			arg: save.AddIntervalParams{
				Person:    alice,
				StartAt:   t0,
				EndBefore: t2,
				Kind:      save.IntervalKindExclusion,
			},
			err: "",
		},
		{
			label: "shift over a shift is ok",
			intervals: []save.Interval{
				{Person: alice, StartAt: t0, EndBefore: t1, Kind: save.IntervalKindShift},
			},
			arg: save.AddIntervalParams{
				Person:    alice,
				StartAt:   t0,
				EndBefore: t2,
				Kind:      save.IntervalKindShift,
			},
			err: "",
		},
		{
			label: "shift over an exclusion is not ok",
			intervals: []save.Interval{
				{Person: alice, StartAt: t0, EndBefore: t1, Kind: save.IntervalKindExclusion},
			},
			arg: save.AddIntervalParams{
				Person:    alice,
				StartAt:   t0,
				EndBefore: t2,
				Kind:      save.IntervalKindShift,
			},
			err: `cannot schedule SHIFT.*"alice".*over existing EXCLUSION`,
		},
		{
			label: "exclusion over a shift is not ok",
			intervals: []save.Interval{
				{Person: alice, StartAt: t0, EndBefore: t1, Kind: save.IntervalKindShift},
			},
			arg: save.AddIntervalParams{
				Person:    alice,
				StartAt:   t0,
				EndBefore: t2,
				Kind:      save.IntervalKindExclusion,
			},
			err: `cannot schedule EXCLUSION.*"alice".*over existing SHIFT`,
		},
	} {
		t.Run(tt.label, func(t *testing.T) {
			q := save.New(testdb(t, ctx))
			assert.Nil(t, q.AddPerson(ctx, alice))
			assert.Nil(t, q.AddPerson(ctx, bob))
			for _, i := range tt.intervals {
				q.AddInterval(ctx, save.AddIntervalParams(i))
			}
			err := cmd.AddInterval(ctx, q, tt.arg)
			if tt.err == "" {
				assert.Nil(t, err)
			} else {
				assert.Error(t, tt.err, err)
			}
		})
	}
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

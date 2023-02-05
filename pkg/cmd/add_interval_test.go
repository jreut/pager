package cmd_test

import (
	"context"
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
			ctx := context.Background()
			q := save.New(testdb(t, ctx))
			const schedule = "default"
			assert.Nil(t, q.AddSchedule(ctx, schedule))
			for _, i := range tt.intervals {
				i.Schedule = schedule
				q.AddInterval(ctx, save.AddIntervalParams(i))
			}
			tt.arg.Schedule = schedule
			err := cmd.AddInterval(ctx, q, tt.arg, false)
			if tt.err == "" {
				assert.Nil(t, err)
			} else {
				assert.Error(t, tt.err, err)
			}
		})
	}
}

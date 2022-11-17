package cmd_test

import (
	"context"
	"testing"
	"time"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/cmd"
	"github.com/jreut/pager/v2/pkg/save"
)

func TestShowSchedule(t *testing.T) {
	const (
		s1 = "s1"
		s2 = "s2"
	)
	const (
		alice = "alice"
		bob   = "bob"
	)
	var (
		t0 = time.Unix(0, 0).In(time.UTC)
		t1 = t0.Add(1 * time.Minute)
		t2 = t0.Add(2 * time.Minute)
		t3 = t0.Add(3 * time.Minute)
		t4 = t0.Add(4 * time.Minute)
	)

	for _, tt := range []struct {
		setup, want []save.Interval
		schedule    string
		start, end  time.Time
	}{
		{},
		{
			setup: []save.Interval{
				{Schedule: s1, Person: alice, StartAt: t0, EndBefore: t1, Kind: save.IntervalKindShift},
			},
			want: []save.Interval{
				{Schedule: s1, Person: alice, StartAt: t0, EndBefore: t1, Kind: save.IntervalKindShift},
			},
			schedule: s1, start: t0, end: t1,
		},
		{
			setup: []save.Interval{
				{Schedule: s1, Person: alice, StartAt: t1, EndBefore: t2, Kind: save.IntervalKindShift},
			},
			want: []save.Interval{
				{Schedule: s1, Person: alice, StartAt: t1, EndBefore: t2, Kind: save.IntervalKindShift},
			},
			schedule: s1, start: t0, end: t3,
		},
		{
			setup: []save.Interval{
				{Schedule: s1, Person: alice, StartAt: t1, EndBefore: t2, Kind: save.IntervalKindShift},
				{Schedule: s1, Person: alice, StartAt: t2, EndBefore: t3, Kind: save.IntervalKindExclusion},
			},
			want: []save.Interval{
				{Schedule: s1, Person: alice, StartAt: t1, EndBefore: t2, Kind: save.IntervalKindShift},
			},
			schedule: s1, start: t0, end: t3,
		},
		{
			setup: []save.Interval{
				{Schedule: s1, Person: alice, StartAt: t1, EndBefore: t2, Kind: save.IntervalKindShift},
				{Schedule: s2, Person: alice, StartAt: t2, EndBefore: t3, Kind: save.IntervalKindShift},
			},
			want: []save.Interval{
				{Schedule: s1, Person: alice, StartAt: t1, EndBefore: t2, Kind: save.IntervalKindShift},
			},
			schedule: s1, start: t0, end: t3,
		},
		{
			setup: []save.Interval{
				{Schedule: s1, Person: alice, StartAt: t0, EndBefore: t2, Kind: save.IntervalKindShift},
				{Schedule: s1, Person: bob, StartAt: t2, EndBefore: t4, Kind: save.IntervalKindShift},
			},
			want: []save.Interval{
				{Schedule: s1, Person: alice, StartAt: t1, EndBefore: t2, Kind: save.IntervalKindShift},
				{Schedule: s1, Person: bob, StartAt: t2, EndBefore: t3, Kind: save.IntervalKindShift},
			},
			schedule: s1, start: t1, end: t3,
		},
	} {
		t.Run("", func(t *testing.T) {
			ctx := context.Background()
			q := save.New(testdb(t, ctx))
			assert.Nil(t, q.AddSchedule(ctx, s1))
			assert.Nil(t, q.AddSchedule(ctx, s2))

			for _, i := range tt.setup {
				assert.Nil(t, q.AddInterval(ctx, save.AddIntervalParams(i)))
			}
			got, err := cmd.ShowSchedule(ctx, q, tt.schedule, tt.start, tt.end)
			assert.Nil(t, err)
			assert.Cmp(t, tt.want, got)
		})
	}
}

package save_test

import (
	"context"
	"testing"
	"time"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/save"
)

func TestListIntervals(t *testing.T) {
	ctx := context.Background()
	q := save.New(testdb(t, ctx))

	var (
		t0 = time.Unix(0, 0).In(time.UTC)
		t1 = t0.Add(1 * time.Minute)
		t2 = t0.Add(2 * time.Minute)
		t3 = t0.Add(3 * time.Minute)
		t4 = t0.Add(4 * time.Minute)
		t5 = t0.Add(5 * time.Minute)
		t6 = t0.Add(6 * time.Minute)
		t7 = t0.Add(7 * time.Minute)
	)

	const (
		s1 = "s1"
		s2 = "s2"
	)
	assert.Nil(t, q.AddSchedule(ctx, s1))
	assert.Nil(t, q.AddSchedule(ctx, s2))
	const (
		alice   = "alice"
		bob     = "bob"
		cindy   = "cindy"
		daria   = "daria"
		evan    = "evan"
		felix   = "felix"
		georgia = "georgia"
		helen   = "helen"
	)
	assert.Nil(t, q.AddPerson(ctx, alice))
	assert.Nil(t, q.AddPerson(ctx, bob))
	assert.Nil(t, q.AddPerson(ctx, cindy))
	assert.Nil(t, q.AddPerson(ctx, daria))
	assert.Nil(t, q.AddPerson(ctx, evan))
	assert.Nil(t, q.AddPerson(ctx, felix))
	assert.Nil(t, q.AddPerson(ctx, georgia))
	assert.Nil(t, q.AddPerson(ctx, helen))

	before := []save.AddIntervalParams{
		{Schedule: s1, StartAt: t0, EndBefore: t1, Kind: save.IntervalKindShift, Person: alice},
		{Schedule: s1, StartAt: t0, EndBefore: t2, Kind: save.IntervalKindShift, Person: bob},
		{Schedule: s1, StartAt: t1, EndBefore: t4, Kind: save.IntervalKindShift, Person: cindy},
		{Schedule: s1, StartAt: t3, EndBefore: t6, Kind: save.IntervalKindShift, Person: daria},
		{Schedule: s1, StartAt: t5, EndBefore: t7, Kind: save.IntervalKindShift, Person: evan},
		{Schedule: s1, StartAt: t6, EndBefore: t7, Kind: save.IntervalKindShift, Person: felix},
		{Schedule: s1, StartAt: t2, EndBefore: t3, Kind: save.IntervalKindShift, Person: georgia},
		{Schedule: s1, StartAt: t4, EndBefore: t5, Kind: save.IntervalKindShift, Person: helen},
	}

	for _, i := range before {
		assert.Nil(t, q.AddInterval(ctx, i))
	}

	{
		got, err := q.ListIntervals(ctx, save.ListIntervalsParams{
			Kind:     save.IntervalKindExclusion,
			Schedule: s1,
			StartAt:  t0, EndBefore: t7,
		})
		assert.Nil(t, err)
		assert.Cmp(t, []save.Interval(nil), got)
	}
	{
		got, err := q.ListIntervals(ctx, save.ListIntervalsParams{
			Kind:     save.IntervalKindShift,
			Schedule: s2,
			StartAt:  t0, EndBefore: t7,
		})
		assert.Nil(t, err)
		assert.Cmp(t, []save.Interval(nil), got)
	}

	{
		got, err := q.ListIntervals(ctx, save.ListIntervalsParams{
			Kind:     save.IntervalKindShift,
			Schedule: s1,
			StartAt:  t2, EndBefore: t5,
		})
		assert.Nil(t, err)
		assert.Cmp(t, []save.Interval{
			{Person: cindy, Schedule: s1, StartAt: t1, EndBefore: t4, Kind: save.IntervalKindShift},
			{Person: daria, Schedule: s1, StartAt: t3, EndBefore: t6, Kind: save.IntervalKindShift},
			{Person: georgia, Schedule: s1, StartAt: t2, EndBefore: t3, Kind: save.IntervalKindShift},
			{Schedule: s1, StartAt: t4, EndBefore: t5, Kind: save.IntervalKindShift, Person: helen},
		}, got)
	}
}

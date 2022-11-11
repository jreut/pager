package cmd

import (
	"context"
	"time"

	"github.com/jreut/pager/v2/pkg/interval"
	"github.com/jreut/pager/v2/pkg/save"
)

func ShowSchedule(ctx context.Context, q *save.Queries, schedule string, start time.Time, end time.Time) ([]save.Interval, error) {
	out, err := q.ListIntervals(ctx, save.ListIntervalsParams{
		Schedule:  schedule,
		Kind:      save.IntervalKindShift,
		StartAt:   start,
		EndBefore: end,
	})
	if err != nil {
		return nil, err
	}
	out = interval.Flatten(out)
	for i := range out {
		if out[i].StartAt.Before(start) {
			out[i].StartAt = start
		}
		if out[i].EndBefore.After(end) {
			out[i].EndBefore = end
		}
	}

	return out, nil
}

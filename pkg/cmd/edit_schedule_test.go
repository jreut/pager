package cmd_test

import (
	"context"
	"testing"
	"time"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/save"
)

func TestEditSchedule(t *testing.T) {
	ctx := context.Background()
	q := save.New(testdb(t, ctx))

	t0 := time.Unix(0, 0).In(time.UTC)

	const (
		alice = "alice"
		s1    = "s1"
	)

	// Can't add an interval for a nonexistent schedule.
	assert.Error(t, "FOREIGN KEY constraint failed", q.AddInterval(ctx, save.AddIntervalParams{
		Person:    alice,
		Schedule:  s1,
		StartAt:   t0,
		EndBefore: t0.Add(time.Hour),
		Kind:      save.IntervalKindShift,
	}))

	assert.Nil(t, q.AddSchedule(ctx, s1))

	assert.Nil(t, q.AddInterval(ctx, save.AddIntervalParams{
		Person:    alice,
		Schedule:  s1,
		StartAt:   t0,
		EndBefore: t0.Add(time.Hour),
		Kind:      save.IntervalKindShift,
	}))
}

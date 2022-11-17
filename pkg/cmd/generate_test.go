package cmd_test

import (
	"context"
	"encoding/csv"
	"testing"
	"time"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/cmd"
	"github.com/jreut/pager/v2/pkg/save"
)

func TestGenerate(t *testing.T) {
	t.Run(cmd.Style, func(t *testing.T) {
		ctx := context.Background()
		db := testdb(t, ctx)
		q := save.New(db)

		const schedule = "schedule"
		assert.Nil(t, q.AddSchedule(ctx, schedule))
		const (
			alice = "alice"
			bob   = "bob"
			cindy = "cindy"
		)
		var (
			initialAdd      = time.Date(2022, 11, 16, 0, 0, 0, 0, time.UTC)
			cindyJoins      = time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC)
			aliceBigExclude = time.Date(2022, 12, 4, 0, 0, 0, 0, time.UTC)
			aliceLeaves     = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
			bobSmallExclude = time.Date(2023, 1, 3, 11, 15, 0, 0, time.UTC)
			generateStart   = time.Date(2022, 11, 22, 0, 0, 0, 0, time.UTC)
			generateEnd     = time.Date(2023, 1, 13, 7, 42, 0, 0, time.UTC)
		)
		assert.Nil(t, cmd.EditSchedule(ctx, q, schedule, []cmd.Action{
			{Who: alice, At: initialAdd, Kind: save.EventKindAdd},
			{Who: bob, At: initialAdd, Kind: save.EventKindAdd},
			{Who: cindy, At: cindyJoins, Kind: save.EventKindAdd},
			{Who: alice, At: aliceLeaves, Kind: save.EventKindRemove},
		}))
		for _, e := range []struct {
			who        string
			start, end time.Time
		}{
			{
				who:   alice,
				start: aliceBigExclude,
				end:   aliceBigExclude.AddDate(0, 0, 8),
			},
			{
				who:   bob,
				start: bobSmallExclude,
				end:   bobSmallExclude.Add(2 * time.Hour),
			},
		} {
			assert.Nil(t, q.AddInterval(ctx, save.AddIntervalParams{
				Person:    e.who,
				Schedule:  schedule,
				StartAt:   e.start,
				EndBefore: e.end,
				Kind:      save.IntervalKindExclusion,
			}))
		}

		assert.Nil(t, cmd.Generate(ctx, q, schedule, cmd.Style, generateStart, generateEnd))
		got, err := cmd.ShowSchedule(ctx, q, schedule, generateStart.AddDate(0, 0, -1), generateEnd.AddDate(0, 0, 1))
		assert.Nil(t, err)
		w := csv.NewWriter(assert.Golden(t, "generated.csv"))
		t.Cleanup(func() {
			w.Flush()
			assert.Nil(t, w.Error())
		})
		for _, i := range got {
			assert.Nil(t, w.Write([]string{
				i.StartAt.Format(time.RFC3339),
				i.EndBefore.Format(time.RFC3339),
				i.Person,
			}))
		}
	})
}

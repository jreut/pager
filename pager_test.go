package pager

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

const (
	alice  Person = "alice"
	bob           = "bob"
	cindy         = "cindy"
	delila        = "delila"
)

func TestNewSchedule(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		bal := balance{
			alice:  -1 * 24 * time.Hour,
			bob:    0,
			cindy:  3 * 24 * time.Hour,
			delila: 0,
		}
		start := time.Date(2022, 2, 26, 17, 24, 8, 0, time.UTC)
		want := schedule{
			{alice, start}, // alice=-1d, bob=0d, delila=0d, cindy=3d
			{bob, time.Date(2022, 2, 28, 12, 0, 0, 0, nyc)},    // bob=0d, delila=0d, alice=~1d, cindy=3d
			{delila, time.Date(2022, 3, 4, 12, 0, 0, 0, nyc)},  // delila=0d, alice=~1d, cindy=3d, bob=4d
			{alice, time.Date(2022, 3, 7, 12, 0, 0, 0, nyc)},   // alice=~1d, cindy=3d, delila=3d, bob=4d
			{cindy, time.Date(2022, 3, 11, 12, 0, 0, 0, nyc)},  // cindy=3d, delila=3d, bob=4d, alice=~5d
			{delila, time.Date(2022, 3, 14, 12, 0, 0, 0, nyc)}, // delila=3d, bob=4d, alice=~5d, cindy=6d
			{bob, time.Date(2022, 3, 18, 12, 0, 0, 0, nyc)},    // bob=4d, alice=~5d, cindy=6d, delila=7d
		}
		got, _ := newschedule(bal, nil, start, 21*24*time.Hour)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("- want, + got:\n%s", diff)
		}
	})
	t.Run("exclusions", func(t *testing.T) {
		bal := balance{alice: 0, bob: 0}
		start := time.Date(2022, 2, 25, 12, 0, 0, 0, nyc)
		got, newbal := newschedule(bal, exclusions{
			{alice, time.Date(2022, 02, 24, 0, 0, 0, 0, nyc), 2 * 24 * time.Hour},
		}, start, 8*24*time.Hour)
		want := schedule{
			{bob, start}, // alice is excluded during this time
			{alice, start.AddDate(0, 0, 3)},
			{bob, start.AddDate(0, 0, 7)},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("- want, + got:\n%s", diff)
		}
		if diff := cmp.Diff(balance{
			bob:   6 * 24 * time.Hour,
			alice: 4 * 24 * time.Hour,
		}, newbal); diff != "" {
			t.Fatalf("- want, + got:\n%s", diff)
		}
	})
}

func TestIntervals(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		next := func(t time.Time) time.Time {
			return t.Add(time.Hour)
		}
		start := time.Unix(0, 0)
		d := (2 * time.Hour) + (15 * time.Minute)
		want := []interval{
			{start.Add(0 * time.Hour), start.Add(1 * time.Hour)},
			{start.Add(1 * time.Hour), start.Add(2 * time.Hour)},
			{start.Add(2 * time.Hour), start.Add(3 * time.Hour)},
		}
		got := intervals(start, d, next)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("- want, + got:\n%s", diff)
		}
	})
	t.Run("nextbreakpoint", func(t *testing.T) {
		start := time.Date(2022, 2, 26, 9, 0, 0, 0, nyc)
		d := 10 * 24 * time.Hour
		want := []interval{
			{start, time.Date(2022, 2, 28, 12, 0, 0, 0, nyc)},
			{time.Date(2022, 2, 28, 12, 0, 0, 0, nyc), time.Date(2022, 3, 4, 12, 0, 0, 0, nyc)},
			{time.Date(2022, 3, 4, 12, 0, 0, 0, nyc), time.Date(2022, 3, 7, 12, 0, 0, 0, nyc)},
			{time.Date(2022, 3, 7, 12, 0, 0, 0, nyc), time.Date(2022, 3, 11, 12, 0, 0, 0, nyc)},
		}
		got := intervals(start, d, nextbreakpoint)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("- want, + got:\n%s", diff)
		}
	})
}

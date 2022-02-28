package schedule

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
		got := newschedule(config{
			start:    start,
			duration: 21 * 24 * time.Hour,
			balance:  bal,
		})
		if diff := cmp.Diff(want, got.Schedule); diff != "" {
			t.Fatalf("- want, + got:\n%s", diff)
		}
	})
	t.Run("exclusions", func(t *testing.T) {
		bal := balance{alice: 0, bob: 0}
		start := time.Date(2022, 2, 25, 12, 0, 0, 0, nyc)
		got := newschedule(config{
			balance: bal,
			exclusions: exclusions{
				{alice, Interval{start, 3 * 24 * time.Hour}},
			},
			start:    start,
			duration: 8 * 24 * time.Hour,
		})
		want := result{schedule{
			{bob, start}, // alice is excluded during this time
			{alice, start.AddDate(0, 0, 3)},
			{bob, start.AddDate(0, 0, 7)},
		}, balance{
			bob:   6 * 24 * time.Hour,
			alice: 4 * 24 * time.Hour,
		}}
		if diff := cmp.Diff(want, got); diff != "" {
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
		want := []Interval{
			{start.Add(0 * time.Hour), time.Hour},
			{start.Add(1 * time.Hour), time.Hour},
			{start.Add(2 * time.Hour), time.Hour},
		}
		got := intervals(start, d, next)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("- want, + got:\n%s", diff)
		}
	})
	t.Run("nextbreakpoint", func(t *testing.T) {
		start := time.Date(2022, 2, 26, 9, 0, 0, 0, nyc)
		d := 10 * 24 * time.Hour
		want := []Interval{
			{start, 51 * time.Hour},
			{time.Date(2022, 2, 28, 12, 0, 0, 0, nyc), 4 * 24 * time.Hour},
			{time.Date(2022, 3, 4, 12, 0, 0, 0, nyc), 3 * 24 * time.Hour},
			{time.Date(2022, 3, 7, 12, 0, 0, 0, nyc), 4 * 24 * time.Hour},
		}
		got := intervals(start, d, nextbreakpoint)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("- want, + got:\n%s", diff)
		}
	})
}

func TestInterval(t *testing.T) {
	var (
		t0 = time.Unix(0, 0)
		t1 = t0.Add(time.Minute)
		t2 = t1.Add(time.Minute)
		t3 = t2.Add(time.Minute)
		t4 = t3.Add(time.Minute)
	)
	var (
		// t0   t1   t2   t3   t4
		// [a---)
		//      [b--------)
		// [c------------------)
		//           [d--------)
		a = Interval{t0, t1.Sub(t0)}
		b = Interval{t1, t3.Sub(t1)}
		c = Interval{t0, t4.Sub(t0)}
		d = Interval{t2, t4.Sub(t2)}
	)
	t.Run("contains", func(t *testing.T) {
		for _, tt := range []struct {
			time.Time
			bool
		}{
			{t0, false},
			{t1, true},
			{t2, true},
			{t3, false},
			{t4, false},
		} {
			t.Run("", func(t *testing.T) {
				if b.Contains(tt.Time) != tt.bool {
					t.Fatalf("want %t, got %t", tt.bool, !tt.bool)
				}
			})
		}
	})
	t.Run("conjoint", func(t *testing.T) {
		for _, tt := range []struct {
			a, b Interval
			want bool
		}{
			{a, a, true},
			{a, b, false},
			{a, c, true},
			{a, d, false},
			{b, c, true},
			{b, d, true},
			{c, d, true},
		} {
			t.Run("", func(t *testing.T) {
				left := Conjoint(tt.a, tt.b)
				right := Conjoint(tt.b, tt.a)
				if (left != right) || (left != tt.want) {
					t.Fatalf("want %t, got %t,%t", tt.want, left, right)
				}
			})
		}
	})
}

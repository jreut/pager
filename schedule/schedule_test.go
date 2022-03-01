package schedule

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

const (
	alice  = "alice"
	bob    = "bob"
	cindy  = "cindy"
	delila = "delila"
)

func TestNewSchedule(t *testing.T) {
	t.Run("with our funky scheduling", func(t *testing.T) {
		bal := Balance{
			alice:  -1 * 24 * time.Hour,
			bob:    0,
			cindy:  3 * 24 * time.Hour,
			delila: 0,
		}
		start := time.Date(2022, 2, 26, 17, 24, 8, 0, time.UTC)
		want := Schedule{
			{alice, start}, // alice=-1d, bob=0d, delila=0d, cindy=3d
			{bob, time.Date(2022, 2, 28, 12, 0, 0, 0, NYC)},    // bob=0d, delila=0d, alice=~1d, cindy=3d
			{delila, time.Date(2022, 3, 4, 12, 0, 0, 0, NYC)},  // delila=0d, alice=~1d, cindy=3d, bob=4d
			{alice, time.Date(2022, 3, 7, 12, 0, 0, 0, NYC)},   // alice=~1d, cindy=3d, delila=3d, bob=4d
			{cindy, time.Date(2022, 3, 11, 12, 0, 0, 0, NYC)},  // cindy=3d, delila=3d, bob=4d, alice=~5d
			{delila, time.Date(2022, 3, 14, 12, 0, 0, 0, NYC)}, // delila=3d, bob=4d, alice=~5d, cindy=6d
			{bob, time.Date(2022, 3, 18, 12, 0, 0, 0, NYC)},    // bob=4d, alice=~5d, cindy=6d, delila=7d
		}
		got := Builder{
			Next:     MondayFridayShifts,
			Interval: Interval{start, 21 * 24 * time.Hour},
			Balance:  bal,
		}.Build()
		if diff := cmp.Diff(want, got.Schedule); diff != "" {
			t.Fatalf("- want, + got:\n%s", diff)
		}
	})
	t.Run("exclusions", func(t *testing.T) {
		bal := Balance{alice: 0, bob: 0}
		start := time.Unix(0, 0)
		got := Builder{
			Next:    func(t time.Time) time.Time { return t.Add(time.Hour) },
			Balance: bal,
			Exclusions: []Exclusion{
				// 0h  1h  2h  3h  4h
				//  xx
				{alice, Interval{start.Add(time.Hour / 4), time.Hour / 2}},
				// 0h  1h  2h  3h  4h
				//     xxxxxx
				{alice, Interval{start.Add(time.Hour), 3 * time.Hour / 2}},
				// 0h  1h  2h  3h  4h
				//                 x
				{alice, Interval{start.Add(4 * time.Hour), time.Hour / 4}},
			},
			Interval: Interval{start, 6 * time.Hour},
		}.Build()
		want := Result{Schedule{
			{bob, start.Add(0 * time.Hour)},   // alice is excluded in the middle of this interval
			{bob, start.Add(1 * time.Hour)},   // alice is excluded from the start of this interval...
			{bob, start.Add(2 * time.Hour)},   // ...to the middle of this interval
			{alice, start.Add(3 * time.Hour)}, // alice has an exclusion starting at the end of this interval
			{bob, start.Add(4 * time.Hour)},
			{alice, start.Add(5 * time.Hour)},
		}, Balance{
			bob:   4 * time.Hour,
			alice: 2 * time.Hour,
		}}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("- want, + got:\n%s", diff)
		}
	})
}

func TestMarshalSchedule(t *testing.T) {
	before := Schedule{
		{alice, time.Date(2022, 2, 26, 12, 0, 0, 0, NYC)},
		{bob, time.Date(2022, 2, 28, 12, 0, 0, 0, NYC)},
		{delila, time.Date(2022, 3, 4, 12, 0, 0, 0, NYC)},
		{alice, time.Date(2022, 3, 7, 12, 0, 0, 0, NYC)},
	}
	var buf bytes.Buffer
	if err := before.WriteCSV(&buf); err != nil {
		t.Fatal(err)
	}
	want := `2022-02-26T12:00:00-05:00,alice
2022-02-28T12:00:00-05:00,bob
2022-03-04T12:00:00-05:00,delila
2022-03-07T12:00:00-05:00,alice
`
	if diff := cmp.Diff(want, buf.String()); diff != "" {
		t.Fatalf("- want, + got:\n%s", diff)
	}
	after, err := FromCSV(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(before, after); diff != "" {
		t.Fatalf("- want, + got:\n%s", diff)
	}
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
	t.Run("invalid", func(t *testing.T) {
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("expected panic")
			}
			t.Logf("panic recovery: %+v", p)
		}()
		intervals(time.Unix(0, 0), time.Hour, func(t time.Time) time.Time {
			return t.Add(-time.Minute)
		})
	})
	t.Run("MondayFridayShifts", func(t *testing.T) {
		start := time.Date(2022, 2, 26, 9, 0, 0, 0, NYC)
		d := 10 * 24 * time.Hour
		want := []Interval{
			{start, 51 * time.Hour},
			{time.Date(2022, 2, 28, 12, 0, 0, 0, NYC), 4 * 24 * time.Hour},
			{time.Date(2022, 3, 4, 12, 0, 0, 0, NYC), 3 * 24 * time.Hour},
			{time.Date(2022, 3, 7, 12, 0, 0, 0, NYC), 4 * 24 * time.Hour},
		}
		got := intervals(start, d, MondayFridayShifts)
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
				if b.contains(tt.Time) != tt.bool {
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
				left := conjoint(tt.a, tt.b)
				right := conjoint(tt.b, tt.a)
				if (left != right) || (left != tt.want) {
					t.Fatalf("want %t, got %t,%t", tt.want, left, right)
				}
			})
		}
	})
}

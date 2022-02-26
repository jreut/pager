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

func TestSchedule(t *testing.T) {
	bal := balance{
		alice:  -1 * 24 * time.Hour,
		bob:    0,
		cindy:  3 * 24 * time.Hour,
		delila: 0,
	}
	start := time.Date(2022, 2, 26, 17, 24, 8, 0, time.UTC)
	s := newschedule(bal, start, 21*24*time.Hour /* 21d */)
	want := schedule{
		{alice, start}, // alice=-1d, bob=0d, delila=0d, cindy=3d
		{bob, time.Date(2022, 2, 28, 12, 0, 0, 0, nyc)}, // bob=0d, delila=0d, alice=~1d, cindy=3d
		{delila, time.Date(2022, 3, 4, 12, 0, 0, 0, nyc)}, // delila=0d, alice=~1d, cindy=3d, bob=4d
		{alice, time.Date(2022, 3, 7, 12, 0, 0, 0, nyc)}, // alice=~1d, cindy=3d, delila=3d, bob=4d
		{cindy, time.Date(2022, 3, 11, 12, 0, 0, 0, nyc)}, // cindy=3d, delila=3d, bob=4d, alice=~5d
		{delila, time.Date(2022, 3, 14, 12, 0, 0, 0, nyc)}, // delila=3d, bob=4d, alice=~5d, cindy=6d
		{bob, time.Date(2022, 3, 18, 12, 0, 0, 0, nyc)}, // bob=4d, alice=~5d, cindy=6d, delila=7d
	}
	if diff := cmp.Diff(want, s); diff != "" {
		t.Fatalf("- want, + got:\n%s", diff)
	}
}

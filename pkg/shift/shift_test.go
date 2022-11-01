package shift

import (
	"testing"
	"time"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/save"
)

const (
	alice = "alice"
	bob   = "bob"
	cindy = "cindy"
	dana  = "dana"
)

var (
	t0 = time.Unix(0, 0).In(time.UTC)
	t1 = t0.Add(1 * time.Minute)
	t2 = t0.Add(2 * time.Minute)
	t3 = t0.Add(3 * time.Minute)
	t4 = t0.Add(4 * time.Minute)
	t5 = t0.Add(5 * time.Minute)
)

func TestFlatten(t *testing.T) {
	for _, tt := range []struct {
		in, want []save.Shift
	}{
		{},
		{
			in: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t1},
			},
			want: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t1},
			},
		},
		{
			in: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t3},
				{Person: bob, StartAt: t1, EndBefore: t2},
			},
			want: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t1},
				{Person: bob, StartAt: t1, EndBefore: t2},
				{Person: alice, StartAt: t2, EndBefore: t3},
			},
		},
		{
			in: []save.Shift{
				{Person: bob, StartAt: t1, EndBefore: t2},
				{Person: cindy, StartAt: t0, EndBefore: t3},
				{Person: alice, StartAt: t0, EndBefore: t3},
			},
			want: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t3},
			},
		},
		{
			in: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t2},
				{Person: bob, StartAt: t1, EndBefore: t3},
			},
			want: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t1},
				{Person: bob, StartAt: t1, EndBefore: t3},
			},
		},
		// {
		// 	in: []save.Shift{
		// 		{Person: alice, StartAt: t0, EndBefore: t1},
		// 		{Person: alice, StartAt: t1, EndBefore: t2},
		// 	},
		// 	want: []save.Shift{
		// 		{Person: alice, StartAt: t0, EndBefore: t2},
		// 	},
		// },
		// {
		// 	in: []save.Shift{
		// 		{Person: alice, StartAt: t0, EndBefore: t2},
		// 		{Person: bob, StartAt: t1, EndBefore: t3},
		// 	},
		// 	want: []save.Shift{
		// 		{Person: alice, StartAt: t0, EndBefore: t1},
		// 		{Person: bob, StartAt: t1, EndBefore: t3},
		// 	},
		// },
		// {
		// 	in: []save.Shift{
		// 		{Person: bob, StartAt: t1, EndBefore: t3},
		// 		{Person: alice, StartAt: t0, EndBefore: t2},
		// 	},
		// 	want: []save.Shift{
		// 		{Person: alice, StartAt: t0, EndBefore: t2},
		// 		{Person: bob, StartAt: t2, EndBefore: t3},
		// 	},
		// },
		// {
		// 	in: []save.Shift{
		// 		{Person: alice, StartAt: t0, EndBefore: t2},
		// 		{Person: bob, StartAt: t2, EndBefore: t3},
		// 		{Person: cindy, StartAt: t3, EndBefore: t5},
		// 		{Person: dana, StartAt: t1, EndBefore: t4},
		// 	},
		// 	want: []save.Shift{
		// 		{Person: alice, StartAt: t0, EndBefore: t1},
		// 		{Person: dana, StartAt: t1, EndBefore: t4},
		// 		{Person: cindy, StartAt: t4, EndBefore: t5},
		// 	},
		// },
		// {
		// 	in: []save.Shift{
		// 		{Person: alice, StartAt: t0, EndBefore: t1},
		// 		{Person: bob, StartAt: t2, EndBefore: t3},
		// 	},
		// },
	} {
		t.Run("", func(t *testing.T) {
			got := Flatten(tt.in)
			assert.Cmp(t, tt.want, got)
		})
	}
}

func TestSortShifts(t *testing.T) {
	for _, tt := range []struct{ in, want []save.Shift }{
		{
			in: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t1},
				{Person: bob, StartAt: t1, EndBefore: t2},
				{Person: cindy, StartAt: t2, EndBefore: t3},
			},
			want: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t1},
				{Person: bob, StartAt: t1, EndBefore: t2},
				{Person: cindy, StartAt: t2, EndBefore: t3},
			},
		},
		{
			in: []save.Shift{
				{Person: bob, StartAt: t1, EndBefore: t2},
				{Person: alice, StartAt: t0, EndBefore: t1},
				{Person: cindy, StartAt: t2, EndBefore: t3},
			},
			want: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t1},
				{Person: bob, StartAt: t1, EndBefore: t2},
				{Person: cindy, StartAt: t2, EndBefore: t3},
			},
		},
		{
			in: []save.Shift{
				{Person: alice, StartAt: t2, EndBefore: t5},
				{Person: bob, StartAt: t1, EndBefore: t3},
				{Person: cindy, StartAt: t0, EndBefore: t2},
			},
			want: []save.Shift{
				{Person: alice, StartAt: t2, EndBefore: t5},
				{Person: bob, StartAt: t1, EndBefore: t3},
				{Person: cindy, StartAt: t0, EndBefore: t2},
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			got := make([]save.Shift, len(tt.in))
			copy(got, tt.in)
			sortshifts(got)
			assert.Cmp(t, tt.want, got)
		})
	}
}

func TestOverlap(t *testing.T) {
	for _, tt := range []struct {
		a, b save.Shift
		want bool
	}{
		{
			a:    save.Shift{StartAt: t0, EndBefore: t1},
			b:    save.Shift{StartAt: t1, EndBefore: t2},
			want: false,
		},
		{
			a:    save.Shift{StartAt: t0, EndBefore: t1},
			b:    save.Shift{StartAt: t0, EndBefore: t2},
			want: true,
		},
		{
			a:    save.Shift{StartAt: t0, EndBefore: t2},
			b:    save.Shift{StartAt: t1, EndBefore: t3},
			want: true,
		},
		{
			a:    save.Shift{StartAt: t0, EndBefore: t1},
			b:    save.Shift{StartAt: t2, EndBefore: t3},
			want: false,
		},
		{
			a:    save.Shift{StartAt: t0, EndBefore: t1},
			b:    save.Shift{StartAt: t0, EndBefore: t1},
			want: true,
		},
	} {
		t.Run("", func(t *testing.T) {
			got := overlap(tt.a, tt.b)
			assert.Cmp(t, tt.want, got)
			got2 := overlap(tt.b, tt.a)
			assert.Cmp(t, tt.want, got2)
		})
	}
}

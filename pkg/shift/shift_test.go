package shift

import (
	"fmt"
	"testing"
	"time"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/save"
)

var (
	t0 = time.Unix(0, 0).In(time.UTC)
	t1 = t0.Add(1 * time.Minute)
	t2 = t0.Add(2 * time.Minute)
	t3 = t0.Add(3 * time.Minute)
	t4 = t0.Add(4 * time.Minute)
	t5 = t0.Add(5 * time.Minute)
)

const (
	alice = "alice"
	bob   = "bob"
	cindy = "cindy"
	dana  = "dana"
)

func ExampleFlatten() {
	t0 := time.Date(2022, 11, 1, 12, 0, 0, 0, time.UTC)
	t2 := t0.AddDate(0, 0, 2)
	t3 := t0.AddDate(0, 0, 3)
	t4 := t0.AddDate(0, 0, 4)
	t7 := t0.AddDate(0, 0, 7)

	shifts := []save.Shift{
		{Person: "alice", StartAt: t0, EndBefore: t3},
		{Person: "bob", StartAt: t3, EndBefore: t7},
	}
	override := save.Shift{Person: "cindy", StartAt: t2, EndBefore: t4}
	out := Flatten(append(shifts, override))
	for _, o := range out {
		fmt.Printf(
			"%s works from %s to %s\n",
			o.Person,
			o.StartAt.Format("2 Jan"),
			o.EndBefore.Format("2 Jan"),
		)
	}
	// Output:
	// alice works from 1 Nov to 3 Nov
	// cindy works from 3 Nov to 5 Nov
	// bob works from 5 Nov to 8 Nov
}

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
		{
			in: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t1},
				{Person: alice, StartAt: t1, EndBefore: t2},
			},
			want: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t2},
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
		{
			in: []save.Shift{
				{Person: bob, StartAt: t1, EndBefore: t3},
				{Person: alice, StartAt: t0, EndBefore: t2},
			},
			want: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t2},
				{Person: bob, StartAt: t2, EndBefore: t3},
			},
		},
		{
			in: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t2},
				{Person: bob, StartAt: t2, EndBefore: t3},
				{Person: cindy, StartAt: t3, EndBefore: t5},
				{Person: dana, StartAt: t1, EndBefore: t4},
			},
			want: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t1},
				{Person: dana, StartAt: t1, EndBefore: t4},
				{Person: cindy, StartAt: t4, EndBefore: t5},
			},
		},
		{
			in: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t1},
				{Person: bob, StartAt: t4, EndBefore: t5},
				{Person: cindy, StartAt: t2, EndBefore: t3},
			},
			want: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t1},
				{Person: cindy, StartAt: t2, EndBefore: t3},
				{Person: bob, StartAt: t4, EndBefore: t5},
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			got := Flatten(tt.in)
			assert.Cmp(t, tt.want, got)
		})
	}
}

func TestBounds(t *testing.T) {
	for _, tt := range []struct {
		xs           []save.Shift
		y            save.Shift
		wantl, wantr int
	}{
		{
			xs:    nil,
			y:     save.Shift{StartAt: t0, EndBefore: t1},
			wantl: 0, wantr: 0,
		},
		{
			xs: []save.Shift{
				{StartAt: t0, EndBefore: t1},
			},
			y:     save.Shift{StartAt: t0, EndBefore: t1},
			wantl: 0, wantr: 1,
		},
		{
			xs: []save.Shift{
				{StartAt: t0, EndBefore: t2},
			},
			y:     save.Shift{StartAt: t1, EndBefore: t2},
			wantl: 0, wantr: 1,
		},
		{
			xs: []save.Shift{
				{StartAt: t0, EndBefore: t2},
			},
			y:     save.Shift{StartAt: t0, EndBefore: t1},
			wantl: 0, wantr: 1,
		},
		{
			xs: []save.Shift{
				{StartAt: t0, EndBefore: t1},
				{StartAt: t1, EndBefore: t2},
				{StartAt: t2, EndBefore: t3},
				{StartAt: t3, EndBefore: t4},
			},
			y:     save.Shift{StartAt: t1, EndBefore: t3},
			wantl: 1, wantr: 3,
		},
	} {
		t.Run("", func(t *testing.T) {
			gotl, gotr := bounds(tt.xs, tt.y)
			assert.Cmp(t,
				fmt.Sprintf("[%d,%d)", tt.wantl, tt.wantr),
				fmt.Sprintf("[%d,%d)", gotl, gotr),
			)
		})
	}
}

func TestCombine(t *testing.T) {
	for _, tt := range []struct {
		xs   []save.Shift
		y    save.Shift
		want []save.Shift
	}{
		{
			xs:   nil,
			y:    save.Shift{StartAt: t0, EndBefore: t1},
			want: []save.Shift{{StartAt: t0, EndBefore: t1}},
		},
		//  xs: [alice)
		//   y: [bob  )
		// out: [bob  )
		{
			xs: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t1},
			},
			y: save.Shift{Person: bob, StartAt: t0, EndBefore: t1},
			want: []save.Shift{
				{Person: bob, StartAt: t0, EndBefore: t1},
			},
		},
		//  xs: [alice)[bob  )
		//   y:    [cindy )
		// out: [a)[c    )[b )
		{
			xs: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t2},
				{Person: bob, StartAt: t2, EndBefore: t4},
			},
			y: save.Shift{Person: cindy, StartAt: t1, EndBefore: t3},
			want: []save.Shift{
				{Person: alice, StartAt: t0, EndBefore: t1},
				{Person: cindy, StartAt: t1, EndBefore: t3},
				{Person: bob, StartAt: t3, EndBefore: t4},
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			got := combine(tt.xs, tt.y)
			assert.Cmp(t, tt.want, got)
		})
	}
}

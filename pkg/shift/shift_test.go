package shift_test

import (
	"testing"
	"time"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/save"
	"github.com/jreut/pager/v2/pkg/shift"
)

func TestFlatten(t *testing.T) {
	const (
		alice = "alice"
		bob   = "bob"
	)
	var (
		t0 = time.Unix(0, 0)
		t1 = t0.AddDate(0, 0, 1)
		t2 = t0.AddDate(0, 0, 2)
		t3 = t0.AddDate(0, 0, 3)
	)
	for _, tt := range []struct {
		in, want []save.Shift
		err      string
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
				{Person: alice, StartAt: t0, EndBefore: t1},
				{Person: bob, StartAt: t2, EndBefore: t3},
			},
			err: "gap",
		},
	} {
		t.Run("", func(t *testing.T) {
			got, err := shift.Flatten(tt.in)
			assert.Cmp(t, tt.want, got)
			if tt.err == "" {
				assert.Nil(t, err)
			} else {
				assert.Error(t, tt.err, err)
			}
		})
	}
}

package main

import (
	"testing"
	"time"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/cmd"
)

func TestTimeflag(t *testing.T) {
	edt, err := time.LoadLocation("America/New_York")
	assert.Nil(t, err)
	for _, tt := range []struct {
		arg  string
		want time.Time
	}{
		{
			arg:  "2022-10-30T00:06:00-04:00",
			want: time.Date(2022, 10, 30, 0, 6, 0, 0, edt),
		},
		{
			arg: "2022-10-30T00:06-04:00",
		},
		{
			arg: "2022-10-30T00:06:00EDT",
		},
	} {
		var val, zero time.Time
		f := timeflag{&val}
		err := f.Set(tt.arg)
		if tt.want != zero {
			assert.Nil(t, err)
			assert.Cmp(t, tt.want, val)
		} else {
			assert.Error(t, "cannot parse", err)
		}
	}

	dt := time.Date(2022, 10, 30, 0, 22, 0, 0, edt)
	got := (timeflag{&dt}).String()
	assert.Cmp(t, "2022-10-30T00:22:00-04:00", got)
}

func TestActionsFlag(t *testing.T) {
	for _, tt := range []struct {
		arg  []string
		want []cmd.Action
	}{
		{},
		{
			arg: []string{"alice=2022-11-06T00:31:00Z"},
			want: []cmd.Action{
				{
					Kind: "add",
					Who:  "alice",
					At:   time.Date(2022, 11, 6, 0, 31, 0, 0, time.UTC),
				},
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			var val []cmd.Action
			f := actionsflag{"add", &val}
			for _, arg := range tt.arg {
				err := f.Set(arg)
				assert.Nil(t, err)
			}
			assert.Cmp(t, tt.want, val)
		})
	}
}

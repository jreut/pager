package main

import (
	"testing"
	"time"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/cmd"
)



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

package cli

import (
	"testing"
	"time"

	"github.com/jreut/pager/v2/pkg/assert"
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

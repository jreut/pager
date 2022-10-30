package main

import (
	"testing"
	"time"
)

func TestTimeflag(t *testing.T) {
	edt, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
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
		f := new(timeflag)
		err := f.Set(tt.arg)
		var empty time.Time
		if tt.want != empty {
			if err != nil {
				t.Fatalf("want %s, got an error: %v", tt.want, err)
			}
			if !tt.want.Equal(f.Time) {
				t.Fatalf("want %s == %s", tt.want, f)
			}
		} else {
			if err == nil {
				t.Fatalf("want an error, got %s", f.Time)
			}
		}
	}

	got := (&timeflag{time.Date(2022, 10, 30, 0, 22, 0, 0, edt)}).String()
	if want := "2022-10-30T00:22:00-04:00"; want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}

package cli

import (
	"flag"
	"fmt"
	"time"
)

func TimeFlags() *timeflags {
	var out timeflags
	flag.Var(timeflag{&out.start}, "start", "start (inclusive)")
	flag.Var(timeflag{&out.end}, "end", "end (exclusive)")
	flag.DurationVar(&out.dur, "for", 0, "duration")
	return &out
}

type timeflags struct {
	start, end time.Time
	dur        time.Duration
}

func (f timeflags) Times() (start, end time.Time, err error) {
	var zero time.Time
	if f.start == zero {
		return zero, zero, fmt.Errorf("provide -start")
	}
	if (f.end == zero) == (f.dur == 0) {
		return zero, zero, fmt.Errorf("provide one of -end or -for")
	}
	if f.dur != 0 {
		return f.start, f.start.Add(f.dur), nil
	}
	return f.start, f.end, nil
}

type timeflag struct{ *time.Time }

func (f timeflag) Set(v string) error {
	out, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return err
	}
	*f.Time = out
	return nil
}

func (f timeflag) String() string {
	if f.Time == nil {
		return "<nil>"
	}
	return f.Format(time.RFC3339)
}

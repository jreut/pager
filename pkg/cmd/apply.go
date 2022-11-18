package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/jreut/pager/v2/pkg/interval"
	"github.com/jreut/pager/v2/pkg/save"
)

type Destination interface {
	// Apply writes the intervals to the schedule, overwriting whatever was there before.
	Apply(context.Context, string, []save.Interval) error
}

type FakeDestination struct{ Writer io.Writer }

func (d FakeDestination) Apply(_ context.Context, schedule string, in []save.Interval) error {
	_, err := fmt.Fprintf(d.Writer, "writing intervals for schedule %q\n", schedule)
	if err != nil {
		return err
	}
	for i, v := range in {
		_, err := fmt.Fprintf(d.Writer, "%d: %s\n", i, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func Apply(ctx context.Context, r io.Reader, dst Destination, schedule string) error {
	in, err := interval.ReadCSV(r, schedule, save.IntervalKindShift)
	if err != nil {
		return err
	}
	return dst.Apply(ctx, schedule, in)
}

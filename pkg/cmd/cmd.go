// Package cmd is the interesting part between command line parsing and the persistence layer.
package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/jreut/pager/v2/pkg/interval"
	"github.com/jreut/pager/v2/pkg/save"
)

var ErrConflict = errors.New("conflict")

func AddInterval(ctx context.Context, q *save.Queries, arg save.AddIntervalParams) error {
	var conflict string
	switch arg.Kind {
	case save.IntervalKindExclusion:
		conflict = save.IntervalKindShift
	case save.IntervalKindShift:
		conflict = save.IntervalKindExclusion
	default:
		return fmt.Errorf("unhandled kind %q", arg.Kind)
	}

	existing, err := q.ListIntervals(ctx, conflict)
	if err != nil {
		return err
	}

	if x, ok := interval.Conflict(existing, save.Interval(arg)); ok {
		return fmt.Errorf("%w: cannot schedule %s over existing %s", ErrConflict, save.Interval(arg), x)
	}

	return q.AddInterval(ctx, arg)
}

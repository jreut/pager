package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jreut/pager/v2/pkg/save"
)

type Action struct {
	Kind string
	Who  string
	At   time.Time
}

func EditSchedule(ctx context.Context, q *save.Queries, schedule string, actions []Action) error {
	for _, action := range actions {
		if err := q.AddEvent(ctx, save.AddEventParams{
			Person:   action.Who,
			Schedule: schedule,
			Kind:     action.Kind,
			At:       action.At,
		}); err != nil {
			return fmt.Errorf("adding participant %s(%s): %w", action.Kind, action.Who, err)
		}
	}
	return nil
}

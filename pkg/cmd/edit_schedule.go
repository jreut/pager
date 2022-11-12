package cmd

import (
	"context"
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
		if err := q.AddPerson(ctx, action.Who); err != nil {
			return err
		}
		if err := q.Participate(ctx, save.ParticipateParams{
			Person:   action.Who,
			Schedule: schedule,
			Kind:     action.Kind,
			At:       action.At,
		}); err != nil {
			return err
		}
	}
	return nil
}

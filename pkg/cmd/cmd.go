// Package cmd is the interesting part between command line parsing and the persistence layer.
package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/jreut/pager/v2/pkg/save"
)

var ErrConflict = errors.New("conflict")

type Action struct {
	Kind string
	Who  string
	At   time.Time
}

func Edit(ctx context.Context, q *save.Queries, schedule string, actions []Action) error {
	for _, action := range actions {
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

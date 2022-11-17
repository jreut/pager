package cmd

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jreut/pager/v2/pkg/global"
	"github.com/jreut/pager/v2/pkg/save"
)

var edt = func() *time.Location {
	edt, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	return edt
}()

const Style = "MondayAndFridayAtNoonEastern"

func Generate(ctx context.Context, q *save.Queries, schedule string, style string, start, end time.Time) error {
	switch style {
	case Style:
	default:
		return fmt.Errorf("unhandled style %q", style)
	}
	events, err := q.ListEvents(ctx, schedule)
	if err != nil {
		return err
	}
	available := make(map[string]bool)
	tally := make(map[string]time.Duration)
	a, b := start, nextShiftTransition(start)
	for a.Before(end) {
		if b.After(end) {
			b = end
		}
		for _, e := range events {
			if e.At.After(a) {
				break
			}
			switch e.Kind {
			case save.EventKindAdd:
				available[e.Person] = true
			case save.EventKindRemove:
				delete(available, e.Person)
			default:
				panic(fmt.Sprintf("unhandled event %q", e.Kind))
			}
		}
		var people []string
		for p := range available {
			people = append(people, p)
		}
		sort.Slice(people, func(i, j int) bool {
			if global.Deterministic() {
				if tally[people[i]] < tally[people[j]] {
					return true
				}
				if tally[people[i]] == tally[people[j]] {
					return people[i] < people[j]
				}
				return false
			}
			return tally[people[i]] < tally[people[j]]
		})
		if len(people) == 0 {
			return fmt.Errorf("nobody available to take shift from %s to %s", a, b)
		}
		person := people[0]
		tally[person] += b.Sub(a)
		params := save.AddIntervalParams{
			Person:    person,
			Schedule:  schedule,
			StartAt:   a,
			EndBefore: b,
			Kind:      save.IntervalKindShift,
		}
		// Bypass cmd.AddInterval's conflict checks.
		// We fix any exclusions for this person later in this function.
		if err := q.AddInterval(ctx, params); err != nil {
			return fmt.Errorf("inserting interval %+v: %w", params, err)
		}
		exclusions, err := q.ListIntervals(ctx, save.ListIntervalsParams{
			Schedule:  schedule,
			Kind:      save.IntervalKindExclusion,
			StartAt:   a,
			EndBefore: b,
		})
		if err != nil {
			return err
		}
		// Find someone to cover each exclusion this person has during this interval.
		for _, e := range exclusions {
			if e.Person == person {
				if len(people) < 2 {
					return fmt.Errorf("nobody available to cover %s from %s to %s", person, a, b)
				}
				params := save.AddIntervalParams{
					Person:    people[1],
					Schedule:  schedule,
					StartAt:   e.StartAt,
					EndBefore: e.EndBefore,
					Kind:      save.IntervalKindShift,
				}
				if err := AddInterval(ctx, q, params); err != nil {
					return fmt.Errorf("inserting interval %+v: %w", params, err)
				}
			}
		}
		a, b = b, nextShiftTransition(b)
	}

	return nil
}

func nextShiftTransition(t time.Time) time.Time {
	noon := time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, edt)
	switch t.In(edt).Weekday() {
	case time.Sunday:
		return noon.AddDate(0, 0, 1)
	case time.Monday:
		if !t.Before(noon) {
			return noon.AddDate(0, 0, 4)
		}
		return noon
	case time.Tuesday:
		return noon.AddDate(0, 0, 3)
	case time.Wednesday:
		return noon.AddDate(0, 0, 2)
	case time.Thursday:
		return noon.AddDate(0, 0, 1)
	case time.Friday:
		if !t.Before(noon) {
			return noon.AddDate(0, 0, 3)
		}
		return noon
	case time.Saturday:
		return noon.AddDate(0, 0, 2)
	default:
		panic(fmt.Sprintf("unhandled t.Weekday() %v", t.Weekday()))
	}
}

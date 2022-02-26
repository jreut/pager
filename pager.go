package pager

import (
	"fmt"
	"sort"
	"time"
)

var nyc *time.Location

func init() {
	var err error
	nyc, err = time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
}

type Person string

type exclusion struct {
	p Person
	t time.Time
	d time.Duration
}

type exclusions []exclusion

func (es exclusions) excluded(from, to time.Time) map[Person]struct{} {
	out := map[Person]struct{}{}
	for _, e := range es {
		if _, ok := out[e.p]; ok {
			continue
		}
		start, end := e.t, e.t.Add(e.d)
		if (from.Before(start) && to.After(start)) || (from.Before(end) && to.After(end)) {
			out[e.p] = struct{}{}
		}
	}
	return out
}

// i think this is the least information we need to encode a schedule.
type schedule []Handoff
type Handoff struct {
	Recipient Person
	At        time.Time
}

func (h Handoff) String() string {
	return fmt.Sprintf("%s@%s", h.Recipient, h.At.Format(time.RFC3339))
}

type balance map[Person]time.Duration

func (b balance) copy() balance {
	out := make(balance)
	for k, v := range b {
		out[k] = v
	}
	return out
}

func (b balance) next() []Person {
	var ks []Person
	for k := range b {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool {
		k_i, k_j := ks[i], ks[j]
		b_i, b_j := b[k_i], b[k_j]
		// when two people have the same balance, arbitraily sort lexicographically.
		if b_i == b_j {
			return k_i < k_j
		}
		return b_i < b_j
	})
	return ks
}

func newschedule(
	bal balance, exclusions exclusions, start time.Time, dur time.Duration,
) (schedule, balance) {
	bal = bal.copy()
	var out schedule
	is := intervals(start, dur, nextbreakpoint)
	for _, i := range is {
		excluded := exclusions.excluded(i.Start, i.End)
		var p Person
		for _, p = range bal.next() {
			if _, ok := excluded[p]; !ok {
				break
			}
		}
		bal[p] += i.Duration()
		out = append(out, Handoff{
			Recipient: p,
			At:        i.Start,
		})
	}
	return out, bal
}

func nextbreakpoint(after time.Time) time.Time {
	y, m, d := after.Date()
	switch after.Weekday() {
	case time.Sunday:
		return time.Date(y, m, d+1, 12, 0, 0, 0, nyc)
	case time.Monday:
		breakpoint := time.Date(y, m, d, 12, 0, 0, 0, nyc)
		if after.Before(breakpoint) {
			return breakpoint
		} else {
			return time.Date(y, m, d+4, 12, 0, 0, 0, nyc)
		}
	case time.Tuesday:
		return time.Date(y, m, d+3, 12, 0, 0, 0, nyc)
	case time.Wednesday:
		return time.Date(y, m, d+2, 12, 0, 0, 0, nyc)
	case time.Thursday:
		return time.Date(y, m, d+1, 12, 0, 0, 0, nyc)
	case time.Friday:
		breakpoint := time.Date(y, m, d, 12, 0, 0, 0, nyc)
		if after.Before(breakpoint) {
			return breakpoint
		} else {
			return time.Date(y, m, d+3, 12, 0, 0, 0, nyc)
		}
	case time.Saturday:
		return time.Date(y, m, d+2, 12, 0, 0, 0, nyc)
	default:
		panic(fmt.Sprintf("unhandled day of week: %s", after.Weekday()))
	}
}

type interval struct{ Start, End time.Time }

func intervals(start time.Time, d time.Duration, next func(time.Time) time.Time) []interval {
	var out []interval
	end := start.Add(d)
	for start.Before(end) {
		end := next(start)
		if end.Before(start) {
			panic(fmt.Sprintf("invariant violation: %s < %s", end, start))
		}
		out = append(out, interval{start, end})
		start = end
	}
	return out
}

func (i interval) Duration() time.Duration { return i.End.Sub(i.Start) }

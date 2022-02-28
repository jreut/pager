package schedule

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

// Interval represents a span of time starting at Time, inclusive, and
// ending at Time.Add(Duration), exclusive.
type Interval struct {
	time.Time
	time.Duration
}

// Build a list of intervals starting at the given time and covering at
// least the given duration.
//
// The next function is called with start to find the start of the next
// interval, and so on until next returns a time exceeding start.Add(d).
// The next function must monotonically increase. If not, this
// function may enter an infinite loop.
func intervals(start time.Time, d time.Duration, next func(time.Time) time.Time) []Interval {
	var out []Interval
	end := start.Add(d)
	for start.Before(end) {
		end := next(start)
		if end.Before(start) {
			panic(fmt.Sprintf("invariant violation: %s < %s", end, start))
		}
		out = append(out, Interval{start, end.Sub(start)})
		start = end
	}
	return out
}

// Contains tests whether the given time is a member of the Interval.
func (i Interval) Contains(t time.Time) bool {
	if i.Equal(t) {
		return true
	}
	if i.Before(t) && i.Add(i.Duration).After(t) {
		return true
	}
	return false
}

// Conjoint tests whether the intervals overlap.
func Conjoint(a, b Interval) bool {
	s_a, e_a := a.Time, a.Add(a.Duration)
	s_b, e_b := b.Time, b.Add(b.Duration)
	if s_a == e_b || s_b == e_a {
		return false
	}
	return a.Contains(s_b) || a.Contains(e_b) || b.Contains(s_a) || b.Contains(e_a)
}

type exclusion struct {
	p Person
	Interval
}

type exclusions []exclusion

func (es exclusions) excluded(i Interval) map[Person]struct{} {
	out := map[Person]struct{}{}
	for _, e := range es {
		if _, ok := out[e.p]; ok {
			continue
		}
		if Conjoint(e.Interval, i) {
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

type config struct {
	start    time.Time
	duration time.Duration
	balance
	exclusions
}

type result struct {
	Schedule schedule
	Balance  balance
}

func newschedule(cfg config) result {
	bal := cfg.balance.copy()
	var out schedule
	is := intervals(cfg.start, cfg.duration, nextbreakpoint)
	for _, i := range is {
		excluded := cfg.exclusions.excluded(i)
		var p Person
		for _, p = range bal.next() {
			if _, ok := excluded[p]; !ok {
				break
			}
		}
		bal[p] += i.Duration
		out = append(out, Handoff{
			Recipient: p,
			At:        i.Time,
		})
	}
	return result{out, bal}
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

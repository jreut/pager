package schedule

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/cockroachdb/errors"
)

var NYC *time.Location

func init() {
	var err error
	NYC, err = time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
}

// Schedule is a sequence of handoffs.
type Schedule []Handoff

func FromCSV(r io.Reader) (Schedule, error) {
	csv := csv.NewReader(r)
	csv.FieldsPerRecord = 2
	all, err := csv.ReadAll()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var out Schedule
	for _, r := range all {
		t, err := time.Parse(time.RFC3339, r[0])
		if err != nil {
			return nil, errors.Wrapf(err, "parsing time from %q", r)
		}
		out = append(out, Handoff{At: t, Recipient: Person(r[1])})
	}
	return out, err
}

func (s Schedule) WriteCSV(w io.Writer) error {
	out := csv.NewWriter(w)
	defer out.Flush()
	for _, h := range s {
		if err := out.Write([]string{
			h.At.Format(time.RFC3339),
			string(h.Recipient),
		}); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// Handoff represents the beginning of a shift. To know who long this
// shift is, one must know when the next handoff is.
type Handoff struct {
	// Recipient is the person assigned to this shift.
	Recipient Person
	// At is the start time of this shift.
	At time.Time
}

func (h Handoff) String() string {
	return fmt.Sprintf("%s@%s", h.Recipient, h.At.Format(time.RFC3339))
}

// Person is the identifier of someone who has shifts in a schedule.
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

// contains tests whether the given time is a member of the Interval.
func (i Interval) contains(t time.Time) bool {
	if i.Equal(t) {
		return true
	}
	if i.Before(t) && i.Add(i.Duration).After(t) {
		return true
	}
	return false
}

// conjoint tests whether the intervals overlap.
func conjoint(a, b Interval) bool {
	s_a, e_a := a.Time, a.Add(a.Duration)
	s_b, e_b := b.Time, b.Add(b.Duration)
	if s_a == e_b || s_b == e_a {
		return false
	}
	return a.contains(s_b) || a.contains(e_b) || b.contains(s_a) || b.contains(e_a)
}

// Exclusion is a pairing of a person to an interval during which they
// will not get a shift.
type Exclusion struct {
	Person
	Interval
}

// Exclude constructs an exclusion for the person over the given times.
func Exclude(p Person, start, end time.Time) Exclusion {
	return Exclusion{p, Interval{start, end.Sub(start)}}
}

// Balance is a record of relatively how much oncall time each person
// has accrued.
type Balance map[Person]time.Duration

func (b Balance) copy() Balance {
	out := make(Balance)
	for k, v := range b {
		out[k] = v
	}
	return out
}

// next sorts the people in this balance in order of ascending time
// accrued.
func (b Balance) next() []Person {
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

// Builder builds schedules.
type Builder struct {
	// Interval is the span this schedule will cover.
	Interval
	// Balance is the collection of people available to serve on this
	// schedule, and their starting balances.
	Balance
	// Exclusions is the list of intervals for which the relevant person
	// will definitely not be scheduled.
	Exclusions []Exclusion
	// Next is used to generate handoff times from the previous handoff
	// time.
	Next func(time.Time) time.Time
}

// Result is a generated schedule and the new balances accrued from
// following that new schedule.
type Result struct {
	Schedule Schedule
	Balance  Balance
}

// Build generates a schedule from the configured builder.
//
// We make an effort to approach an even balance for all people while
// respecting exclusions.
//
// The returned balance is a copy of the original balance, updated to
// reflect the shifts in the generated schedule.
func (b Builder) Build() Result {
	bal := b.Balance.copy()
	var out Schedule
	is := intervals(b.Interval.Time, b.Interval.Duration, b.Next)
	for _, i := range is {
		excluded := map[Person]struct{}{}
		for _, e := range b.Exclusions {
			if _, ok := excluded[e.Person]; ok {
				continue
			}
			if conjoint(e.Interval, i) {
				excluded[e.Person] = struct{}{}
			}
		}
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
	return Result{out, bal}
}

// MondayFridayShifts is the handoff schedule used by SRE at the time of
// writing.
func MondayFridayShifts(after time.Time) time.Time {
	y, m, d := after.In(NYC).Date()
	switch after.Weekday() {
	case time.Sunday:
		return time.Date(y, m, d+1, 12, 0, 0, 0, NYC)
	case time.Monday:
		breakpoint := time.Date(y, m, d, 12, 0, 0, 0, NYC)
		if after.Before(breakpoint) {
			return breakpoint
		} else {
			return time.Date(y, m, d+4, 12, 0, 0, 0, NYC)
		}
	case time.Tuesday:
		return time.Date(y, m, d+3, 12, 0, 0, 0, NYC)
	case time.Wednesday:
		return time.Date(y, m, d+2, 12, 0, 0, 0, NYC)
	case time.Thursday:
		return time.Date(y, m, d+1, 12, 0, 0, 0, NYC)
	case time.Friday:
		breakpoint := time.Date(y, m, d, 12, 0, 0, 0, NYC)
		if after.Before(breakpoint) {
			return breakpoint
		} else {
			return time.Date(y, m, d+3, 12, 0, 0, 0, NYC)
		}
	case time.Saturday:
		return time.Date(y, m, d+2, 12, 0, 0, 0, NYC)
	default:
		panic(fmt.Sprintf("unhandled day of week: %s", after.Weekday()))
	}
}

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
type shift struct {
	person Person
	// [begin, end)
	begin, end time.Time
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

// balance is 
type balance map[Person]time.Duration

func (b balance) next() Person {
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
	return ks[0]
}

func newschedule(bal balance, start time.Time, dur time.Duration) schedule {
	end := start.Add(dur)
	local := start.In(nyc)
	var out schedule
	recipient := bal.next()
	out = append(out, Handoff{
		Recipient: recipient,
		At:        local,
	})
	last := local

	for {
		local = nextbreakpoint(local)
		if local.Before(end) {
			between := local.Sub(last)
			bal[recipient] += between
			recipient = bal.next()
			out = append(out, Handoff{
				Recipient: recipient,
				At:        local,
			})
			last = local
		} else {
			break
		}
	}
	return out
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

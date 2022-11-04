// Package interval manipulates [save.Interval]s.
package interval

import (
	"fmt"

	"github.com/jreut/pager/v2/pkg/save"
)

// Flatten turns a list of "overrides" into a flat timeline.
//
// It treats the order of the list as the order in which intervals were added to the system.
// Intervals added later overwrite previous intervals.
// Flatten returns a chronologically sorted list.
func Flatten(in []save.Interval) []save.Interval {
	if len(in) < 2 {
		return in
	}

	var out []save.Interval
	for _, s := range in {
		validate(s)
		// Find the slice bounds of the relevant, existing intervals.
		l, r := bounds(out, s)
		// Make a slice to splice in.
		tosplice := combine(out[l:r], s)
		// Splice in the new slice.
		out = append(out[:l], append(tosplice, out[r:]...)...)
	}

	// Merge consecutive intervals for the same person.
	out = merge(out)

	return out
}

// bounds finds the smallest slice of xs that are relevant to y.
//
// A given x in xs is relevant if it overlaps with y.
// If no xs are relevant to y, the returned slice bounds will be equal, and they will represent the place to insert y chronologically.
//
// We assume xs is chronologically sorted.
func bounds(xs []save.Interval, y save.Interval) (int, int) {
	if len(xs) == 0 {
		return 0, 0
	}
	var l, r int
	for l = 0; l < len(xs); l++ {
		if xs[l].EndBefore.After(y.StartAt) {
			break
		}
	}
	for r = len(xs); r >= 0; r-- {
		if xs[r-1].StartAt.Before(y.EndBefore) {
			break
		}
	}
	return l, r
}

// combine adds y to xs as an override.
//
// Each x in xs is shortened or deleted as necessary to let y fit.
func combine(xs []save.Interval, y save.Interval) []save.Interval {
	if len(xs) == 0 {
		return []save.Interval{y}
	}
	var out []save.Interval
	l, r := xs[0], xs[len(xs)-1]
	if l.StartAt.Before(y.StartAt) {
		l.EndBefore = y.StartAt
		out = append(out, l)
	}
	out = append(out, y)
	if r.EndBefore.After(y.EndBefore) {
		r.StartAt = y.EndBefore
		out = append(out, r)
	}
	return out
}

// merge combines consecutive intervals for the same person.
func merge(xs []save.Interval) []save.Interval {
	var out []save.Interval
	for _, x := range xs {
		if len(out) > 0 {
			last := out[len(out)-1]
			if last.EndBefore.Equal(x.StartAt) && last.Person == x.Person {
				out[len(out)-1].EndBefore = x.EndBefore
				continue
			}
		}
		out = append(out, x)
	}
	return out
}

// Conflict tells which interval in xs conflicts with y.
//
// y overlaps with an x in xs if x and y have the same Person and if their times overlap.
func Conflict(xs []save.Interval, y save.Interval) (save.Interval, bool) {
	validate(y)
	xs = Flatten(xs)
	for _, x := range xs {
		if x.Person == y.Person && overlap(x, y) {
			return x, true
		}
	}
	return save.Interval{}, false
}

func overlap(a, b save.Interval) bool {
	if !a.EndBefore.After(b.StartAt) {
		return false
	}
	if !b.EndBefore.After(a.StartAt) {
		return false
	}
	return true
}

func validate(s save.Interval) {
	if !s.StartAt.Before(s.EndBefore) {
		panic(fmt.Sprintf("invalid interval %+v: %s >= %s", s, s.StartAt, s.EndBefore))
	}
}

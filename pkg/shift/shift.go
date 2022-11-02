package shift

import (
	"github.com/jreut/pager/v2/pkg/save"
)

// Flatten turns a list of "overrides" into a flat timeline.
//
// It treats the order of the list as the order in which shifts were added to the system.
// Shifts added later overwrite previous shifts.
// Flatten returns a chronologically sorted list.
func Flatten(in []save.Shift) []save.Shift {
	if len(in) < 2 {
		return in
	}

	var out []save.Shift
	for _, s := range in {
		if !s.StartAt.Before(s.EndBefore) {
			panic("invalid shift")
		}
		// Find the slice bounds of the relevant, existing shifts.
		l, r := bounds(out, s)
		// Make a slice to splice in.
		tosplice := combine(out[l:r], s)
		// Splice in the new slice.
		out = append(out[:l], append(tosplice, out[r:]...)...)
	}

	// Merge consecutive shifts for the same person.
	out = merge(out)

	return out
}

// bounds finds the smallest slice of xs that are relevant to y.
//
// A given x in xs is relevant if it overlaps with y.
// If no xs are relevant to y, the returned slice bounds will be equal, and they will represent the place to insert y chronologically.
//
// We assume xs is chronologically sorted.
func bounds(xs []save.Shift, y save.Shift) (int, int) {
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
func combine(xs []save.Shift, y save.Shift) []save.Shift {
	if len(xs) == 0 {
		return []save.Shift{y}
	}
	var out []save.Shift
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

// merge combines consecutive shifts for the same person.
func merge(xs []save.Shift) []save.Shift {
	var out []save.Shift
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

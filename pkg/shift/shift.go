package shift

import (
	"fmt"
	"sort"

	"github.com/jreut/pager/v2/pkg/save"
)

func Flatten(in []save.Shift) []save.Shift {
	if len(in) < 2 {
		return in
	}
	out := []save.Shift{in[0]}
	for _, i := range in[1:] {
		i_start, i_end := i.StartAt, i.EndBefore
		for j := range out {
			j_start, j_end := out[j].StartAt, out[j].EndBefore
			// j: alice: [-------)
			// i:   bob:   [---)
			// -------------------
			//    alice: [-)   [-)
			//      bob:   [---)
			if j_start.Before(i_start) && j_end.After(i_end) {
				left := save.Shift{Person: out[j].Person, StartAt: out[j].StartAt, EndBefore: i.StartAt}
				right := save.Shift{Person: out[j].Person, StartAt: i.EndBefore, EndBefore: out[j].EndBefore}
				out = append(append(out[:j], left, i, right), out[j+1:]...)
				break
			}
			// j:   bob:   [---)
			// i: alice: [-------)
			// -------------------
			//    alice: [-------)
			if !j_start.Before(i_start) && !j_end.After(i_end) {
				out = append(append(out[:j], i), out[j+1:]...)
				break
			}
		}
	}
	return out
}

func sortshifts(in []save.Shift) {
	sort.Slice(in, func(i, j int) bool {
		if overlap(in[i], in[j]) {
			return i < j
		}
		return in[i].StartAt.Before(in[j].StartAt)
	})
}

func overlap(a, b save.Shift) bool {
	a_start, a_end := a.StartAt, a.EndBefore
	b_start, b_end := b.StartAt, b.EndBefore
	if a_start.After(a_end) {
		panic(fmt.Sprintf("invalid shift: start > end: %+v", a))
	}
	if b_start.After(b_end) {
		panic(fmt.Sprintf("invalid shift: start > end: %+v", b))
	}
	if !a_end.After(b_start) {
		return false
	}
	if !b_end.After(a_start) {
		return false
	}
	return true
}

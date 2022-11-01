package shift

import (
	"sort"

	"github.com/jreut/pager/v2/pkg/save"
)

func Flatten(in []save.Shift) ([]save.Shift, error) {
	var out []save.Shift
	for i := range in {
		var in2 []save.Shift
		n := copy(in2, in[i:])
		if n == 0 {
			break
		}
		sort.Slice(in2, func(i, j int) bool {
			t_i, t_j := in2[i].StartAt, in2[j].StartAt
			switch {
			case t_i.Before(t_j):
				return true
			case t_i.Equal(t_j):
				return i < j
			default:
				return false
			}
		})
	}
	return out, nil
}

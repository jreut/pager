package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jreut/pager/schedule"
)

func TestGenerate(t *testing.T) {
	var w bytes.Buffer
	bal := strings.NewReader(`carp,0
darius,6h
jason,-96h
joel,12h
josh,-18h
logston,72h
reuter,-72h
`)
	exclusions := strings.NewReader(`2022-02-26T12:00:00-05:00,120h,reuter
`)
	want := `2022-02-20T00:00:00Z,jason
2022-02-20T12:00:00-05:00,jason
2022-02-21T12:00:00-05:00,reuter
2022-02-25T12:00:00-05:00,jason
2022-02-28T12:00:00-05:00,josh
2022-03-04T12:00:00-05:00,carp
2022-03-07T12:00:00-05:00,darius
2022-03-11T12:00:00-05:00,joel
2022-03-14T12:00:00-04:00,jason
2022-03-18T12:00:00-04:00,reuter
`
	err := generate(&w, schedule.Interval{
		Time:     time.Date(2022, 2, 20, 0, 0, 0, 0, time.UTC),
		Duration: 28 * 24 * time.Hour,
	}, bal, exclusions)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, w.String()); diff != "" {
		t.Fatalf("- want, + got:\n%s", diff)
	}
}

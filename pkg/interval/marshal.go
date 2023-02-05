package interval

import (
	"encoding/csv"
	"errors"
	"io"
	"time"

	"github.com/jreut/pager/v2/pkg/save"
)

func WriteCSV(w io.Writer, out []save.Interval) error {
	csv := csv.NewWriter(w)
	defer csv.Flush()
	if err := csv.Write([]string{"start_at", "end_before", "person"}); err != nil {
		return err
	}
	for _, i := range out {
		if err := csv.Write([]string{
			i.StartAt.Format(time.RFC3339),
			i.EndBefore.Format(time.RFC3339),
			i.Person,
		}); err != nil {
			return err
		}
	}
	return nil
}

func ReadCSV(r io.Reader, schedule, kind string) ([]save.Interval, error) {
	var out []save.Interval
	csv := csv.NewReader(r)
	csv.ReuseRecord = true
	csv.FieldsPerRecord = 3
	csv.Comment = '#'
	for {
		record, err := csv.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if record[0] == "start_at" {
			continue
		}
		start, err := time.Parse(time.RFC3339, record[0])
		if err != nil {
			return out, err
		}
		end, err := time.Parse(time.RFC3339, record[1])
		if err != nil {
			return out, err
		}
		out = append(out, save.Interval{
			Person:    record[2],
			Schedule:  schedule,
			StartAt:   start,
			EndBefore: end,
			Kind:      kind,
		})
	}
	return out, nil
}

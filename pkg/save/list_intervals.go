package save

import (
	"context"
	"time"
)

type ListIntervalsParams struct {
	Schedule  string
	Kind      string
	StartAt   time.Time
	EndBefore time.Time
}

func (q *Queries) ListIntervals(ctx context.Context, arg ListIntervalsParams) ([]Interval, error) {
	const sql = `
SELECT
  person
, schedule
, start_at
, end_before
, kind
FROM interval
WHERE schedule = ?
AND kind = ?
AND start_at < ?
AND end_before > ?
`
	rows, err := q.db.QueryContext(ctx, sql, arg.Schedule, arg.Kind, arg.EndBefore, arg.StartAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Interval
	for rows.Next() {
		var i Interval
		if err := rows.Scan(
			&i.Person,
			&i.Schedule,
			&i.StartAt,
			&i.EndBefore,
			&i.Kind,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

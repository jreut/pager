-- name: AddSchedule :exec
INSERT INTO schedule(name) VALUES (?);

-- name: AddInterval :exec
INSERT INTO interval(person, schedule, start_at, end_before, kind)
VALUES (?, ?, ?, ?, ?);

-- name: AddEvent :exec
INSERT INTO event(person, schedule, kind, at)
VALUES (?, ?, ?, ?);

-- name: ListEvents :many
SELECT * FROM event WHERE schedule = ? ORDER BY at ASC

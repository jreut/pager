-- name: AddPerson :exec
INSERT INTO person(handle) VALUES (?);

-- name: AddSchedule :exec
INSERT INTO schedule(name) VALUES (?);

-- name: AddInterval :exec
INSERT INTO interval(person, schedule, start_at, end_before, kind)
VALUES (?, ?, ?, ?, ?);

-- name: Participate :exec
INSERT INTO participate(person, schedule, kind, at)
VALUES (?, ?, ?, ?);

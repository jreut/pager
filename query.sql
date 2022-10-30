-- name: ListAllShifts :many
SELECT * FROM shift
WHERE datetime(start_at) < ?
AND datetime(end_before) > ?
;

-- name: ListShiftsForPerson :many
SELECT * FROM shift
WHERE datetime(start_at) < ?
AND datetime(end_before) > ?
AND person = ?
;

-- name: AddShift :exec
INSERT INTO shift(person, start_at, end_before)
VALUES (?,?,?);


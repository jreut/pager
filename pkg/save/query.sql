-- name: AddInterval :exec
INSERT INTO interval(person, start_at, end_before, kind)
VALUES (?,?,?,?);

-- name: ListIntervals :many
SELECT * FROM interval
WHERE kind = ?
-- TODO: sqlc doesn't seem to support nested WHERE clauses for sqlite.
-- AND (
-- 	start_at >= ?
-- 	OR end_before <= ?
-- )
;

-- name: AddPerson :exec
INSERT INTO person(handle) VALUES (?);

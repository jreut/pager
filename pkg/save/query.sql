-- name: AddShift :exec
INSERT INTO shift(person, start_at, end_before)
VALUES (?,?,?);

-- name: AddPerson :exec
INSERT INTO person(handle) VALUES (?);

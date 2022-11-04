-- name: AddInterval :exec
INSERT INTO interval(person, start_at, end_before, kind)
VALUES (?,?,?,?);

-- name: AddPerson :exec
INSERT INTO person(handle) VALUES (?);

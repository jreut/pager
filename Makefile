all: test

test: internal/save
	go test ./...

test/record:
	UPDATE_GOLDEN=1 go test ./...
	go test ./...

run: internal/save db.sqlite3
	go run .

internal/save: bin/sqlc schema.sql sqlc.yaml
	bin/sqlc generate

bin/sqlc:
	mkdir -p bin
	GOBIN=$(abspath bin) go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

db.sqlite3: schema.sql
	sqlite3 db.sqlite3 < schema.sql

all: test

test: pkg/save
	go test ./...

test/record: pkg/save
	sh test-record go test ./...

run: pkg/save db.sqlite3
	go run .

pkg/save: bin/sqlc sqlc.yaml pkg/save/schema.sql pkg/save/query.sql 
	bin/sqlc generate

bin/sqlc:
	mkdir -p bin
	GOBIN=$(abspath bin) go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

db.sqlite3: pkg/save/schema.sql
	sqlite3 db.sqlite3 < pkg/save/schema.sql

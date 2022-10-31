all: test

test: pkg/save
	find . -type d -name testdata | xargs rm -rv
	UPDATE_GOLDEN=1 go test ./...
	go test ./...

run: pkg/save db.sqlite3
	go run .

pkg/save: bin/sqlc schema.sql sqlc.yaml
	bin/sqlc generate

bin/sqlc:
	mkdir -p bin
	GOBIN=$(abspath bin) go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

db.sqlite3: schema.sql
	sqlite3 db.sqlite3 < schema.sql

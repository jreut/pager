src = $(shell find . -type f -name '*.go')

.PHONY: all
all: test

.PHONY: test
test: pkg/save
	go test ./...

.PHONY: test/record
test/record: pkg/save
	sh test-record go test ./...

.PHONY: fmt
fmt: $(src)
	gofmt -w $?

pkg/save/db.go: bin/sqlc sqlc.yaml pkg/save/schema.sql pkg/save/query.sql 
	bin/sqlc generate

bin/sqlc:
	mkdir -p bin
	GOBIN=$(abspath bin) go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

.PHONY: db.sqlite3
db.sqlite3: pkg/save/schema.sql
	sqlite3 db.sqlite3 < pkg/save/schema.sql

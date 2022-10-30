all: generate test run

test:
	go test ./...
run:
	go run .

generate: bin/sqlc
	bin/sqlc generate

bin/sqlc:
	GOBIN=$(abspath bin) go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

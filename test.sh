#!/bin/bash

out=$(mktemp)
trap 'rm $out' EXIT
go run . generate \
	-from=2022-02-20T00:00:00Z \
	-for=672h \
	-balance=testdata/sh/balance.csv \
	-exclusions=testdata/sh/exclusions.csv \
	> $out
diff -u $out testdata/sh/want.csv

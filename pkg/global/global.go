package global

import (
	"flag"
	"os"
)

// Deterministic tells whether the program should vary due to randomness or time.
//
// This value is usually false.
// Callers can set DETERMINISTIC=1 in the environment to make it true.
// Deterministic will always be true during `go test`.
func Deterministic() bool {
	return flag.Lookup("test.short") != nil || os.Getenv("DETERMINISTIC") == "1"
}

package global

import (
	"flag"
	"os"
)

func Deterministic() bool {
	return flag.Lookup("test.short") != nil || os.Getenv("DETERMINISTIC") == "1"
}

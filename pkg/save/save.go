// Package save is the persistence layer.
package save

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

const (
	IntervalKindShift     = "SHIFT"
	IntervalKindExclusion = "EXCLUSION"
)

//go:embed schema.sql
var Schema string

func Open(path string, opts url.Values) (*sql.DB, error) {
	if opts == nil {
		opts = make(url.Values)
	}
	opts.Set("_fk", "true")
	return sql.Open("sqlite3", path+"?"+opts.Encode())
}

func (i Interval) String() string {
	return fmt.Sprintf("%s for %q [%s, %s)", i.Kind, i.Person, i.StartAt, i.EndBefore)
}

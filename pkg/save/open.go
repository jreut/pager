package save

import (
	"database/sql"
	"net/url"

	_ "github.com/mattn/go-sqlite3"
)

func Open(path string, opts url.Values) (*sql.DB, error) {
	if opts == nil {
		opts = make(url.Values)
	}
	opts.Set("_fk", "true")
	return sql.Open("sqlite3", path+"?"+opts.Encode())
}

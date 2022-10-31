package main

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/jreut/pager/v2/internal/save"
)

func TestTimeflag(t *testing.T) {
	edt, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		arg  string
		want time.Time
	}{
		{
			arg:  "2022-10-30T00:06:00-04:00",
			want: time.Date(2022, 10, 30, 0, 6, 0, 0, edt),
		},
		{
			arg: "2022-10-30T00:06-04:00",
		},
		{
			arg: "2022-10-30T00:06:00EDT",
		},
	} {
		var val time.Time
		f := timeflag{&val}
		err := f.Set(tt.arg)
		var empty time.Time
		if tt.want != empty {
			if err != nil {
				t.Fatalf("want %s, got an error: %v", tt.want, err)
			}
			if !tt.want.Equal(*f.Time) {
				t.Fatalf("want %s == %s", tt.want, f)
			}
		} else {
			if err == nil {
				t.Fatalf("want an error, got %s", f.Time)
			}
		}
	}

	gott := time.Date(2022, 10, 30, 0, 22, 0, 0, edt)
	got := (timeflag{&gott}).String()
	if want := "2022-10-30T00:22:00-04:00"; want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestAddPerson(t *testing.T) {
	ctx := context.Background()
	q := save.New(testdb(t, ctx))
	err := q.AddPerson(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
}

func testdb(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, string(schema))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

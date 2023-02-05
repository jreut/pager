package main_test

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/save"
)

//go:embed fixtures/*
var fixtures embed.FS

func TestIntegration(t *testing.T) {
	ctx := context.Background()
	prog := filepath.Join(t.TempDir(), "prog")
	out, err := exec.
		CommandContext(ctx, "go", "build", "-o", prog, ".").
		CombinedOutput()
	if err != nil {
		t.Fatal(string(out), err)
	}

	type tc []struct {
		args   []string
		status int
	}

	for _, tt := range [][]struct {
		args   []string
		stdin  io.Reader
		status int
	}{
		{{
			args:   nil,
			status: 1,
		}},
		{{
			args:   []string{"unknown"},
			status: 1,
		}},
		{
			{
				args:   []string{"add-schedule", "-name", "default"},
				status: 0,
			},
			{
				args:   []string{"add-interval", "-schedule", "default", "-who", "alice", "-start", "2022-10-31T15:40:00-04:00", "-for", "24h"},
				status: 0,
			},
		},
		{
			{
				args:   []string{"add-schedule", "-name", "default"},
				status: 0,
			},
			{
				args:   []string{"add-interval", "-schedule", "default", "-who", "alice", "-start", "2022-10-31T15:40:00-04:00", "-for", "24h"},
				status: 0,
			},
			{
				args:   []string{"add-interval", "-schedule", "default", "-who", "alice", "-start", "2022-11-01T09:00:00-04:00", "-for", "1h", "-kind", "EXCLUSION"},
				status: 17,
			},
		},
		{
			{
				args:   []string{"add-schedule", "-name", "default"},
				status: 0,
			},
			{
				args:   []string{"add-interval", "-schedule", "default", "-who", "alice", "-start", "2022-11-01T00:00:00Z", "-for", "24h"},
				status: 0,
			},
			{
				args:   []string{"add-interval", "-schedule", "default", "-who", "bob", "-start", "2022-11-02T00:00:00Z", "-for", "24h"},
				status: 0,
			},
			{
				args:   []string{"show-schedule", "-schedule", "default", "-start", "2022-11-01T00:00:00Z", "-for", "48h"},
				status: 0,
			},
		},
		{
			{
				args:   []string{"add-schedule", "-name=default"},
				status: 0,
			},
			{
				args:   []string{"edit", "-schedule=default", "-add=bob=2023-01-01T00:00:00Z", "-add=alice=2023-01-01T00:00:00Z"},
				status: 0,
			},
			{
				args:   []string{"add-interval", "-schedule=default", "-who=alice", "-kind=EXCLUSION", "-start=2023-01-04T00:00:00Z", "-end=2023-01-09T00:00:00Z"},
				status: 0,
			},
			{
				args:   []string{"generate", "-schedule=default", "-start=2023-01-01T00:00:00Z", "-end=2023-02-01T00:00:00Z", "-style=MondayAndFridayAtNoonEastern"},
				status: 0,
			},
			{
				args:   []string{"show-schedule", "-schedule", "default", "-start=2023-01-01T00:00:00Z", "-end=2023-02-01T00:00:00Z"},
				status: 0,
			},
		},
		{
			{
				args:  []string{"apply", "-schedule=testschedule"},
				stdin: fixture(t, "fixtures/apply.in.csv"),
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			dbname := filepath.Join(t.TempDir(), "db.sqlite3")
			db, err := save.Open(dbname, nil)
			assert.Nil(t, err)
			defer db.Close()
			_, err = db.ExecContext(ctx, save.Schema)
			assert.Nil(t, err)

			stdout, stderr := assert.Golden(t, "out.txt"), assert.Golden(t, "err.txt")

			for _, ttt := range tt {
				t.Run("", func(t *testing.T) {
					cmd := exec.CommandContext(ctx, prog, ttt.args...)
					cmd.Env = append(os.Environ(),
						"DETERMINISTIC=1",
						"DB="+dbname,
					)
					fmt.Fprintf(io.MultiWriter(stdout, stderr), "# %s\n", strings.Join(cmd.Args[1:], " "))
					var buf bytes.Buffer
					cmd.Stdin = ttt.stdin
					cmd.Stdout = stdout
					cmd.Stderr = io.MultiWriter(&buf, stderr)
					t.Log(cmd)
					err = cmd.Run()
					t.Log(buf.String())
					if ttt.status == 0 {
						assert.Nil(t, err)
					} else {
						var exiterr *exec.ExitError
						ok := errors.As(err, &exiterr)
						if !ok {
							t.Fatalf("not an *exec.ExitError: %v", err)
						}
						assert.Cmp(t, ttt.status, exiterr.ExitCode())
					}
				})
			}
		})
	}
}

func fixture(t *testing.T, name string) io.Reader {
	f, err := fixtures.Open(name)
	assert.Nil(t, err)
	return f
}

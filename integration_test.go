package main_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/jreut/pager/v2/assert"
	"github.com/jreut/pager/v2/internal/save"
)

func TestIntegration(t *testing.T) {
	prefix := []string{"go", "run", "."}
	ctx := context.Background()
	for _, tt := range [][]struct {
		args   []string
		status int
	}{
		{{
			args:   nil,
			status: 1,
		}},
		{{
			args:   []string{"add-person", "-who", "alice"},
			status: 0,
		}},
		{
			{
				args:   []string{"add-person", "-who", "alice"},
				status: 0,
			},
			{
				args:   []string{"add-shift", "-who", "alice", "-start", "2022-10-31T15:40:00.0-04:00", "-for", "24h"},
				status: 0,
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			f, err := os.CreateTemp("", "db-*.sqlite3")
			assert.Nil(t, err)
			assert.Nil(t, f.Close())
			t.Log(f.Name())
			db, err := save.Open(f.Name(), nil)
			assert.Nil(t, err)
			defer db.Close()
			schema, err := os.ReadFile("schema.sql")
			assert.Nil(t, err)
			_, err = db.ExecContext(ctx, string(schema))
			assert.Nil(t, err)

			for _, ttt := range tt {
				t.Run("", func(t *testing.T) {
					args := append(prefix, ttt.args...)
					cmd := exec.CommandContext(ctx, args[0], args[1:]...)
					cmd.Env = append(os.Environ(),
						"DETERMINISTIC=1",
						"DB="+f.Name(),
					)
					var stderr bytes.Buffer
					cmd.Stdout = assert.Golden(t, "out.txt")
					cmd.Stderr = io.MultiWriter(&stderr, assert.Golden(t, "err.txt"))
					t.Log(cmd)
					err = cmd.Run()
					t.Log(stderr.String())
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

package main_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/save"
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
			dbname := filepath.Join(t.TempDir(), "db.sqlite3")
			db, err := save.Open(dbname, nil)
			assert.Nil(t, err)
			defer db.Close()
			_, err = db.ExecContext(ctx, save.Schema)
			assert.Nil(t, err)

			for _, ttt := range tt {
				t.Run("", func(t *testing.T) {
					args := append(prefix, ttt.args...)
					cmd := exec.CommandContext(ctx, args[0], args[1:]...)
					cmd.Env = append(os.Environ(),
						"DETERMINISTIC=1",
						"DB="+dbname,
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

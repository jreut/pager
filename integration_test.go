package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/jreut/pager/v2/assert"
)

func TestIntegration(t *testing.T) {
	prefix := []string{"go", "run", "."}
	ctx := context.Background()
	for _, tt := range []struct {
		args   []string
		status int
	}{
		{
			args:   nil,
			status: 1,
		},
		{
			args:   []string{"add-person", "-who", "alice"},
			status: 0,
		},
	} {
		t.Run("", func(t *testing.T) {
			f, err := os.CreateTemp("", "db-*.sqlite3")
			assert.Nil(t, err)
			assert.Nil(t, f.Close())
			t.Log(f.Name())
			db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_fk=true", f.Name()))
			assert.Nil(t, err)
			defer db.Close()
			schema, err := os.ReadFile("schema.sql")
			assert.Nil(t, err)
			_, err = db.ExecContext(ctx, string(schema))
			assert.Nil(t, err)

			args := append(prefix, tt.args...)
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
			if tt.status == 0 {
				assert.Nil(t, err)
			} else {
				var exiterr *exec.ExitError
				ok := errors.As(err, &exiterr)
				if !ok {
					t.Fatalf("not an *exec.ExitError: %v", err)
				}
				assert.Cmp(t, tt.status, exiterr.ExitCode())
			}
		})
	}
}

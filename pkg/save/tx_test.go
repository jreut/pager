package save_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/save"
)

func TestTx(t *testing.T) {
	ctx := context.Background()
	db := testdb(t, ctx)

	assertrows := func(t *testing.T, want []int) {
		t.Helper()
		var got []int
		rows, err := db.QueryContext(ctx, `SELECT rowid FROM t`)
		assert.Nil(t, err)
		defer rows.Close()
		for rows.Next() {
			var id int
			assert.Nil(t, rows.Scan(&id))
			got = append(got, id)
		}
		assert.Cmp(t, want, got)
	}

	_, err := db.ExecContext(ctx, `CREATE TABLE t(a)`)
	assert.Nil(t, err)
	assertrows(t, nil)

	_, err = db.ExecContext(ctx, `INSERT INTO t(rowid) VALUES (1);`)
	assert.Nil(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO t(rowid) VALUES (1);`)
	assert.Error(t, "UNIQUE constraint failed", err)
	assertrows(t, []int{1})

	assert.Error(t, "UNIQUE constraint failed", save.WithTx(db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO t(rowid) VALUES (2);`)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `INSERT INTO t(rowid) VALUES (2);`)
		return err
	}))
	assertrows(t, []int{1})

	var myerr = errors.New("myerr")
	err = save.WithTx(db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO t(rowid) VALUES (3);`)
		if err != nil {
			return err
		}
		return myerr
	})
	assert.Error(t, myerr.Error(), err)
	assertrows(t, []int{1})

	err = save.WithTx(db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO t(rowid) VALUES (4);`)
		return err
	})
	assert.Nil(t, err)
	assertrows(t, []int{1, 4})
}

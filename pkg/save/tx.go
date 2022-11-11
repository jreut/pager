package save

import (
	"database/sql"
	"fmt"
)

func WithTx(db *sql.DB, f func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err := f(tx); err != nil {
		if err2 := tx.Rollback(); err2 != nil {
			return fmt.Errorf("%w: rollback error while handling %v", err2, err)
		}
		return err
	}
	return tx.Commit()
}

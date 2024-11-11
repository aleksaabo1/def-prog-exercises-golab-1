package safesql

import (
	"context"
	"database/sql"
)

type DB struct {
	db *sql.DB
}

func (db *DB) QueryContext(ctx context.Context, query TrustedSQL, args ...any) (*Rows, error) {
	r, err := db.db.QueryContext(ctx, query.s, args...)
	return r, err
}

type Rows = sql.Rows

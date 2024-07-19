package db

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB holds the database connection pool.
type DB struct {
	config *pgxpool.Config
	pool   *pgxpool.Pool
}

// Option is a function that configures a DB.
type Option func(*DB)

// NewDB creates a new database connection pool.
func NewDB(ctx context.Context, connString string, opts ...Option) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	db := &DB{config: poolConfig}
	for _, opt := range opts {
		opt(db)
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}
	db.pool = pool
	return db, nil
}

// Close closes the database connection pool.
func (d *DB) Close() {
	d.pool.Close()
}

// Ping verifies a connection to the database is still alive.
func (d *DB) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

// SelectOne executes a query that is expected to return at most one row.
func (d *DB) SelectOne(ctx context.Context, result any, query string, args ...interface{}) error {
	row, err := d.pool.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	return pgxscan.ScanOne(result, row)
}

// Select executes a query that returns rows.
func (d *DB) Select(ctx context.Context, result any, query string, args ...interface{}) error {
	rows, err := d.pool.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	return pgxscan.ScanAll(result, rows)
}

// Exec executes a query without returning any rows.
// The pgconn.CommandTag structure returns the operation that is executed in the database.
// There is a bug when doing an INSERT..ON CONFLIC..DO UPDATE, if the update part of the query is executed
// the commandTag still returns that an INSERT was performed and not an UPDATE. If you need to be precise
// about what operation was executed, use the ExecAndScan method and return a field that is only updated.
func (d *DB) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return d.pool.Exec(ctx, query, args...)
}

// ExecAndScan executes a query and scans the result into the provided struct.
func (d *DB) ExecAndScan(ctx context.Context, result any, query string, args ...interface{}) error {
	return d.pool.QueryRow(ctx, query, args...).Scan(result)
}

// BeginTx starts a new transaction.
func (d *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return d.pool.Begin(ctx)
}

// Commit commits the transaction.
func (d *DB) Commit(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}

// Rollback rolls back the transaction.
func (d *DB) Rollback(ctx context.Context, tx pgx.Tx) error {
	return tx.Rollback(ctx)
}

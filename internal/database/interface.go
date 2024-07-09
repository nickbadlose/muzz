package database

import (
	"context"
	"database/sql"

	"github.com/upper/db/v4"
)

// Writer interface which only allows write operations.
type Writer interface {
	Preparer

	// InsertInto prepares and returns an Inserter targeted at the given table.
	InsertInto(string) db.Inserter

	// Update prepares and returns an Updater targeted at the given table.
	Update(string) db.Updater
}

// Preparer generic operations shared between reading and writing.
type Preparer interface {
	// ExecContext executes a SQL query that does not return any rows, like sql.ExecContext.
	ExecContext(ctx context.Context, query interface{}, args ...interface{}) (sql.Result, error)

	// PrepareContext creates a prepared statement on the given context for later
	// queries or executions. The caller must call the statement's Close method
	// when the statement is no longer needed.
	PrepareContext(ctx context.Context, query interface{}) (*sql.Stmt, error)
}

// Reader interface which only allows read operations.
type Reader interface {
	Preparer

	// Select initialises and returns a Selector, it accepts column names as
	// parameters.
	//
	// The returned Selector does not initially point to any table, a call to
	// From() is required after Select() to complete a valid query.
	Select(columns ...interface{}) db.Selector

	// SelectFrom creates a Selector that selects all columns (like SELECT *)
	// from the given table.
	SelectFrom(table ...interface{}) db.Selector

	// QueryContext executes a SQL query that returns rows, like
	// sql.QueryContext.
	QueryContext(ctx context.Context, query interface{}, args ...interface{}) (*sql.Rows, error)

	// QueryRowContext executes a SQL query that returns one row, like
	// sql.QueryRowContext.
	QueryRowContext(ctx context.Context, query interface{}, args ...interface{}) (*sql.Row, error)
}

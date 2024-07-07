package database

import (
	"context"
	"database/sql"

	"github.com/upper/db/v4"
)

// type aliases allow us to loosely decouple from the upper/db lib.
type (
	// Selector represents a SELECT statement.
	Selector = db.Selector
	// Updater represents an UPDATE statement.
	Updater = db.Updater
	// Inserter represents an INSERT statement.
	Inserter = db.Inserter
	// Deleter represents a DELETE statement.
	Deleter = db.Deleter
	// Iterator provides methods for iterating over query results.
	Iterator = db.Iterator
	// SQL defines methods that can be used to build a SQL query with chainable
	// method calls.
	//
	// Queries are immutable, so every call to any method will return a new
	// pointer, if you want to build a query using variables you need to reassign
	// them, like this:
	//
	//	a = builder.Select("name").From("foo") // "a" is created
	//
	//	a.Where(...) // No effect, the value returned from Where is ignored.
	//
	//	a = a.Where(...) // "a" is reassigned and points to a different address.
	SQL = db.SQL
)

type Client interface {
	SQL() db.SQL
	TxContext(ctx context.Context, fn func(tx Client) error, opts *sql.TxOptions) error
	WithContext(ctx context.Context) Client
	Ping() error
	Close() error
}

// Writer interface which only allows write operations.
type Writer interface {
	Preparer

	// InsertInto prepares and returns an Inserter targeted at the given table.
	InsertInto(string) Inserter

	// Update prepares and returns an Updater targeted at the given table.
	Update(string) Updater
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
	Select(columns ...interface{}) Selector

	// SelectFrom creates a Selector that selects all columns (like SELECT *)
	// from the given table.
	SelectFrom(table ...interface{}) Selector

	// QueryContext executes a SQL query that returns rows, like
	// sql.QueryContext.
	QueryContext(ctx context.Context, query interface{}, args ...interface{}) (*sql.Rows, error)

	// QueryRowContext executes a SQL query that returns one row, like
	// sql.QueryRowContext.
	QueryRowContext(ctx context.Context, query interface{}, args ...interface{}) (*sql.Row, error)
}

package database

import (
	"context"
	"database/sql"

	"github.com/upper/db/v4"
)

// type aliases allow us to loosely decouple from the upper/db lib
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
)

// TODO docs

type Database interface {
	ReadSessionContext(ctx context.Context) (Reader, error)
	WriteSessionContext(ctx context.Context) (Writer, error)
	TxContext(ctx context.Context, fn TransactionFunc, opts *sql.TxOptions) error
	Ping() error
	Close() error
}

type Client interface {
	SQL() SQL
	TxContext(ctx context.Context, fn TransactionFunc, opts *sql.TxOptions) error
	WithContext(ctx context.Context) Client
	Ping() error
	Close() error
}

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
type SQL interface {

	// Select initializes and returns a Selector, it accepts column names as
	// parameters.
	//
	// The returned Selector does not initially point to any table, a call to
	// From() is required after Select() to complete a valid query.
	//
	// Example:
	//
	//  q := sqlbuilder.Select("first_name", "last_name").From("people").Where(...)
	Select(columns ...interface{}) Selector

	// SelectFrom creates a Selector that selects all columns (like SELECT *)
	// from the given table.
	//
	// Example:
	//
	//  q := sqlbuilder.SelectFrom("people").Where(...)
	SelectFrom(table ...interface{}) Selector

	// InsertInto prepares and returns an Inserter targeted at the given table.
	//
	// Example:
	//
	//   q := sqlbuilder.InsertInto("books").Columns(...).Values(...)
	InsertInto(table string) Inserter

	// DeleteFrom prepares a Deleter targeted at the given table.
	//
	// Example:
	//
	//  q := sqlbuilder.DeleteFrom("tasks").Where(...)
	DeleteFrom(table string) Deleter

	// Update prepares and returns an Updater targeted at the given table.
	//
	// Example:
	//
	//  q := sqlbuilder.Update("profile").Set(...).Where(...)
	Update(table string) Updater

	// Exec executes a SQL query that does not return any rows, like sql.Exec.
	//
	// Example:
	//
	//  sqlbuilder.Exec(`INSERT INTO books (title) VALUES("La Ciudad y los Perros")`)
	Exec(query interface{}, args ...interface{}) (sql.Result, error)

	// ExecContext executes a SQL query that does not return any rows, like sql.ExecContext.
	//
	// Example:
	//
	//  sqlbuilder.ExecContext(ctx, `INSERT INTO books (title) VALUES(?)`, "La Ciudad y los Perros")
	ExecContext(ctx context.Context, query interface{}, args ...interface{}) (sql.Result, error)

	// Prepare creates a prepared statement for later queries or executions. The
	// caller must call the statement's Close method when the statement is no
	// longer needed.
	Prepare(query interface{}) (*sql.Stmt, error)

	// PrepareContext creates a prepared statement on the guiven context for later
	// queries or executions. The caller must call the statement's Close method
	// when the statement is no longer needed.
	PrepareContext(ctx context.Context, query interface{}) (*sql.Stmt, error)

	// Query executes a SQL query that returns rows, like sql.Query.
	//
	// Example:
	//
	//  sqlbuilder.Query(`SELECT * FROM people WHERE name = "Mateo"`)
	Query(query interface{}, args ...interface{}) (*sql.Rows, error)

	// QueryContext executes a SQL query that returns rows, like
	// sql.QueryContext.
	//
	// Example:
	//
	//  sqlbuilder.QueryContext(ctx, `SELECT * FROM people WHERE name = ?`, "Mateo")
	QueryContext(ctx context.Context, query interface{}, args ...interface{}) (*sql.Rows, error)

	// QueryRow executes a SQL query that returns one row, like sql.QueryRow.
	//
	// Example:
	//
	//  sqlbuilder.QueryRow(`SELECT * FROM people WHERE name = "Haruki" AND last_name = "Murakami" LIMIT 1`)
	QueryRow(query interface{}, args ...interface{}) (*sql.Row, error)

	// QueryRowContext executes a SQL query that returns one row, like
	// sql.QueryRowContext.
	//
	// Example:
	//
	//  sqlbuilder.QueryRowContext(ctx, `SELECT * FROM people WHERE name = "Haruki" AND last_name = "Murakami" LIMIT 1`)
	QueryRowContext(ctx context.Context, query interface{}, args ...interface{}) (*sql.Row, error)

	// Iterator executes a SQL query that returns rows and creates an Iterator
	// with it.
	//
	// Example:
	//
	//  sqlbuilder.Iterator(`SELECT * FROM people WHERE name LIKE "M%"`)
	Iterator(query interface{}, args ...interface{}) Iterator

	// IteratorContext executes a SQL query that returns rows and creates an Iterator
	// with it.
	//
	// Example:
	//
	//  sqlbuilder.IteratorContext(ctx, `SELECT * FROM people WHERE name LIKE "M%"`)
	IteratorContext(ctx context.Context, query interface{}, args ...interface{}) Iterator

	// NewIterator converts a *sql.Rows value into an Iterator.
	NewIterator(rows *sql.Rows) Iterator

	// NewIteratorContext converts a *sql.Rows value into an Iterator.
	NewIteratorContext(ctx context.Context, rows *sql.Rows) Iterator
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

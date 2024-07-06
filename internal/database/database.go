package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/upper/db/v4/adapter/postgresql"
)

const (
	// databaseDriver the name of the database driver to use.
	databaseDriver = "postgres"

	// db.Session configurations
	maxIdleConnections = 2
	maxOpenConnections = 5
	maxConnLifetime    = 30 * time.Minute
)

// errNoConnection error returned when there is no connection initialised.
var errNoConnection = errors.New(`no database connection`)

// Config the set of configurations required for a database connection.
type Config struct {
	// Username the username of the user to authenticate with.
	Username string
	// Password the password of the user to authenticate with.
	Password string
	// Name the database name.
	Name string
	// Host the URL / host of the database.
	Host string
	// DebugEnabled whether debug settings should be configured.
	DebugEnabled bool
}

// Database creates sessions for the client to interact with.
type Database struct{ client Client }

// ReadSessionContext returns the current database read session.
func (d Database) ReadSessionContext(ctx context.Context) (Reader, error) {
	if d.client == nil {
		return nil, errNoConnection
	}
	return d.client.WithContext(ctx).SQL(), nil
}

// WriteSessionContext returns the current write database session.
func (d Database) WriteSessionContext(ctx context.Context) (Writer, error) {
	if d.client == nil {
		return nil, errNoConnection
	}
	return d.client.WithContext(ctx).SQL(), nil
}

// TransactionFunc wraps a database transaction.
type TransactionFunc = func(tx Client) error

// TxContext performs the provided TransactionFunc against the database.
func (d Database) TxContext(ctx context.Context, fn func(tx Client) error, opts *sql.TxOptions) error {
	if d.client == nil {
		return errNoConnection
	}
	return d.client.TxContext(ctx, fn, opts)
}

// New instantiates a new Database and opens a connection via the defaultClientFunc.
func New(ctx context.Context, p *Config) (*Database, error) {
	return NewWithClientFunc(ctx, p, defaultClientFunc)
}

// NewWithClientFunc instantiates a new Database and opens a connection via the provided ClientFunc.
func NewWithClientFunc(ctx context.Context, p *Config, fn ClientFunc) (*Database, error) {
	client, err := fn(ctx, p)

	return &Database{client: client}, err
}

// ClientFunc is a function which can open a connection to the database.
type ClientFunc func(ctx context.Context, p *Config) (Client, error)

var defaultClientFunc ClientFunc = func(ctx context.Context, p *Config) (Client, error) {
	db, err := sql.Open(
		databaseDriver,
		fmt.Sprintf("%s://%s:%s@%s/%s?sslmode=disable", databaseDriver, p.Username, p.Password, p.Host, p.Name),
	)
	if err != nil {
		return nil, err
	}

	// bind this to upper's sqlbuilder.
	conn, err := postgresql.New(db)
	if err != nil {
		return nil, err
	}

	conn = conn.WithContext(ctx)
	conn.SetMaxIdleConns(maxIdleConnections)
	conn.SetMaxOpenConns(maxOpenConnections)
	conn.SetConnMaxLifetime(maxConnLifetime)

	return wrapSession(conn), nil
}

// Ping pings the current connection to the database.
func (d Database) Ping() error {
	if d.client == nil {
		return errNoConnection
	}
	return d.client.Ping()
}

// Close closes the current connection to the database.
func (d Database) Close() error {
	if d.client == nil {
		return nil
	}
	return d.client.Close()
}

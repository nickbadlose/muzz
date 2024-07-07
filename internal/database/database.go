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

var (
	// errNoConnection error returned when there is no connection initialised.
	errNoConnection = errors.New(`no database connection`)
)

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

// TxContext performs the provided transaction func against the database.
func (d Database) TxContext(ctx context.Context, fn func(tx Client) error, opts *sql.TxOptions) error {
	if d.client == nil {
		return errNoConnection
	}
	return d.client.TxContext(ctx, fn, opts)
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

// New instantiates a new Database and opens a connection via the defaultClientFunc.
func New(ctx context.Context, cfg *Config, fn ...func(context.Context, *Config) (Client, error)) (*Database, error) {
	client, err := openConnection(ctx, cfg, fn...)

	return &Database{client: client}, err
}

// openConnection opens a new connection to the database using the supplied
// username u and password p.
func openConnection(ctx context.Context, cfg *Config, fn ...func(context.Context, *Config) (Client, error)) (Client, error) {
	var dft = defaultClientFunc
	if len(fn) > 0 {
		dft = fn[0]
	}

	return dft(ctx, cfg)
}

var defaultClientFunc = func(ctx context.Context, cfg *Config) (Client, error) {
	db, err := sql.Open(
		databaseDriver,
		fmt.Sprintf("%s://%s:%s@%s/%s?sslmode=disable", databaseDriver, cfg.Username, cfg.Password, cfg.Host, cfg.Name),
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

	return WrapSession(conn), nil
}

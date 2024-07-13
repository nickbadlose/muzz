package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/postgresql"
	"go.nhat.io/otelsql"
	"go.opentelemetry.io/otel"
)

const (
	// databaseDriver the name of the database driver to use.
	databaseDriver = "postgres"

	// db.Session configurations.
	maxIdleConnections = 2
	maxOpenConnections = 5
	maxConnLifetime    = 30 * time.Minute
)

var (
	// errNoConnection error returned when there is no connection initialised.
	errNoConnection = errors.New(`no database connection`)
)

// Credentials for connecting and authenticating the database connection.
type Credentials struct {
	// Username the username of the user to authenticate with.
	Username string
	// Password the password of the user to authenticate with.
	Password string
	// Name the database name.
	Name string
	// Host the URL / host of the database.
	Host string
}

// Database creates sessions for the client to interact with.
type Database struct{ client db.Session }

// SQLSessionContext returns the current database SQL session.
func (d Database) SQLSessionContext(ctx context.Context) (db.SQL, error) {
	if d.client == nil {
		return nil, errNoConnection
	}
	return d.client.WithContext(ctx).SQL(), nil
}

// TxContext performs the provided transaction func against the database.
func (d Database) TxContext(ctx context.Context, fn func(tx db.Session) error, opts *sql.TxOptions) error {
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
func New(ctx context.Context, c *Credentials, opts ...Option) (*Database, error) {
	if c == nil {
		return nil, errors.New("credentials must be provided")
	}

	cfg := &Config{
		clientFunc:  defaultClientFunc,
		credentials: c,

		MaxIdleConns:    maxIdleConnections,
		MaxOpenConns:    maxOpenConnections,
		ConnMaxLifetime: maxConnLifetime,
		DebugEnabled:    false,
		// returns the global tracer provider or a noop if none is set.
		TracerProvider: otel.GetTracerProvider(),
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	client, err := cfg.clientFunc(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Database{client: client}, nil
}

// defaultClientFunc configures a db.Session with the given configurations and decorates it with trace information.
var defaultClientFunc = func(ctx context.Context, cfg *Config) (db.Session, error) {
	queryOpt := otelsql.TraceQueryWithoutArgs()
	if cfg.DebugEnabled {
		queryOpt = otelsql.TraceQueryWithArgs()
	}

	// register the trace provider with psql.
	driverName, err := otelsql.Register(
		databaseDriver,
		otelsql.WithTracerProvider(cfg.TracerProvider),
		otelsql.WithDatabaseName(cfg.credentials.Name),
		otelsql.TracePing(),
		otelsql.TraceRowsClose(),
		otelsql.TraceRowsNext(),
		queryOpt,
	)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(
		driverName,
		fmt.Sprintf(
			"%s://%s:%s@%s/%s?sslmode=disable",
			databaseDriver,
			cfg.credentials.Username,
			cfg.credentials.Password,
			cfg.credentials.Host,
			cfg.credentials.Name,
		),
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
	conn.SetMaxIdleConns(cfg.MaxIdleConns)
	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return conn, nil
}

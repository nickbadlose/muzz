package database

import (
	"context"
	"database/sql"

	"github.com/upper/db/v4"
)

// wrapSession generates a Client instance from an upper db.Session instance. This allows us to restrict which methods
// are available in transactions.
func wrapSession(database db.Session) Client { return &wrappedSession{database} }

type wrappedSession struct {
	db.Session
}

func (w *wrappedSession) WithContext(ctx context.Context) Client {
	return &wrappedSession{w.Session.WithContext(ctx)}
}

func (w *wrappedSession) SQL() SQL {
	return w.Session.SQL()
}

func (w *wrappedSession) TxContext(ctx context.Context, fn TransactionFunc, opts *sql.TxOptions) error {
	return w.Session.TxContext(ctx, func(sess db.Session) error {
		return fn(&wrappedSession{sess})
	}, opts)
}

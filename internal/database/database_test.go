package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	mockdatabase "github.com/nickbadlose/muzz/internal/database/mocks"
	"github.com/stretchr/testify/require"
	"github.com/upper/db/v4"
)

func TestNew(t *testing.T) {
	t.Run("database client should not be nil", func(t *testing.T) {
		dbClient, mockSQL, err := mockdatabase.NewWrappedMock()
		require.NoError(t, err)

		dbase, err := New(
			context.Background(),
			&Credentials{},
			WithClientFunc(func(_ context.Context, _ *Config) (db.Session, error) {
				return dbClient, err
			}),
		)
		require.NoError(t, err)
		require.NotNil(t, dbase)

		sess, err := dbase.SQLSessionContext(context.TODO())
		require.NoError(t, err)
		require.NotNil(t, sess)

		mockSQL.ExpectBegin()
		mockSQL.ExpectCommit()
		err = dbase.TxContext(context.TODO(), func(tx db.Session) error {
			require.NotNil(t, tx)
			return nil
		}, &sql.TxOptions{})
		require.NoError(t, err)

		require.NoError(t, dbase.Ping())

		mockSQL.ExpectClose()
		require.NoError(t, dbase.Close())

		require.NoError(t, mockSQL.ExpectationsWereMet())
	})

	t.Run("default options should be configured", func(t *testing.T) {
		dbase, err := New(
			context.Background(),
			&Credentials{
				Username: "test",
				Password: "test",
				Name:     "test",
				Host:     "test",
			},
			WithClientFunc(func(_ context.Context, cfg *Config) (db.Session, error) {
				require.NotNil(t, cfg.clientFunc)
				require.NotNil(t, cfg.credentials)
				require.Equal(t, "test", cfg.credentials.Username)
				require.Equal(t, "test", cfg.credentials.Password)
				require.Equal(t, "test", cfg.credentials.Name)
				require.Equal(t, "test", cfg.credentials.Host)
				require.Equal(t, maxIdleConnections, cfg.MaxIdleConns)
				require.Equal(t, maxOpenConnections, cfg.MaxOpenConns)
				require.Equal(t, maxConnLifetime, cfg.ConnMaxLifetime)
				require.False(t, cfg.DebugEnabled)
				require.NotNil(t, cfg.TracerProvider)
				return nil, nil
			}),
		)
		require.NoError(t, err)
		require.NotNil(t, dbase)
	})

	t.Run("options should configure database correctly", func(t *testing.T) {
		dbase, err := New(
			context.Background(),
			&Credentials{
				Username: "test",
				Password: "test",
				Name:     "test",
				Host:     "test",
			},
			WithClientFunc(func(_ context.Context, cfg *Config) (db.Session, error) {
				require.NotNil(t, cfg.clientFunc)
				require.NotNil(t, cfg.credentials)
				require.Equal(t, "test", cfg.credentials.Username)
				require.Equal(t, "test", cfg.credentials.Password)
				require.Equal(t, "test", cfg.credentials.Name)
				require.Equal(t, "test", cfg.credentials.Host)
				require.Equal(t, 1, cfg.MaxIdleConns)
				require.Equal(t, 1, cfg.MaxOpenConns)
				require.Equal(t, 1*time.Minute, cfg.ConnMaxLifetime)
				require.True(t, cfg.DebugEnabled)
				require.Nil(t, cfg.TracerProvider)
				return nil, nil
			}),
			WithDebugMode(true),
			WithConnMaxLifetime(1*time.Minute),
			WithMaxIdleConns(1),
			WithMaxOpenConns(1),
			WithTraceProvider(nil),
		)
		require.NoError(t, err)
		require.NotNil(t, dbase)
	})

	t.Run("fail, error building client", func(t *testing.T) {
		dbase, err := New(
			context.Background(),
			&Credentials{},
			WithClientFunc(func(_ context.Context, _ *Config) (db.Session, error) {
				return nil, errors.New("database error")
			}),
		)
		require.Nil(t, dbase)
		require.Error(t, err)
		require.Equal(t, "database error", err.Error())
	})

	t.Run("fail, no connection", func(t *testing.T) {
		dbase, err := New(
			context.Background(),
			&Credentials{},
			WithClientFunc(func(_ context.Context, _ *Config) (db.Session, error) {
				return nil, nil
			}),
		)
		require.NoError(t, err)
		require.NotNil(t, dbase)

		sess, err := dbase.SQLSessionContext(context.TODO())
		require.Nil(t, sess)
		require.Error(t, err)
		require.ErrorIs(t, err, errNoConnection)

		err = dbase.TxContext(context.TODO(), func(tx db.Session) error { return nil }, &sql.TxOptions{})
		require.Error(t, err)
		require.ErrorIs(t, err, errNoConnection)

		err = dbase.Ping()
		require.Error(t, err)
		require.ErrorIs(t, err, errNoConnection)

		require.NoError(t, dbase.Close())
	})
}

package mockdatabase

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nickbadlose/muzz/internal/database"
	"github.com/upper/db/v4/adapter/postgresql"
)

// NewWrappedMock returns a new mocked DB wrapped in the upper query builder for testing.
func NewWrappedMock() (database.Client, sqlmock.Sqlmock, error) {
	db, mockSQL, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	mockSQL.ExpectQuery(`SELECT CURRENT_DATABASE\(\) AS name`).
		WillReturnRows(
			sqlmock.NewRows([]string{`name`}).AddRow(`mock`),
		)
	session, err := postgresql.New(db)
	if err != nil {
		return nil, nil, err
	}

	return database.WrapSession(session), mockSQL, mockSQL.ExpectationsWereMet()
}

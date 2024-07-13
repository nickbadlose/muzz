package mockdatabase

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/postgresql"
)

// NewWrappedMock returns a new mocked DB wrapped in the upper query builder for testing.
func NewWrappedMock() (db.Session, sqlmock.Sqlmock, error) {
	database, mockSQL, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	mockSQL.ExpectQuery(`SELECT CURRENT_DATABASE\(\) AS name`).
		WillReturnRows(
			sqlmock.NewRows([]string{`name`}).AddRow(`mock`),
		)
	session, err := postgresql.New(database)
	if err != nil {
		return nil, nil, err
	}

	return session, mockSQL, mockSQL.ExpectationsWereMet()
}

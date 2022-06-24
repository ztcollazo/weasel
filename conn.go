package weasel

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// Connection holds all of the connection information, to be used in the models.
// It also contains the configured raw query builder and DB API, which both are not
// recommended to directly be used, but are useful if you need a complex query that is
// not officially supported. The query builder type comes from Squirrel, and the DB type is
// sqlx.DB.
type Connection struct {
	Builder sq.StatementBuilderType
	DB      *sqlx.DB
	driver  string
}

// The connect function creates a connection to the database. The opts string as the second parameter
// is a direct wrapper of sqlx.Connect; see sqlx's documentation for more details.
func Connect(driver string, opts string) Connection {
	db := sqlx.MustConnect(driver, opts)
	builder := sq.StatementBuilder.RunWith(db)
	if driver == "postgres" {
		builder = builder.PlaceholderFormat(sq.Dollar)
	}
	conn := Connection{
		DB:      db,
		Builder: builder,
		driver:  driver,
	}

	return conn
}

package weasel

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type Connection struct {
	Builder sq.StatementBuilderType
	DB      *sqlx.DB
	driver  string
}

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

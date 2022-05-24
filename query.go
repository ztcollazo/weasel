package weasel

import (
	"strings"

	sq "github.com/Masterminds/squirrel"
)

type InsertQuery[Doc any] struct {
	builder   sq.InsertBuilder
	ex        Doc
	conn      Connection
	tableName string
	pk        string
}

func (i InsertQuery[Doc]) Columns(columns ...string) InsertQuery[Doc] {
	i.builder = i.builder.Columns(columns...)
	return i
}

func (i InsertQuery[Doc]) Options(options ...string) InsertQuery[Doc] {
	i.builder = i.builder.Options(options...)
	return i
}

func (i InsertQuery[Doc]) Values(values ...any) InsertQuery[Doc] {
	i.builder = i.builder.Values(values...)
	return i
}

func (i InsertQuery[Doc]) Exec() (Doc, error) {
	ex := i.ex
	var id int64
	if i.conn.driver == "postgres" {
		s, g := i.builder.MustSql()
		s = strings.TrimSuffix(s, ";")
		i.conn.DB.QueryRow(s+" RETURNING "+i.pk, g...).Scan(&id)
	} else {
		res, err := i.builder.Exec()
		if err != nil {
			return ex, err
		}
		id, err = res.LastInsertId()
		if err != nil {
			return ex, err
		}
	}
	sql, args := i.conn.Builder.Select("*").From(i.tableName).Where(sq.Eq{i.pk: id}).MustSql()
	err := i.conn.DB.Get(&ex, sql, args...)
	return ex, err
}

func Insert[Doc any](ex Doc, tableName, pk string, conn Connection) InsertQuery[Doc] {
	return InsertQuery[Doc]{
		builder:   conn.Builder.Insert(tableName),
		ex:        ex,
		tableName: tableName,
		conn:      conn,
		pk:        pk,
	}
}

type SelectQuery[Doc any] struct {
	builder   sq.SelectBuilder
	ex        Doc
	conn      Connection
	tableName string
}

func (s SelectQuery[Doc]) Columns(columns ...string) SelectQuery[Doc] {
	s.builder = s.builder.Columns(columns...)
	return s
}

func (s SelectQuery[Doc]) Column(column any, args ...any) SelectQuery[Doc] {
	s.builder = s.builder.Column(column, args...)
	return s
}

func (s SelectQuery[Doc]) CrossJoin(join string, rest ...any) SelectQuery[Doc] {
	s.builder = s.builder.CrossJoin(join, rest...)
	return s
}

func (s SelectQuery[Doc]) Distinct() SelectQuery[Doc] {
	s.builder = s.builder.Distinct()
	return s
}

func (s SelectQuery[Doc]) GroupBy(groupBys ...string) SelectQuery[Doc] {
	s.builder = s.builder.GroupBy(groupBys...)
	return s
}

func (s SelectQuery[Doc]) Having(pred any, rest ...any) SelectQuery[Doc] {
	s.builder = s.builder.Having(pred, rest...)
	return s
}

func (s SelectQuery[Doc]) InnerJoin(join string, rest ...any) SelectQuery[Doc] {
	s.builder = s.builder.InnerJoin(join, rest...)
	return s
}

func (s SelectQuery[Doc]) Join(join string, rest ...any) SelectQuery[Doc] {
	s.builder = s.builder.Join(join, rest...)
	return s
}

func (s SelectQuery[Doc]) JoinClause(pred any, rest ...any) SelectQuery[Doc] {
	s.builder = s.builder.JoinClause(pred, rest...)
	return s
}

func (s SelectQuery[Doc]) LeftJoin(join string, rest ...any) SelectQuery[Doc] {
	s.builder = s.builder.LeftJoin(join, rest...)
	return s
}

func (s SelectQuery[Doc]) Limit(limit uint64) SelectQuery[Doc] {
	s.builder = s.builder.Limit(limit)
	return s
}

func (s SelectQuery[Doc]) Offset(offset uint64) SelectQuery[Doc] {
	s.builder = s.builder.Offset(offset)
	return s
}

func (s SelectQuery[Doc]) Options(options ...string) SelectQuery[Doc] {
	s.builder = s.builder.Options(options...)
	return s
}

func (s SelectQuery[Doc]) OrderBy(orderBys ...string) SelectQuery[Doc] {
	s.builder = s.builder.OrderBy(orderBys...)
	return s
}

func (s SelectQuery[Doc]) OrderByClause(pred any, args ...any) SelectQuery[Doc] {
	s.builder = s.builder.OrderByClause(pred, args...)
	return s
}

func (s SelectQuery[Doc]) RightJoin(join string, rest ...any) SelectQuery[Doc] {
	s.builder = s.builder.RightJoin(join, rest...)
	return s
}

func (s SelectQuery[Doc]) Where(pred any, args ...any) SelectQuery[Doc] {
	s.builder = s.builder.Where(pred, args...)
	return s
}

func (s SelectQuery[Doc]) Exec() (Doc, error) {
	sql, args := s.builder.MustSql()
	ex := s.ex
	err := s.conn.DB.Get(&ex, sql, args...)
	return ex, err
}

func Select[Doc any](ex Doc, columns []string, tableName string, conn Connection) SelectQuery[Doc] {
	return SelectQuery[Doc]{
		builder:   conn.Builder.Select(columns...).From(tableName),
		ex:        ex,
		tableName: tableName,
		conn:      conn,
	}
}

type SelectManyQuery[Doc any] struct {
	builder   sq.SelectBuilder
	ex        Doc
	conn      Connection
	tableName string
}

func (s SelectManyQuery[Doc]) Columns(columns ...string) SelectManyQuery[Doc] {
	s.builder = s.builder.Columns(columns...)
	return s
}

func (s SelectManyQuery[Doc]) Column(column any, args ...any) SelectManyQuery[Doc] {
	s.builder = s.builder.Column(column, args...)
	return s
}

func (s SelectManyQuery[Doc]) CrossJoin(join string, rest ...any) SelectManyQuery[Doc] {
	s.builder = s.builder.CrossJoin(join, rest...)
	return s
}

func (s SelectManyQuery[Doc]) Distinct() SelectManyQuery[Doc] {
	s.builder = s.builder.Distinct()
	return s
}

func (s SelectManyQuery[Doc]) GroupBy(groupBys ...string) SelectManyQuery[Doc] {
	s.builder = s.builder.GroupBy(groupBys...)
	return s
}

func (s SelectManyQuery[Doc]) Having(pred any, rest ...any) SelectManyQuery[Doc] {
	s.builder = s.builder.Having(pred, rest...)
	return s
}

func (s SelectManyQuery[Doc]) InnerJoin(join string, rest ...any) SelectManyQuery[Doc] {
	s.builder = s.builder.InnerJoin(join, rest...)
	return s
}

func (s SelectManyQuery[Doc]) Join(join string, rest ...any) SelectManyQuery[Doc] {
	s.builder = s.builder.Join(join, rest...)
	return s
}

func (s SelectManyQuery[Doc]) JoinClause(pred any, rest ...any) SelectManyQuery[Doc] {
	s.builder = s.builder.JoinClause(pred, rest...)
	return s
}

func (s SelectManyQuery[Doc]) LeftJoin(join string, rest ...any) SelectManyQuery[Doc] {
	s.builder = s.builder.LeftJoin(join, rest...)
	return s
}

func (s SelectManyQuery[Doc]) Limit(limit uint64) SelectManyQuery[Doc] {
	s.builder = s.builder.Limit(limit)
	return s
}

func (s SelectManyQuery[Doc]) Offset(offset uint64) SelectManyQuery[Doc] {
	s.builder = s.builder.Offset(offset)
	return s
}

func (s SelectManyQuery[Doc]) Options(options ...string) SelectManyQuery[Doc] {
	s.builder = s.builder.Options(options...)
	return s
}

func (s SelectManyQuery[Doc]) OrderBy(orderBys ...string) SelectManyQuery[Doc] {
	s.builder = s.builder.OrderBy(orderBys...)
	return s
}

func (s SelectManyQuery[Doc]) OrderByClause(pred any, args ...any) SelectManyQuery[Doc] {
	s.builder = s.builder.OrderByClause(pred, args...)
	return s
}

func (s SelectManyQuery[Doc]) RightJoin(join string, rest ...any) SelectManyQuery[Doc] {
	s.builder = s.builder.RightJoin(join, rest...)
	return s
}

func (s SelectManyQuery[Doc]) Where(pred any, args ...any) SelectManyQuery[Doc] {
	s.builder = s.builder.Where(pred, args...)
	return s
}

func (s SelectManyQuery[Doc]) Exec() ([]Doc, error) {
	sql, args := s.builder.MustSql()
	ex := []Doc{s.ex}
	err := s.conn.DB.Select(&ex, sql, args...)
	return ex, err
}

func SelectMany[Doc any](ex Doc, columns []string, tableName string, conn Connection) SelectManyQuery[Doc] {
	return SelectManyQuery[Doc]{
		builder:   conn.Builder.Select(columns...).From(tableName),
		ex:        ex,
		tableName: tableName,
		conn:      conn,
	}
}

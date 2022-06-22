package weasel

import (
	"reflect"

	sq "github.com/Masterminds/squirrel"
)

type InsertQuery[Doc DocumentBase] struct {
	builder sq.InsertBuilder
	model   *Model[Doc]
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
	ex := clone(i.model.ex, i.model)
	var id int64
	if i.model.Conn.driver == "postgres" {
		i.builder.Suffix("RETURNING id").QueryRow().Scan(&id)
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
	sql, args := i.model.Conn.Builder.Select("*").From(i.model.tableName).Where(Eq{i.model.pk: id}).MustSql()
	err := i.model.Conn.DB.Get(ex, sql, args...)
	return ex, err
}

func Insert[Doc DocumentBase](model *Model[Doc]) InsertQuery[Doc] {
	return InsertQuery[Doc]{
		builder: model.Conn.Builder.Insert(model.tableName),
		model:   model,
	}
}

type SelectQuery[Doc DocumentBase] struct {
	builder sq.SelectBuilder
	model   *Model[Doc]
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
	ex := clone(s.model.ex, s.model)
	err := s.model.Conn.DB.Get(ex, sql, args...)
	return ex, err
}

func Select[Doc DocumentBase](columns []string, model *Model[Doc]) SelectQuery[Doc] {
	return SelectQuery[Doc]{
		builder: model.Conn.Builder.Select(columns...).From(model.tableName),
		model:   model,
	}
}

type SelectManyQuery[Doc DocumentBase] struct {
	builder sq.SelectBuilder
	model   *Model[Doc]
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
	p := clone(s.model.ex, s.model)
	ex := []Doc{p}
	err := s.model.Conn.DB.Select(&ex, sql, args...)
	for _, d := range ex {
		callInit(d, s.model)
	}
	return ex, err
}

func SelectMany[Doc DocumentBase](columns []string, model *Model[Doc]) SelectManyQuery[Doc] {
	return SelectManyQuery[Doc]{
		builder: model.Conn.Builder.Select(columns...).From(model.tableName),
		model:   model,
	}
}

func clone[Doc DocumentBase](d Doc, m *Model[Doc]) Doc {
	t := reflect.Indirect(reflect.ValueOf(d)).Type()
	v := reflect.New(t)
	v.MethodByName("Create").Call([]reflect.Value{v, reflect.ValueOf(m)})
	return v.Interface().(Doc)
}

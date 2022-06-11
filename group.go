package weasel

import (
	"errors"

	"github.com/carlmjohnson/truthy"
)

type whereable interface {
	ToSql() (string, []any, error)
}

type Group[Doc DocumentBase] struct {
	Where     whereable
	Model     Model[Doc]
	innerJoin string
	on        whereable
}

func (m Group[Doc]) Find(value any) (Doc, error) {
	stmt := Select([]string{"*"}, m.Model).Where(m.Where).Where(Eq{m.Model.pk: value})
	if truthy.Value(m.innerJoin) {
		stmt.InnerJoin(m.innerJoin, m.on)
	}
	doc, err := stmt.Exec()
	if err == nil {
		callInit(doc, &m.Model)
		if len(doc.errors()) > 0 {
			return doc, errors.New("document is invalid")
		}
		return doc, nil
	}
	return doc, err
}

func (m Group[Doc]) FindBy(name string, value any) (Doc, error) {
	stmt := Select([]string{"*"}, m.Model).Where(m.Where).Where(Eq{name: value})
	if truthy.Value(m.innerJoin) {
		stmt.InnerJoin(m.innerJoin, m.on)
	}
	doc, err := stmt.Exec()
	if err == nil {
		callInit(doc, &m.Model)
		if len(doc.errors()) > 0 {
			return doc, errors.New("document is invalid")
		}
		return doc, nil
	}
	return doc, err
}

func (m Group[Doc]) All() SelectManyQuery[Doc] {
	stmt := SelectMany([]string{"*"}, m.Model).Where(m.Where)
	if truthy.Value(m.innerJoin) {
		stmt.InnerJoin(m.innerJoin, m.on)
	}
	return stmt
}

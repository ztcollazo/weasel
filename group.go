package weasel

import (
	"errors"
	"reflect"
	"strings"

	"github.com/carlmjohnson/truthy"
)

type whereable interface {
	ToSql() (string, []any, error)
}

type Group[Doc DocumentBase] struct {
	Where     whereable
	Model     *Model[Doc]
	innerJoin string
	on        string
	id        string
	groups    map[string]*Group[Doc]
	order     string
}

func (m Group[Doc]) Find(value any) (Doc, error) {
	stmt := Select([]string{m.Model.tableName + ".*"}, m.Model).Where(m.Where).Where(Eq{m.Model.tableName + "." + m.Model.pk: value})
	if truthy.Value(m.innerJoin) {
		stmt = stmt.InnerJoin(m.innerJoin + " ON (" + m.on + " = " + m.id + ")")
	}
	doc, err := stmt.Exec()
	if err == nil {
		callInit(doc, m.Model)
		if len(doc.AllErrors()) > 0 {
			return doc, errors.New("document is invalid")
		}
		return doc, nil
	}
	return doc, err
}

func (m Group[Doc]) FindBy(name string, value any) (Doc, error) {
	stmt := Select([]string{"*"}, m.Model).Where(m.Where).Where(Eq{name: value})
	if truthy.Value(m.innerJoin) {
		stmt = stmt.InnerJoin(m.innerJoin, m.on)
	}
	doc, err := stmt.Exec()
	if err == nil {
		callInit(doc, m.Model)
		if len(doc.AllErrors()) > 0 {
			return doc, errors.New("document is invalid")
		}
		return doc, nil
	}
	return doc, err
}

func (m Group[Doc]) All() SelectManyQuery[Doc] {
	stmt := SelectMany([]string{"*"}, m.Model).Where(m.Where).OrderBy(m.order)
	if truthy.Value(m.innerJoin) {
		stmt = stmt.InnerJoin(m.innerJoin, m.on)
	}
	return stmt
}

// Create creates a document and adds it to the database
// TODO: Make create conform to `where` clause
func (m Group[Doc]) Create(d Doc) (Doc, error) {
	callInit(d, m.Model)
	if len(d.AllErrors()) > 0 {
		return d, errors.New("document is invalid")
	}
	v := reflect.Indirect(reflect.ValueOf(d))
	columns := make([]string, 0)
	values := make([]any, 0)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		t := v.Type().Field(i)
		if f, ok := m.Model.fields[t.Tag.Get("db")]; ok && !f.PrimaryKey {
			columns = append(columns, f.DBName)
			values = append(values, field.Interface())
		}
	}
	doc, err := Insert(m.Model).Columns(columns...).Values(values...).Exec()
	if err == nil {
		callInit(doc, m.Model)
		// And just in case
		if len(doc.AllErrors()) > 0 {
			return doc, errors.New("document is invalid")
		}
		return doc, nil
	}
	return doc, err
}

func (m *Group[Doc]) CreateGroup(name string, expr whereable) {
	m.groups[name] = &Group[Doc]{
		Where:  And{m.Where, expr},
		Model:  m.Model,
		groups: make(map[string]*Group[Doc]),
	}
}

func (m Group[Doc]) FromGroup(name string) *Group[Doc] {
	return m.groups[name]
}

func (m Group[Doc]) Count() (int, error) {
	var cnt int
	err := m.Model.Conn.Builder.Select("COUNT(*)").From(m.Model.tableName).Where(m.Where).Scan(&cnt)
	return cnt, err
}

func (m Group[Doc]) Exists(id any) (bool, error) {
	var cnt int
	err := m.Model.Conn.Builder.Select("COUNT(*)").From(m.Model.tableName).Where(m.Where).Where(Eq{m.Model.pk: id}).Scan(&cnt)
	return cnt != 0, err
}

func (m *Group[Doc]) Order(by string) {
	m.order = by
}

func (m Group[Doc]) First() (Doc, error) {
	return m.Nth(1)
}

func (m Group[Doc]) Second() (Doc, error) {
	return m.Nth(2)
}

func (m Group[Doc]) Third() (Doc, error) {
	return m.Nth(3)
}

func (m Group[Doc]) Fourth() (Doc, error) {
	return m.Nth(4)
}

func (m Group[Doc]) Fifth() (Doc, error) {
	return m.Nth(5)
}

func (m Group[Doc]) Last() (Doc, error) {
	return m.NthToLast(1)
}

func (m Group[Doc]) SecondToLast() (Doc, error) {
	return m.NthToLast(2)
}

func (m Group[Doc]) NthToLast(id int) (Doc, error) {
	var order string
	if strings.Contains(m.order, "ASC") {
		order = strings.Replace(m.order, "ASC", "DESC", 1)
	} else {
		order = strings.Replace(m.order, "DESC", "ASC", 1)
	}
	return Select([]string{"*"}, m.Model).Where(m.Where).Limit(1).OrderBy(order).Offset(uint64(id - 1)).Exec()
}

func (m Group[Doc]) Nth(id int) (Doc, error) {
	return Select([]string{"*"}, m.Model).Where(m.Where).Limit(1).Offset(uint64(id - 1)).Exec()
}

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

// Group is the foundation on both Model and Group functionality.
// It has many of the methods used in the model, and contains all of the
// querying utilities.
type Group[Doc DocumentBase] struct {
	Where     whereable
	Model     *Model[Doc]
	innerJoin string
	on        string
	id        string
	groups    map[string]*Group[Doc]
	order     string
}

func NewGroupWith[Doc DocumentBase](where whereable, model *Model[Doc], innerJoin, on, id, order string, groups map[string]*Group[Doc]) *Group[Doc] {
	return &Group[Doc]{
		Where:     where,
		Model:     model,
		innerJoin: innerJoin,
		on:        on,
		id:        id,
		order:     order,
		groups:    groups,
	}
}

// Find takes the primary key value and finds the corresponding document.
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

// FindBy takes a column name and value and finds the corresponding document.
// If you want to find multiple, use All().Where(weasel.Eq{key: value}).
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

// All returns all of the documents in the group or model.
// It returns a query builder that contains functions including Where, OrderBy, GroupBy.
// For more information and functions, see SelectManyQuery and its methods.
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

// CreateGroup adds a group to the model or group that can be accessed by FromGroup.
// It respects the current group's where clause and contains all of the querying functionality
// and utilities.
func (m *Group[Doc]) CreateGroup(name string, expr whereable) {
	m.groups[name] = &Group[Doc]{
		Where:  And{m.Where, expr},
		Model:  m.Model,
		groups: make(map[string]*Group[Doc]),
	}
}

// FromGroup returns the group that the name parameter points to.
// See CreateGroup() and Group for more information.
func (m Group[Doc]) FromGroup(name string) *Group[Doc] {
	return m.groups[name]
}

// Count returns the number of documents in the group or model.
func (m Group[Doc]) Count() (int, error) {
	var cnt int
	err := m.Model.Conn.Builder.Select("COUNT(*)").From(m.Model.tableName).Where(m.Where).Scan(&cnt)
	return cnt, err
}

// Exists checks if the document with the given primary key exists.
func (m Group[Doc]) Exists(id any) (bool, error) {
	var cnt int
	err := m.Model.Conn.Builder.Select("COUNT(1)").From(m.Model.tableName).Where(m.Where).Where(Eq{m.Model.pk: id}).Scan(&cnt)
	return cnt != 0, err
}

// Order sets the order that the documents should be sorted by when queried.
func (m *Group[Doc]) Order(by string) {
	m.order = by
}

// Get order returns the order of the documents used when queried, for example "id ASC"
func (m *Group[Doc]) GetOrder() string {
	return m.order
}

// First returns the first document from the table, via the set order clause.
// See Nth for more information.
func (m Group[Doc]) First() (Doc, error) {
	return m.Nth(1)
}

// Second returns the second document.
// See Nth for more information.
func (m Group[Doc]) Second() (Doc, error) {
	return m.Nth(2)
}

// Third returns the third document.
// See Nth for more information.
func (m Group[Doc]) Third() (Doc, error) {
	return m.Nth(3)
}

// Fourth returns the fourth document.
// See Nth for more information.
func (m Group[Doc]) Fourth() (Doc, error) {
	return m.Nth(4)
}

// Fifth returns the fifth document.
// See Nth for more information.
func (m Group[Doc]) Fifth() (Doc, error) {
	return m.Nth(5)
}

// Last returns the last document in the table, via the opposite of order.
// See NthToLast for more information.
func (m Group[Doc]) Last() (Doc, error) {
	return m.NthToLast(1)
}

// SecondToLast returns the second to last document in the table.
// See NthToLast for more information.
func (m Group[Doc]) SecondToLast() (Doc, error) {
	return m.NthToLast(2)
}

// NthToLast returns the last document at the given index.
// For example:
//
//	Person.NthToLast(3) // Returns the third to last document.
func (m Group[Doc]) NthToLast(id int) (Doc, error) {
	var order string
	if strings.Contains(m.order, "ASC") {
		order = strings.Replace(m.order, "ASC", "DESC", 1)
	} else {
		order = strings.Replace(m.order, "DESC", "ASC", 1)
	}
	return Select([]string{"*"}, m.Model).Where(m.Where).Limit(1).OrderBy(order).Offset(uint64(id - 1)).Exec()
}

// Nth returns the document at the given index.
// For example:
//
//	Person.Nth(6) // Returns the sixth document.
func (m Group[Doc]) Nth(id int) (Doc, error) {
	return Select([]string{"*"}, m.Model).Where(m.Where).Limit(1).Offset(uint64(id - 1)).Exec()
}

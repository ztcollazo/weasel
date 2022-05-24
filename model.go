package weasel

import (
	"reflect"

	sq "github.com/Masterminds/squirrel"
)

type Field struct {
	Name       string
	Type       string
	Default    string
	NotNil     bool
	PrimaryKey bool
}

type Model[Doc any] struct {
	Conn      Connection
	tableName string
	pk        string
	fields    map[string]Field
	ex        Doc
}

func (m Model[Doc]) Create(d Doc) (Doc, error) {
	v := reflect.ValueOf(d)
	columns := make([]string, 0)
	values := make([]any, 0)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		t := v.Type().Field(i)
		if f, ok := m.fields[t.Name]; ok && !f.PrimaryKey {
			columns = append(columns, m.fields[t.Name].Name)
			values = append(values, field.Interface())
		}
	}
	return Insert(m.ex, m.tableName, m.pk, m.Conn).Columns(columns...).Values(values...).Exec()
}

func (m Model[Doc]) Find(value any) (Doc, error) {
	return Select(m.ex, []string{"*"}, m.tableName, m.Conn).Where(sq.Eq{m.pk: value}).Exec()
}

func (m Model[Doc]) FindBy(name string, value any) (Doc, error) {
	return Select(m.ex, []string{"*"}, m.tableName, m.Conn).Where(sq.Eq{name: value}).Exec()
}

func (m Model[Doc]) All() SelectManyQuery[Doc] {
	return SelectMany(m.ex, []string{"*"}, m.tableName, m.Conn)
}

func Create[Doc any](conn Connection, ex Doc, name string) Model[Doc] {
	var pk string
	var fields = make(map[string]Field, 0)
	t := reflect.TypeOf(ex)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if name, ok := field.Tag.Lookup("db"); !ok {
			continue
		} else {
			f := Field{
				Name: name,
			}
			if tp, to := field.Tag.Lookup("type"); !to {
				f.Type = field.Type.Name()
			} else {
				f.Type = tp
			}
			f.Default = field.Tag.Get("default")
			_, f.NotNil = field.Tag.Lookup("notnil")
			if _, isP := field.Tag.Lookup("pk"); isP {
				pk = name
				f.PrimaryKey = true
			} else {
				f.PrimaryKey = false
			}
			fields[field.Name] = f
		}
	}
	return Model[Doc]{
		Conn:      conn,
		tableName: name,
		pk:        pk,
		fields:    fields,
		ex:        ex,
	}
}

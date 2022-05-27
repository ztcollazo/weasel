package weasel

import (
	"reflect"
)

type Field struct {
	Name       string
	DBName     string
	Type       string
	Default    string
	NotNil     bool
	PrimaryKey bool
}

type Model[Doc docbase] struct {
	Conn      Connection
	tableName string
	pk        string
	fields    map[string]Field
	ex        Doc
}

func (m Model[Doc]) Create(d Doc) (Doc, error) {
	callInit(d)
	v := reflect.Indirect(reflect.ValueOf(d))
	columns := make([]string, 0)
	values := make([]any, 0)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		t := v.Type().Field(i)
		if f, ok := m.fields[t.Tag.Get("db")]; ok && !f.PrimaryKey {
			columns = append(columns, f.DBName)
			values = append(values, field.Interface())
		}
	}
	return Insert(m).Columns(columns...).Values(values...).Exec()
}

func (m Model[Doc]) Find(value any) (Doc, error) {
	return Select([]string{"*"}, m).Where(Eq{m.pk: value}).Exec()
}

func (m Model[Doc]) FindBy(name string, value any) (Doc, error) {
	return Select([]string{"*"}, m).Where(Eq{name: value}).Exec()
}

func (m Model[Doc]) All() SelectManyQuery[Doc] {
	return SelectMany([]string{"*"}, m)
}

func Create[Doc document[Doc]](conn Connection, ex Doc, name string) Model[Doc] {
	doc := ex
	var pk string
	var fields = make(map[string]Field, 0)
	t := reflect.Indirect(reflect.ValueOf(doc)).Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if name, ok := field.Tag.Lookup("db"); !ok {
			continue
		} else {
			f := Field{
				Name:   field.Name,
				DBName: name,
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
			fields[name] = f
		}
	}
	model := Model[Doc]{
		Conn:      conn,
		tableName: name,
		pk:        pk,
		fields:    fields,
		ex:        doc,
	}
	doc.init(doc, model)
	return model
}

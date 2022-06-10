package weasel

import (
	"errors"
	"reflect"

	"github.com/carlmjohnson/truthy"
)

type Field struct {
	Name       string
	DBName     string
	Type       string
	Default    string
	NotNil     bool
	PrimaryKey bool
}

type Relation struct {
	Name       string
	Variant    string
	Key        string
	ForeignKey string
	Table      string
	Through    string
}

type Model[Doc DocumentBase] struct {
	Conn      Connection
	tableName string
	pk        string
	fields    map[string]Field
	relations map[string]Relation
	ex        Doc
}

func (m Model[Doc]) Create(d Doc) (Doc, error) {
	callInit(d, &m)
	if len(d.errors()) > 0 {
		return d, errors.New("document is invalid")
	}
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
	doc, err := Insert(m).Columns(columns...).Values(values...).Exec()
	if err == nil {
		callInit(doc, &m)
		// And just in case
		if len(doc.errors()) > 0 {
			return doc, errors.New("document is invalid")
		}
		return doc, nil
	}
	return doc, err
}

func (m Model[Doc]) Find(value any) (Doc, error) {
	doc, err := Select([]string{"*"}, m).Where(Eq{m.pk: value}).Exec()
	if err == nil {
		callInit(doc, &m)
		if len(doc.errors()) > 0 {
			return doc, errors.New("document is invalid")
		}
		return doc, nil
	}
	return doc, err
}

func (m Model[Doc]) FindBy(name string, value any) (Doc, error) {
	doc, err := Select([]string{"*"}, m).Where(Eq{name: value}).Exec()
	if err == nil {
		callInit(doc, &m)
		if len(doc.errors()) > 0 {
			return doc, errors.New("document is invalid")
		}
		return doc, nil
	}
	return doc, err
}

func (m Model[Doc]) All() SelectManyQuery[Doc] {
	return SelectMany([]string{"*"}, m)
}

func Create[Doc document[Doc]](conn Connection, ex Doc, name string) Model[Doc] {
	doc := ex
	var pk string
	var relations = map[string]Relation{}
	var fields = make(map[string]Field, 0)
	t := reflect.Indirect(reflect.ValueOf(doc)).Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if name, ok := field.Tag.Lookup("db"); !ok {
			if belongsTo, bt := field.Tag.Lookup("belongsto"); bt {
				fk := or(field.Tag.Get("fk"), "id")
				key := or(field.Tag.Get("key"), belongsTo+"_id")
				relation := Relation{
					Name:       field.Name,
					Table:      belongsTo,
					ForeignKey: fk,
					Key:        key,
					Variant:    "belongsTo",
				}
				relations["belongsTo"+belongsTo] = relation
			} else if hasMany, hm := field.Tag.Lookup("hasmany"); hm {
				foreignKey := or(field.Tag.Get("fk"), name+"_id")
				key := or(field.Tag.Get("key"), "id")
				relation := Relation{
					Name:       field.Name,
					Table:      hasMany,
					ForeignKey: foreignKey,
					Variant:    "hasMany",
					Key:        key,
				}
				if through, hmt := field.Tag.Lookup("through"); hmt {
					relation.Through = through
				}
				relations["hasMany"+hasMany] = relation
			} else if hasOne, ho := field.Tag.Lookup("hasone"); ho {
				foreignKey := or(field.Tag.Get("fk"), name+"_id")
				key := or(field.Tag.Get("key"), "id")
				relation := Relation{
					Name:       field.Name,
					ForeignKey: foreignKey,
					Table:      hasOne,
					Variant:    "hasOne",
					Key:        key,
				}
				relations["hasOne"+hasOne] = relation
			} else {
				continue
			}
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
		relations: relations,
	}
	doc.Create(doc, &model)
	return model
}

func or[T any](vals ...T) T {
	return truthy.First(vals...)
}

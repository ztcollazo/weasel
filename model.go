package weasel

import (
	"reflect"

	"github.com/carlmjohnson/truthy"
)

type Init[Doc document[Doc]] func(*Model[Doc])

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
	*Group[Doc]
	Conn      Connection
	tableName string
	pk        string
	fields    map[string]Field
	relations map[string]Relation
	ex        Doc
	vals      map[string]any
}

func (m *Model[Doc]) Set(key string, val any) {
	m.vals[key] = val
}

func (m Model[Doc]) Get(key string) any {
	return m.vals[key]
}

func Create[Doc document[Doc]](conn Connection, ex Doc, name string, inits ...Init[Doc]) *Model[Doc] {
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
					relation.Key = or(field.Tag.Get("key"), name+"_id")
					relation.ForeignKey = or(field.Tag.Get("fk"), hasMany+"_id")
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
	model := &Model[Doc]{
		Conn:      conn,
		tableName: name,
		pk:        pk,
		fields:    fields,
		ex:        doc,
		relations: relations,
		vals:      make(map[string]any),
	}
	model.Group = &Group[Doc]{
		Model:  model,
		Where:  Eq{},
		groups: make(map[string]*Group[Doc]),
		order:  pk + " ASC",
	}
	doc.Create(doc, model)
	for _, init := range inits {
		init(model)
	}
	return model
}

func or[T any](vals ...T) T {
	return truthy.First(vals...)
}

package weasel

import (
	"errors"
	"fmt"
)

// Type HasMany is the type used to represent a one-to-many or many-to-many relationship in a schema.
// Use the following struct tags to give it more information:
//  - key: default is the primary key. This is the column that the foreign key points to.
//  - fk: the name of the foreign key column.
//  - through: the join table for many-to-many relationships.
//  - hasmany: the table that it has many of.
type HasMany[Doc document[Doc]] func(...Doc) Group[Doc]

// Type BelongsTo is the type used to represent the flipside of a one-to-many relationship in a schema.
// Use the following struct tags to give it more information:
//  - key: default is the primary key. This is the column that the foreign key points to.
//  - fk: the name of the foreign key column.
//  - belongsto: the table that it belongs to.
type BelongsTo[Doc document[Doc]] func(...Doc) (Doc, error)

// Type HasMany is the type used to represent a one-to-one relationship in a schema.
// Use the following struct tags to give it more information:
//  - key: default is the primary key. This is the column that the foreign key points to.
//  - fk: the name of the foreign key column.
//  - hasone: the table that it has one of.
type HasOne[Doc document[Doc]] func(...Doc) (Doc, error)

// Use has many populates the field that returns the has many relationship, specified in the schema.
// It takes a document and the model that the document has many of.
func UseHasMany[Doc document[Doc], Rel document[Rel]](doc Doc, model *Model[Rel]) {
	rel := doc.model().relations["hasMany"+model.tableName]
	var fn HasMany[Rel] = func(d ...Rel) Group[Rel] {
		if rel.Through != "" {
			where := Eq{}
			return Group[Rel]{
				Model:     model,
				Where:     where,
				innerJoin: rel.Through,
				on:        rel.Through + "." + rel.Key,
				id:        fmt.Sprint(doc.Get(doc.model().pk)),
				groups:    make(map[string]*Group[Rel]),
				order:     model.order,
			}
		} else {
			return Group[Rel]{
				Where:  Eq{rel.ForeignKey: doc.Get(rel.Key)},
				Model:  model,
				groups: make(map[string]*Group[Rel]),
				order:  model.order,
			}
		}
	}
	doc.Set(rel.Name, fn)
}

// UseBelongsTo populates the field that returns the belongs to relationship, specified in the schema.
// It takes a document and the model that the document belongs to.
func UseBelongsTo[Doc document[Doc], Rel document[Rel]](doc Doc, model *Model[Rel]) {
	rel := doc.model().relations["belongsTo"+model.tableName]
	var fn BelongsTo[Rel] = func(d ...Rel) (Rel, error) {
		e, err := Select([]string{"*"}, model).Where(Eq{rel.ForeignKey: doc.Get(rel.Key)}).Exec()
		if err == nil {
			callInit(e, model)
			if len(e.AllErrors()) > 0 {
				return e, errors.New("document is invalid")
			}
			return e, nil
		}
		return e, err
	}
	doc.Set(rel.Name, fn)
}

// UseHasOne populates the field that returns the has one relationship, specified in the schema.
// it takes a document and the model that the document has one of.
func UseHasOne[Doc document[Doc], Rel document[Rel]](doc Doc, model *Model[Rel]) {
	rel := doc.model().relations["hasOne"+model.tableName]
	var fn HasOne[Rel] = func(d ...Rel) (Rel, error) {
		e, err := Select([]string{"*"}, model).Where(Eq{rel.ForeignKey: doc.Get(rel.Key)}).Exec()
		if err == nil {
			callInit(e, model)
			if len(e.AllErrors()) > 0 {
				return e, errors.New("document is invalid")
			}
			return e, nil
		}
		return e, err
	}
	doc.Set(rel.Name, fn)
}

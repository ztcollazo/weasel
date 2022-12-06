package weasel

import (
	"errors"
	"fmt"
)

// Type HasMany is the type used to represent a one-to-many or many-to-many relationship in a schema.
// Use the following struct tags to give it more information:
//   - key: default is the primary key. This is the column that the foreign key points to.
//   - fk: the name of the foreign key column.
//   - through: the join table for many-to-many relationships.
//   - hasmany: the table that it has many of.
type HasMany[Doc document[Doc]] func() *Group[Doc]

// Type BelongsTo is the type used to represent the flipside of a one-to-many relationship in a schema.
// Use the following struct tags to give it more information:
//   - key: default is the primary key. This is the column that the foreign key points to.
//   - fk: the name of the foreign key column.
//   - belongsto: the table that it belongs to.
type BelongsTo[Doc document[Doc]] func() (Doc, error)

// Type HasMany is the type used to represent a one-to-one relationship in a schema.
// Use the following struct tags to give it more information:
//   - key: default is the primary key. This is the column that the foreign key points to.
//   - fk: the name of the foreign key column.
//   - hasone: the table that it has one of.
type HasOne[Doc document[Doc]] func() (Doc, error)

// Use has many populates the field that returns the has many relationship, specified in the schema.
// It takes a document and the model that the document has many of.
//
// Deprecated: the use package now contains the functions HasMany, BelongsTo, and HasOne to take
// the place of UseHasMany, UseBelongsTo, and UseHasOne. These functions follow the middleware format.
// These will be removed in a later release. You may still use the types provided in the main package.
func UseHasMany[Doc document[Doc], Rel document[Rel]](doc Doc, model *Model[Rel]) {
	rel := doc.GetModel().Relations()["hasMany"+model.Name()]
	var fn HasMany[Rel] = func() *Group[Rel] {
		if rel.Through != "" {
			return &Group[Rel]{
				Model:     model,
				Where:     Eq{},
				innerJoin: rel.Through,
				on:        rel.Through + "." + rel.Key,
				id:        fmt.Sprint(doc.Get(doc.PrimaryKey())),
				groups:    make(map[string]*Group[Rel]),
				order:     model.GetOrder(),
			}
		} else {
			return &Group[Rel]{
				Where:  Eq{rel.ForeignKey: doc.Get(rel.Key)},
				Model:  model,
				groups: make(map[string]*Group[Rel]),
				order:  model.GetOrder(),
			}
		}
	}
	doc.Set(rel.Name, fn)
}

// UseBelongsTo populates the field that returns the belongs to relationship, specified in the schema.
// It takes a document and the model that the document belongs to.
//
// Deprecated: the use package now contains the functions HasMany, BelongsTo, and HasOne to take
// the place of UseHasMany, UseBelongsTo, and UseHasOne. These functions follow the middleware format.
// These will be removed in a later release. You may still use the types provided in the main package.
func UseBelongsTo[Doc document[Doc], Rel document[Rel]](doc Doc, model *Model[Rel]) {
	rel := doc.GetModel().Relations()["belongsTo"+model.Name()]
	var fn BelongsTo[Rel] = func() (Rel, error) {
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
//
// Deprecated: the use package now contains the functions HasMany, BelongsTo, and HasOne to take
// the place of UseHasMany, UseBelongsTo, and UseHasOne. These functions follow the middleware format.
// These will be removed in a later release. You may still use the types provided in the main package.
func UseHasOne[Doc document[Doc], Rel document[Rel]](doc Doc, model *Model[Rel]) {
	rel := doc.GetModel().Relations()["hasOne"+model.Name()]
	var fn HasOne[Rel] = func() (Rel, error) {
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

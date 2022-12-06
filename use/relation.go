package use

import (
	"errors"
	"fmt"

	"github.com/ztcollazo/weasel"
)

type document[Doc weasel.DocumentBase] interface {
	weasel.DocumentBase
	Create(Doc, *weasel.Model[Doc])
	GetModel() *weasel.Model[Doc]
}

// use.HasMany populates the field that returns the has many relationship, specified in the schema.
// It takes a document and the model that the document has many of.
func HasMany[Doc document[Doc], Rel document[Rel]](model *weasel.Model[Rel]) weasel.Middleware {
	return func(doc weasel.DocumentBase) {
		dm := doc.Get("Model").(*weasel.Model[Doc])
		rel := dm.Relations()["hasMany"+model.Name()]
		var fn weasel.HasMany[Rel] = func() *weasel.Group[Rel] {
			if rel.Through != "" {
				return weasel.NewGroupWith(weasel.Eq{}, model, rel.Through, rel.Through+"."+rel.Key, fmt.Sprint(doc.Get(doc.PrimaryKey())), model.GetOrder(), make(map[string]*weasel.Group[Rel]))
			} else {
				return weasel.NewGroupWith(weasel.Eq{rel.ForeignKey: doc.Get(rel.Key)}, model, "", "", "", model.GetOrder(), make(map[string]*weasel.Group[Rel]))
			}
		}
		doc.Set(rel.Name, fn)
	}
}

// func UseHasMany[Doc document[Doc], Rel document[Rel]](doc Doc, model *weasel.Model[Rel]) {
// 	rel := doc.GetModel().Relations()["hasMany"+model.Name()]
// 	var fn weasel.HasMany[Rel] = func(d ...Rel) weasel.Group[Rel] {
// 		if rel.Through != "" {
// 			return weasel.Group[Rel]{
// 				Model:     model,
// 				Where:     weasel.Eq{},
// 				innerJoin: rel.Through,
// 				on:        rel.Through + "." + rel.Key,
// 				id:        fmt.Sprint(doc.Get(doc.PrimaryKey())),
// 				groups:    make(map[string]*Group[Rel]),
// 				order:     model.GetOrder(),
// 			}
// 		} else {
// 			return weasel.Group[Rel]{
// 				Where:  weasel.Eq{rel.ForeignKey: doc.Get(rel.Key)},
// 				Model:  model,
// 				groups: make(map[string]*weasel.Group[Rel]),
// 				order:  model.GetOrder(),
// 			}
// 		}
// 	}
// 	doc.Set(rel.Name, fn)
// }

// use.BelongsTo populates the field that returns the belongs to relationship, specified in the schema.
// It takes a document and the model that the document belongs to.
func BelongsTo[Doc document[Doc], Rel document[Rel]](model *weasel.Model[Rel]) weasel.Middleware {
	return func(doc weasel.DocumentBase) {
		dm := doc.Get("Model").(*weasel.Model[Doc])
		rel := dm.Relations()["belongsTo"+model.Name()]
		var fn weasel.BelongsTo[Rel] = func() (Rel, error) {
			e, err := weasel.Select([]string{"*"}, model).Where(weasel.Eq{rel.ForeignKey: doc.Get(rel.Key)}).Exec()
			if err == nil {
				e.Create(e, model)
				e.Init()
				if len(e.AllErrors()) > 0 {
					return e, errors.New("document is invalid")
				}
				return e, nil
			}
			return e, err
		}
		doc.Set(rel.Name, fn)
	}
}

// func UseBelongsTo[Doc document[Doc], Rel document[Rel]](doc Doc, model *weasel.Model[Rel]) {
// 	rel := doc.GetModel().Relations()["belongsTo"+model.Name()]
// 	var fn weasel.BelongsTo[Rel] = func(d ...Rel) (Rel, error) {
// 		e, err := weasel.Select([]string{"*"}, model).Where(weasel.Eq{rel.ForeignKey: doc.Get(rel.Key)}).Exec()
// 		if err == nil {
// 			callInit(e, model)
// 			if len(e.AllErrors()) > 0 {
// 				return e, errors.New("document is invalid")
// 			}
// 			return e, nil
// 		}
// 		return e, err
// 	}
// 	doc.Set(rel.Name, fn)
// }

// use.HasOne populates the field that returns the has one relationship, specified in the schema.
// it takes a document and the model that the document has one of.
func HasOne[Doc document[Doc], Rel document[Rel]](model *weasel.Model[Rel]) weasel.Middleware {
	return func(doc weasel.DocumentBase) {
		dm := doc.Get("Model").(*weasel.Model[Doc])
		rel := dm.Relations()["hasOne"+model.Name()]
		var fn weasel.HasOne[Rel] = func() (Rel, error) {
			e, err := weasel.Select([]string{"*"}, model).Where(weasel.Eq{rel.ForeignKey: doc.Get(rel.Key)}).Exec()
			if err == nil {
				e.Create(e, model)
				e.Init()
				if len(e.AllErrors()) > 0 {
					return e, errors.New("document is invalid")
				}
				return e, nil
			}
			return e, err
		}
		doc.Set(rel.Name, fn)
	}
}

// func UseHasOne[Doc document[Doc], Rel document[Rel]](doc Doc, model *weasel.Model[Rel]) {
// 	rel := doc.GetModel().Relations()["hasOne"+model.Name()]
// 	var fn weasel.HasOne[Rel] = func(d ...Rel) (Rel, error) {
// 		e, err := weasel.Select([]string{"*"}, model).Where(weasel.Eq{rel.ForeignKey: doc.Get(rel.Key)}).Exec()
// 		if err == nil {
// 			callInit(e, model)
// 			if len(e.AllErrors()) > 0 {
// 				return e, errors.New("document is invalid")
// 			}
// 			return e, nil
// 		}
// 		return e, err
// 	}
// 	doc.Set(rel.Name, fn)
// }

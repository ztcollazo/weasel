package weasel

import (
	"errors"
	"fmt"
)

type HasMany[Doc document[Doc]] func(...Doc) Group[Doc]

type BelongsTo[Doc document[Doc]] func(...Doc) (Doc, error)

type HasOne[Doc document[Doc]] func(...Doc) (Doc, error)

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

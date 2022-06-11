package weasel

type HasMany[Doc document[Doc]] func(...Doc) Group[Doc]

type BelongsTo[Doc document[Doc]] func(...Doc) (Doc, error)

type HasOne[Doc document[Doc]] func(...Doc) (Doc, error)

func UseHasMany[Doc document[Doc], Rel document[Rel]](doc Doc, model *Model[Rel]) {
	g := get(doc)
	s := set(doc)
	rel := doc.model().relations["hasMany"+model.tableName]
	var fn HasMany[Rel] = func(d ...Rel) Group[Rel] {
		if rel.Through != "" {
			return Group[Rel]{
				Where:     Eq{rel.Through + rel.ForeignKey: g(rel.Key)},
				Model:     *model,
				innerJoin: model.tableName,
				on:        Eq{model.tableName + "_" + model.pk: model.pk},
			}
		} else {
			return Group[Rel]{
				Where: Eq{rel.ForeignKey: g(rel.Key)},
				Model: *model,
			}
		}
	}
	s(rel.Name, fn)
}

func UseBelongsTo[Doc document[Doc], Rel document[Rel]](doc Doc, model *Model[Rel]) {
	s := set(doc)
	g := get(doc)
	rel := doc.model().relations["belongsTo"+model.tableName]
	var fn BelongsTo[Rel] = func(d ...Rel) (Rel, error) {
		return Select([]string{"*"}, *model).Where(Eq{rel.ForeignKey: g(rel.Key)}).Exec()
	}
	s(rel.Name, fn)
}

func UseHasOne[Doc document[Doc], Rel document[Rel]](doc Doc, model *Model[Rel]) {
	g := get(doc)
	s := set(doc)
	rel := doc.model().relations["hasOne"+model.tableName]
	var fn HasOne[Rel] = func(d ...Rel) (Rel, error) {
		return Select([]string{"*"}, *model).Where(Eq{rel.ForeignKey: g(rel.Key)}).Exec()
	}
	s(rel.Name, fn)
}

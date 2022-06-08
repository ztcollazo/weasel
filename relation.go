package weasel

type HasMany[Doc document[Doc]] func(...Doc) SelectManyQuery[Doc]

type BelongsTo[Doc document[Doc]] func(...Doc) (Doc, error)

type HasOne[Doc document[Doc]] func(...Doc) SelectQuery[Doc]

func UseHasMany[Doc document[Doc], Rel document[Rel]](doc Doc, model *Model[Rel]) {
	g := get(doc)
	s := set(doc)
	rel := doc.model().relations["hasMany"+model.tableName]
	var fn HasMany[Rel] = func(d ...Rel) SelectManyQuery[Rel] {
		if rel.Through != "" {
			return SelectManyQuery[Rel]{
				builder: model.Conn.Builder.Select(model.tableName+".*").
					From(rel.Through).
					Where(Eq{rel.Through + rel.ForeignKey: g(rel.Key)}).
					InnerJoin(model.tableName, Eq{model.tableName + "_" + model.pk: model.pk}),
				model: *model,
			}
		} else {
			return SelectManyQuery[Rel]{
				builder: model.Conn.Builder.Select("*").
					From(model.tableName).
					Where(Eq{rel.ForeignKey: g(rel.Key)}),
				model: *model,
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
	var fn HasOne[Rel] = func(d ...Rel) SelectQuery[Rel] {
		return SelectQuery[Rel]{
			builder: model.Conn.Builder.Select("*").
				From(model.tableName).
				Where(Eq{rel.ForeignKey: g(rel.Key)}),
			model: *model,
		}
	}
	s(rel.Name, fn)
}

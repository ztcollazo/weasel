package weasel

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

# Weasel

*Note: Weasel is just now reaching a stable API. Most necessary features are available or easily integratable. You may still expect a few API changes, but the majority should be usable.*

1. [About](#about)
2. [API](#api)
3. [Documentation](#documentation)
4. [Roadmap](#roadmap)
5. [License](#license)

## About

Weasel is **the last ORM for Golang you'll ever need.** Built with Generics, so it requires at least Go 1.18+.

## API

See [the docs](https://go.dev/pkg/github.com/ztcollazo/weasel)

## Documentation

Here's an example of the API:

```go
package main

import (
  "github.com/ztcollazo/weasel" // Main package
  "github.com/ztcollazo/weasel/use" // Use package contains middleware and validations
  "github.com/lib/pq" // Or whatever db you're using
)

func main() {
  // Connect is a wrapper around SQLX's connect; see that for more details.
  conn := weasel.Connect("postgres", "user=foo dbname=bar sslmode=off")

  // Let's create the schema now
  type PersonSchema struct {
    weasel.Document[*PersonSchema] // Note the pointer!!!
    // PK denotes it as the primary key
    Id        int                            `db:"id" pk:"" type:"serial"`
    FirstName string                         `db:"first_name" type:"text"`
    LastName  string                         `db:"last_name" type:"text"`
    Email     string                         `db:"email" type:"text"`
    PlaceId   int                            `db:"place_id" type:"integer"`
    // Relations
    Place     weasel.BelongsTo[*PlaceSchema] `belongsto:"place"` // Again with the required pointer
    Hello     string
  }

  // You can define an `Init` function that is called whenever a document is created.
  // Again, the method MUST have a pointer receiver.
  func (p *PersonSchema) Init() {
    p.Hello = "world"
    // Relations, supports: BelongsTo, HasMany (and through), HasOne
    // Note: This format has been changed.
    // Old: weasel.UseBelongsTo(p, Place)
    p.Use(use.BelongsTo[*PersonSchema](Place))

    // The other relations are fairly straightforward:
    // HasOne is basically the same as BelongsTo
    // HasMany is a Group (see below)
    // HasMany through is slightly different
    // In that case, you would still use the HasMany function
    // But in the schema, you would add a `through` tag
    // with the intermittent model


    // this is also where you would do your validations
    // d.Errors is an []error
    if p.FirstName == "" {
      p.AddError(errors.New("missing first name"))
    }
    // or
    p.Use(use.ValidatePresenceOf[string /* validate presence requires data type */]("first_name"))
    // Also supports:
    // Custom: Validate(field, func(val type) bool)
    // Unique: ValidateUniquenessOf(field)
    // Unique Combination: ValidateUniqueCombination(...fields)
    // Format: ValidateFormatOf(field, regexp)
  }

  // Now for the fun part
  // Types are inferred from the second parameter; it's only there so that we can copy it
  Person := weasel.Create(conn, &PersonSchema{}, "person") // returns *Model[*PersonSchema]

  // Or you can define an init (or multiple) function also
  Person := weasel.Create(conn, &PersonSchema, "person", func (m *Model[*PersonSchema]) {
    // You can define properties on the model
    m.Set("key", "value")
    m.Get("key") //=> "value"
  })

  // Done! use it like you would Active Record
  p, _ := Person.Find(1)
  p.FirstName // 🤯 🥳
  p.Hello //=> "world"

  john, err /* error handling also */ = Person.Create(&PersonSchema{
    FirstName: "John",
    LastName: "Doe",
    Email: "john@doe.com",
    PlaceId: 1,
  })

  john.Email //=> john@doe.com

  john.Email = "johndoe@whatever.com"
  john.Save() // Pretty intuitive

  // And then when you're done
  john.Delete()

  // You can also do batch queries
  people, _ := Person.All().Where(weasel.Eq{"first_name": "John"}).Limit(3).Offset(6).Exec() // For built queries, make sure that you append exec.
  // people => []*PersonSchema{...}

  // Or specific queries
  jane := Person.FindBy("first_name", "Jane")

  // Now let's get the place
  jane.Place() //=> *PlaceSchema{...}

  // You can also check if a document is valid
  jane.IsValid() //=> true
  jane.FirstName = ""
  jane.IsValid() //=> false
  jane.IsInvalid() //=> true

  // You can add groups to group together documents with certain properties
  Person.AddGroup("FromUS", weasel.Eq{"place_id": 1})

  // And now
  Person.FromGroup("FromUS").All().Exec() // Same API as Model
  // To learn more about groups, please see below

  // You can do many other useful features such as:
  Person.Exists(1)
  Person.First() // Up to fifth
  Person.Last() // up to second
  Person.Nth(7)
  Person.NthFromLast(3)
  // To change to order of the documents, you can do:
  Person.Order("first_name DESC") // Etc.

  // You can also serialize documents:
  p, _ := Person.First()
  json, _ := Person.ToJSON()
  // Or for custom serialization:
  mp := p.ToMap() // Creates a map of all of the fields
}
```

### On the topic of Groups

Groups are an extremely valuable feature in ORMs. They allow you to combine documents with similar features, all without having to repeat your queries over and over. Groups in weasel are the foundation of not only themselves, but also models. A few rules:

1. **Groups depend on models**

   Models are what give groups the data about the table itself. This is **not** left to the groups.

2. **Models depend on groups**

   If you look in the code, you will find that `Model[Doc]` actually extends `*Group[Doc]`. This is interesting, because that means that, while models give groups all of the data, groups give models all of the functionality.

## Roadmap

- [x] Connection + multiple drivers
- [X] Query builder
- [x] Create
- [x] Read
- [X] Update
- [X] Delete
- [X] Relations
- [X] Validations
  - [X] Check valid
  - [X] Errors
  - [X] Validate Presence
  - [X] Validate Format
  - [X] Validate Custom
  - [X] Validate Uniqueness
- [X] Model Groups
- [X] Model utilities
  - [X] Find nth document
  - [X] Find nth to last document
  - [X] Check if document exists
  - [X] Count of documents
  - [X] Serialize documents
- [ ] CLI?
  - [ ] Create model files
  - [ ] Migrations?
- [ ] ~~Better config format~~ Many drivers include their own structs that you can format into an opts string for a better UX.

...and any that may come up in the future.

## License

Weasel is licensed under the MIT license. View the [LICENSE.txt](./LICENSE.txt) for more information.

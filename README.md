# Weasel

*Note: Weasel is just now reaching a stable API. Most necessary features are available or easily integratable. You may still expect a few API changes, but the majority should be usable.*

Weasel is **the last ORM for Golang you'll ever need.** Built with Generics, so it requires at least Go 1.18+.

1. [Documentation](#documentation)
2. [API](#api)
3. [Roadmap](#roadmap)
4. [License](#license)

## Documentation

See [The documentation](https://go.dev/pkg/github.com/ztcollazo/weasel)

## API

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
    // Relations, supports: BelongsTo, HasMany (through), HasOne
    weasel.UseBelongsTo(p, &Place /* pointer to model as second param */)

    // HasMany through is slightly different
    // In that case, you would still use the UseHasMany function
    // But in the schema, you would add a `through` tag
    // with the intermittent model


    // this is where you would do your validations
    // d.Errors is an []error
    if p.FirstName == "" {
      p.AddError(errors.New("missing first name"))
    }
    // or
    p.Use(use.ValidatePresenceOf[string /* required data type */]("first_name"))
    // Also supports:
    // Custom: Validates(field, func(val type) bool)
    // Unique: ValidatesUniquenessOf(field) probably not production-ready; validate on DB level instead for now
    // Format: ValidatesFormatOf(field, regexp)
  }

  // Now for the fun part
  // Types are inferred from the second parameter; it's only there so that we can copy it
  Person := weasel.Create(conn, &PersonSchema{}, person)

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
  // people => []PersonSchema{...}

  // Or specific queries
  jane := Person.FindBy("first_name", "Jane")

  // Now let's get the place
  jane.Place() //=> PlaceSchema{...}

  // You can also check if a document is valid
  jane.IsValid() //=> true
  jane.FirstName = ""
  jane.IsValid() //=> false
  jane.IsInvalid() //=> true
}
```

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

...and any that may come up in the future.

## License

Weasel is licensed under the MIT license. View the [LICENSE.txt](./LICENSE.txt) for more information.

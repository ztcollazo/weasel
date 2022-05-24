# Weasel

**WARNING :warning:: WORK IN PROGRESS** Most basic features are available. See the roadmap below.

Weasel is **the last ORM for Golang you'll ever need.** Built with Generics, so it requires at least Go 1.18+.

## API

Start by creating a connection:

```go
package main

import "github.com/ztcollazo/weasel"

func main() {
  // Connect is a wrapper around SQLX's connect; se that for more details.
  conn := weasel.Connect("postgres", "user=foo dbname=bar sslmode=off")

  // Let's create the schema now
  type PersonSchema struct {
    // PK denotes it as the primary key
    Id        int    `db:"id" pk:"" type:"integer"`
    FirstName string `db:"first_name" type:"text"`
    LastName  string `db:"last_name" type:"text"`
    Email     string `db:"email" type:"text"`
  }

  // Now for the fun part
  // Types are inferred from the second parameter; it's only there so that we can copy it
  Person := weasel.Create(conn, Person{}, person)

  // Done! use it like you would Active Record
  p, _ := Person.Find(1).FirstName
  p.FirstName // 🤯 🥳
  // Also supports: FindBy, All

  john, _ /* error handling also */ = Person.Create(PersonSchema{
    FirstName: "John",
    LastName: "Doe",
    Email: "john@doe.com",
  })

  john.Email //=> john@doe.com
}
```

## Roadmap

- [x] Connection + multiple DBs
- [X] Query builder
- [x] Create
- [x] Read
- [ ] Update
- [ ] Delete
- [ ] Validations

...and any that may come up in the future.

## License

Weasel is licensed under the MIT license. View the [LICENSE.txt](./LICENSE.txt) for more information.

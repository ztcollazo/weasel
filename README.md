# Weasel

**WARNING :warning:: WORK IN PROGRESS** Most basic features are available. See the roadmap below.

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
  "github.com/ztcollazo/weasel"
  "github.com/lib/pq" // Or whatever db you're using
)

func main() {
  // Connect is a wrapper around SQLX's connect; se that for more details.
  conn := weasel.Connect("postgres", "user=foo dbname=bar sslmode=off")

  // Let's create the schema now
  type PersonSchema struct {
    weasel.Document[*PersonSchema] // Note the pointer!!!
    // PK denotes it as the primary key
    Id        int    `db:"id" pk:"" type:"serial"`
    FirstName string `db:"first_name" type:"text"`
    LastName  string `db:"last_name" type:"text"`
    Email     string `db:"email" type:"text"`
  }

  // Now for the fun part
  // Types are inferred from the second parameter; it's only there so that we can copy it
  Person := weasel.Create(conn, &PersonSchema{}, person)

  // Done! use it like you would Active Record
  p, _ := Person.Find(1).FirstName
  p.FirstName // 🤯 🥳
  // Also supports: FindBy, All

  john, _ /* error handling also */ = Person.Create(&PersonSchema{
    FirstName: "John",
    LastName: "Doe",
    Email: "john@doe.com",
  })

  john.Email //=> john@doe.com

  john.Email = "johndoe@whatever.com"
  john.Save() // Pretty intuitive

  // And then when you're done
  john.Delete()
}
```

## Roadmap

- [x] Connection + multiple drivers
- [X] Query builder
- [x] Create
- [x] Read
- [X] Update
- [X] Delete
- [ ] Relations
- [ ] Validations

...and any that may come up in the future.

## License

Weasel is licensed under the MIT license. View the [LICENSE.txt](./LICENSE.txt) for more information.

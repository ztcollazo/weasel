package use

import (
	"fmt"
	"regexp"

	"github.com/carlmjohnson/truthy"
	"github.com/ztcollazo/weasel"
)

// ValidatePresenceOf takes a type parameter of the data type of the field and the field itself,
// and checks that the field exists in the record.
//	doc.Use(use.ValidatePresenceOf[string]("email"))
func ValidatePresenceOf[T any](field string) weasel.Middleware {
	return func(d weasel.DocumentBase) {
		if !truthy.Value(d.Get(field).(T)) {
			d.AddError(fmt.Errorf("field %s is not present in document", field))
		}
	}
}

// Validate takes a field and a function with a parameter as the value of that field returning a bool.
// It makes sure that the value satisfies the function, or else it adds an error to the model.
//	doc.Use(use.Validate("email", func (email string) bool {
// 		return email != "some@thing.com"
//	}))
func Validate[T any](field string, validator func(val T) bool) weasel.Middleware {
	return func(d weasel.DocumentBase) {
		if !validator(d.Get(field).(T)) {
			d.AddError(fmt.Errorf("field %s is not valid", field))
		}
	}
}

// ValidateUniquenessOf checks that the value for the field specified is unique in the DB.
// Do not use ValidateUniquenessOf on the Primary Key, as it will always pass.
// For example, the query would look like: SELECT COUNT(*) FROM table WHERE id = id AND id != id;
//
// Use autoincrement or validate their uniqueness internally.
//	doc.Use(use.ValidateUniquenessOf("email"))
func ValidateUniquenessOf(field string) weasel.Middleware {
	return func(d weasel.DocumentBase) {
		var count int
		err := d.Conn().Builder.Select("COUNT(*)").
			From(d.Table()).
			Where(weasel.And{weasel.Eq{field: d.Get(field)}, weasel.NotEq{d.PrimaryKey(): d.Get(d.PrimaryKey())}}).
			QueryRow().
			Scan(&count)

		if err != nil {
			d.AddError(err)
		}

		if count > 0 {
			d.AddError(fmt.Errorf("value %v for field %s is not unique", d.Get(field), field))
		}
	}
}

// ValidateFormatOf takes a regular expression and checks that the field matches the pattern.
//	doc.Use(use.ValidateFormatOf("email", regexp.MustCompile(`[^@ \t\r\n]+@[^@ \t\r\n]+\.[^@ \t\r\n]+`)))
func ValidateFormatOf(field string, format *regexp.Regexp) weasel.Middleware {
	return func(d weasel.DocumentBase) {
		if !format.MatchString(d.Get(field).(string)) {
			d.AddError(fmt.Errorf("field %s does not match the specified pattern %s", field, format))
		}
	}
}

package use

import (
	"fmt"
	"regexp"

	"github.com/carlmjohnson/truthy"
	"github.com/ztcollazo/weasel"
)

func ValidatePresenceOf[T any](field string) weasel.Middleware {
	return func(d weasel.DocumentBase) {
		if !truthy.Value(d.Get(field).(T)) {
			d.AddError(fmt.Errorf("field %s is not present in document", field))
		}
	}
}

func Validate[T any](field string, validator func(val T) bool) weasel.Middleware {
	return func(d weasel.DocumentBase) {
		if !validator(d.Get(field).(T)) {
			d.AddError(fmt.Errorf("field %s is not valid", field))
		}
	}
}

// TODO: make less computationally expensive for large dbs
// It is not recommended to use ValidateUniquenessOf in production yet
func ValidateUniquenessOf(field string) weasel.Middleware {
	return func(d weasel.DocumentBase) {
		var ids []int
		sql, args, err := d.Conn().Builder.Select(d.PrimaryKey()).From(d.Table()).Where(weasel.Eq{field: d.Get(field)}).ToSql()
		if err != nil {
			d.AddError(err)
			return
		}
		err = d.Conn().DB.Select(&ids, sql, args...)
		if err != nil {
			d.AddError(err)
			return
		}
		if len(ids) > 0 {
			if len(ids) > 1 {
				d.AddError(fmt.Errorf("value %v for field %s is not unique", d.Get(field), field))
			} else if len(ids) == 1 && ids[0] != d.Get(d.PrimaryKey()) {
				d.AddError(fmt.Errorf("value %v for field %s is not unique", d.Get(field), field))
			}
		}
	}
}

func ValidateFormatOf(field string, format *regexp.Regexp) weasel.Middleware {
	return func(d weasel.DocumentBase) {
		if !format.MatchString(d.Get(field).(string)) {
			d.AddError(fmt.Errorf("field %s does not match the specified pattern %s", field, format))
		}
	}
}

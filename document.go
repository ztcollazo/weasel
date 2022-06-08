package weasel

import (
	"errors"
	"reflect"

	"github.com/carlmjohnson/truthy"
)

type docbase interface {
	Delete() error
	Save() error
	Get(string) any
	Set(string, any)
	errors() []error
}

type document[Doc docbase] interface {
	docbase
	Create(Doc, *Model[Doc])
	model() *Model[Doc]
}

type Document[Doc document[Doc]] struct {
	document[Doc]
	Model  *Model[Doc]
	Errors []error
	get    func(string) any
	set    func(string, any)
}

// This is an internal function, exported only for use with reflect
// Do not use.
func (d *Document[Doc]) Create(doc Doc, model *Model[Doc]) {
	d.Model = model
	d.get = get(doc)
	d.set = set(doc)
}

func (d Document[Doc]) model() *Model[Doc] {
	return d.Model
}

func (d Document[Doc]) errors() []error {
	return d.Errors
}

func (d Document[Doc]) Get(name string) any {
	return d.get(name)
}

func (d Document[Doc]) Set(name string, value any) {
	d.set(name, value)
}

func (d Document[Doc]) Delete() error {
	_, err := d.Model.Conn.Builder.Delete(d.Model.tableName).Where(Eq{d.Model.pk: d.Get(d.Model.pk)}).Exec()
	return err
}

func (d Document[Doc]) Save() error {
	q := d.Model.Conn.Builder.Update(d.Model.tableName).Where(Eq{d.Model.pk: d.Get(d.Model.pk)})
	for k := range d.Model.fields {
		q = q.Set(k, d.Get(k))
	}

	callInit(d)
	if len(d.errors()) > 0 {
		return errors.New("document is invalid")
	}
	_, err := q.Exec()
	return err
}

func get[Doc document[Doc]](d Doc) func(string) any {
	v := reflect.Indirect(reflect.ValueOf(d))
	return func(name string) any {
		if truthy.Value(d.model().fields[name]) {
			return v.FieldByName(d.model().fields[name].Name).Interface()
		} else if v.IsValid() && v.CanInterface() {
			return v.FieldByName(name).Interface()
		} else {
			return nil
		}
	}
}

func set[Doc document[Doc]](d Doc) func(string, any) {
	v := reflect.Indirect(reflect.ValueOf(&d)).Elem()
	return func(name string, value any) {
		n := reflect.ValueOf(value)
		field := d.model().fields[name]
		if truthy.Value(field) {
			v.FieldByName(field.Name).Set(n)
		} else {
			v.FieldByName(name).Set(n)
		}
	}
}

func callInit[Doc docbase](d Doc, model ...*Model[Doc]) {
	v := reflect.ValueOf(d)
	if truthy.Value(model) && truthy.Value(model[0]) {
		anonymous := make([]reflect.Value, 0)
		x := v.Elem()
		for i := 0; i < x.NumField(); i++ {
			if f := x.Type().Field(i); f.Anonymous {
				anonymous = append(anonymous, x.Field(i).Addr())
			}
		}
		for _, a := range anonymous {
			if m := a.MethodByName("Create"); m.IsValid() {
				m.Call([]reflect.Value{v, reflect.ValueOf(model[0])})
			}
		}
	}
	t := v.MethodByName("Init")
	if t.IsValid() {
		t.Call([]reflect.Value{})
	} else {
		m := reflect.Indirect(v).MethodByName("Init")
		if m.IsValid() {
			m.Call([]reflect.Value{})
		}
	}
}

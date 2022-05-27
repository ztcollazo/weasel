package weasel

import (
	"errors"
	"reflect"

	"github.com/carlmjohnson/truthy"
)

type docbase interface {
	Delete() error
	Save() error
}

type document[Doc docbase] interface {
	docbase
	init(Doc, Model[Doc])
	model() Model[Doc]
}

type Document[Doc document[Doc]] struct {
	document[Doc]
	Model  Model[Doc]
	Errors []error
	Get    func(string) any
	Set    func(string, any)
}

func (d *Document[Doc]) init(doc Doc, model Model[Doc]) {
	d.Model = model
	d.Get = get(doc)
	d.Set = set(doc)
	callInit(doc)
}

func (d *Document[Doc]) model() Model[Doc] {
	return d.Model
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
	if len(d.Errors) > 0 {
		return errors.New("document is not valid")
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

func callInit[Doc docbase](d Doc) {
	v := reflect.ValueOf(d)
	t := v.MethodByName("Init")
	if t.IsValid() {
		t.Call([]reflect.Value{v})
	} else {
		t = reflect.Indirect(v).MethodByName("Init")
		if t.IsValid() {
			t.Call([]reflect.Value{v})
		}
	}
}

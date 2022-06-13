package weasel

import (
	"errors"
	"reflect"

	"github.com/carlmjohnson/truthy"
)

type DocumentBase interface {
	Delete() error
	Save() error
	Get(string) any
	Set(string, any)
	AllErrors() []error
	AddError(error)
	SetErrors([]error)
	RemoveError(int)
	Error(int) error
	PrimaryKey() string
	Init()
	IsValid() bool
	IsInvalid() bool
	Table() string
	Conn() Connection
}

type document[Doc DocumentBase] interface {
	DocumentBase
	Create(Doc, *Model[Doc])
	Use(Middleware)
	model() *Model[Doc]
}

type Document[Doc document[Doc]] struct {
	document[Doc]
	Model  *Model[Doc]
	Errors []error
	get    func(string) any
	set    func(string, any)
	use    func(Middleware)
}

type Middleware func(DocumentBase)

// This is an internal function, exported only for use with reflect
// Do not use.
func (d *Document[Doc]) Create(doc Doc, model *Model[Doc]) {
	d.Errors = []error{}
	d.Model = model
	d.get = get(doc)
	d.set = set(doc)
	d.use = use(doc)
}

func (d *Document[Doc]) Use(m Middleware) {
	d.use(m)
}

// You can define a custom Init function to run on document creation.
func (d Document[Doc]) Init() {}

func (d Document[Doc]) model() *Model[Doc] {
	return d.Model
}

func (d Document[Doc]) AllErrors() []error {
	return d.Errors
}

func (d *Document[Doc]) AddError(es error) {
	d.Errors = append(d.Errors, es)
}

func (d *Document[Doc]) SetErrors(errs []error) {
	d.Errors = errs
}

func (d *Document[Doc]) RemoveError(id int) {
	d.Errors = append(d.Errors[:id], d.Errors[id+1:]...)
}

func (d Document[Doc]) Error(id int) error {
	return d.Errors[id]
}

func (d Document[Doc]) PrimaryKey() string {
	return d.Model.pk
}

func (d Document[Doc]) Table() string {
	return d.Model.tableName
}

func (d Document[Doc]) Conn() Connection {
	return d.Model.Conn
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
	callInit(&d)
	if len(d.Errors) > 0 {
		return errors.New("document is invalid")
	}

	q := d.Model.Conn.Builder.Update(d.Model.tableName).Where(Eq{d.Model.pk: d.Get(d.Model.pk)})
	for k := range d.Model.fields {
		q = q.Set(k, d.Get(k))
	}
	_, err := q.Exec()
	return err
}

func (d Document[Doc]) IsValid() bool {
	callInit(&d)
	return len(d.Errors) <= 0
}

func (d Document[Doc]) IsInvalid() bool {
	return !d.IsValid()
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

func use[Doc document[Doc]](doc Doc) func(Middleware) {
	return func(m Middleware) {
		m(doc)
	}
}

func callInit[Doc DocumentBase](d Doc, model ...*Model[Doc]) {
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
	d.Init()
}

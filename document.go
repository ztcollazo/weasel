package weasel

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/carlmjohnson/truthy"
)

// DocumentBase provides an interface to be used for times when you may not know the schema
// type. All documents conform to it; it is used as the constraint for type parameters. You
// may find yourself using DocumentBase if you want to write custom validation and middleware.
type DocumentBase interface {
	Delete() error
	Save() error
	ToJSON() (string, error)
	ToMap() map[string]any
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
	Use(Middleware)
}

type document[Doc DocumentBase] interface {
	DocumentBase
	Create(Doc, *Model[Doc])
	GetModel() *Model[Doc]
}

// Document provides a struct to extend your schemas. It contains errors and model information and
// extends DocumentBase. You typically will not have to use Document except for in defining your schema,
// for example:
//	type PersonSchema struct {
//		weasel.Document[*PersonSchema]
//	}
type Document[Doc document[Doc]] struct {
	document[Doc]
	Model  *Model[Doc]
	Errors []error
	get    func(string) any
	set    func(string, any)
	use    func(Middleware)
}

// Middleware is a type that all middleware (passed to the Use function) should be/return.
// it takes an argument in DocumentBase and can use `Get` and `Set` to get and change properties.
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

// ToJSON returns a JSON string of all of the document's fields.
// This is useful for serialization in HTTP responses.
func (d Document[Doc]) ToJSON() (string, error) {
	mp := d.ToMap()
	b, err := json.Marshal(mp)
	return string(b), err
}

// ToMap returns a map of the document.
// This is useful for custom serialization.
func (d Document[Doc]) ToMap() map[string]any {
	mp := make(map[string]any)
	for f := range d.Model.fields {
		mp[f] = d.Get(f)
	}
	return mp
}

// Use takes middleware and runs it on document creation.
func (d *Document[Doc]) Use(m Middleware) {
	d.use(m)
}

// You can define a custom Init function to run on document creation.
func (d Document[Doc]) Init() {}

func (d Document[Doc]) GetModel() *Model[Doc] {
	return d.Model
}

// AllErrors returns all of the document's errors.
func (d Document[Doc]) AllErrors() []error {
	return d.Errors
}

// You can use AddError to append an error to the list. This is very useful in middleware.
func (d *Document[Doc]) AddError(es error) {
	d.Errors = append(d.Errors, es)
}

// SetErrors completely sets all of the errors. It is not recommended to use unless you
// really need to or want to reset all of the errors.
func (d *Document[Doc]) SetErrors(errs []error) {
	d.Errors = errs
}

// RemoveError takes the index of an error and removes it from the list.
func (d *Document[Doc]) RemoveError(id int) {
	d.Errors = append(d.Errors[:id], d.Errors[id+1:]...)
}

// Error returns the error at the given ID.
func (d Document[Doc]) Error(id int) error {
	return d.Errors[id]
}

// PrimaryKey returns the table's primary key. This is useful for middleware.
func (d Document[Doc]) PrimaryKey() string {
	return d.Model.pk
}

// Table returns the table name of the document. This is useful for middleware.
func (d Document[Doc]) Table() string {
	return d.Model.tableName
}

// Conn returns the current connection. See the connection docs.
func (d Document[Doc]) Conn() Connection {
	return d.Model.Conn
}

// Get returns a property (DB or struct) on the document. You may need to use type assertion.
func (d Document[Doc]) Get(name string) any {
	return d.get(name)
}

// Set sets a property on the document (DB or struct).
func (d Document[Doc]) Set(name string, value any) {
	d.set(name, value)
}

// Delete completely removes the document from the database.
func (d Document[Doc]) Delete() error {
	_, err := d.Model.Conn.Builder.Delete(d.Model.tableName).Where(Eq{d.Model.pk: d.Get(d.Model.pk)}).Exec()
	return err
}

// Save saves the document's changes, changed either by Set or manually.
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

// IsValid checks that the document does not contain any errors.
func (d Document[Doc]) IsValid() bool {
	callInit(&d)
	return len(d.Errors) <= 0
}

// IsInvalid checks if the document has any errors.
func (d Document[Doc]) IsInvalid() bool {
	return !d.IsValid()
}

func get[Doc document[Doc]](d Doc) func(string) any {
	v := reflect.Indirect(reflect.ValueOf(d))
	return func(name string) any {
		if truthy.Value(d.GetModel().fields[name]) {
			return v.FieldByName(d.GetModel().fields[name].Name).Interface()
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
		field := d.GetModel().fields[name]
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

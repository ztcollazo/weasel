package weasel_test

import (
	"errors"
	"testing"

	"github.com/carlmjohnson/truthy"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/ztcollazo/weasel"
)

var schema = `
DROP TABLE IF EXISTS person;
DROP TABLE IF EXISTS place;

CREATE TABLE person (
		id serial primary key,
    first_name text,
    last_name text,
    email text,
		place_id integer
);

CREATE TABLE place (
		id serial primary key,
    country text,
    city text NULL,
    telcode integer
);

INSERT INTO person (first_name, last_name, email, place_id) VALUES ('John', 'Doe', 'john@doe.com', 1);
INSERT INTO person (first_name, last_name, email, place_id) VALUES ('Jane', 'Doe', 'jane@doe.net', 1);
INSERT INTO place (country, city, telcode) VALUES ('United States of America', 'Chicago', 1);`

type PersonSchema struct {
	weasel.Document[*PersonSchema]
	Id        int                            `db:"id" pk:"" type:"serial"`
	FirstName string                         `db:"first_name" type:"text"`
	LastName  string                         `db:"last_name" type:"text"`
	Email     string                         `db:"email" type:"text"`
	PlaceId   int                            `db:"place_id" type:"integer"`
	Place     weasel.BelongsTo[*PlaceSchema] `belongsto:"place" fk:"id" key:"place_id"`
	Hello     string
}

type PlaceSchema struct {
	weasel.Document[*PlaceSchema]
	Id      int                           `db:"id" pk:"" type:"serial"`
	Country string                        `db:"country" type:"text"`
	City    string                        `db:"city" type:"text"`
	Telcode int                           `db:"telcode" type:"integer"`
	People  weasel.HasMany[*PersonSchema] `hasmany:"person" fk:"place_id" key:"id"`
}

var conn = weasel.Connect("postgres", "user=ztcollazo dbname=postgres sslmode=disable")

var Place = weasel.Create(conn, &PlaceSchema{}, "place")

var Person = weasel.Create(conn, &PersonSchema{}, "person")

func (p *PersonSchema) Init() {
	p.Hello = "world"
	weasel.UseBelongsTo(p, &Place)
	if !truthy.Value(p.Email) {
		p.Errors = append(p.Errors, errors.New("missing email"))
	}
}

func (p *PlaceSchema) Init() {
	weasel.UseHasMany(p, &Person)
}

type WeaselTestSuite struct {
	suite.Suite
	assert *assert.Assertions
}

func (s *WeaselTestSuite) SetupTest() {
	s.assert = assert.New(s.T())
	conn.DB.MustExec(schema)
}

func (s *WeaselTestSuite) TestInsert() {
	p, err := Person.Create(&PersonSchema{
		FirstName: "Zachary",
		LastName:  "Collazo",
		Email:     "ztcollazo08@gmail.com",
		PlaceId:   1,
	})
	s.assert.Nil(err)
	s.assert.Equal("ztcollazo08@gmail.com", p.Email)
	s.assert.Equal("Zachary", p.FirstName)
	s.assert.Equal("Collazo", p.LastName)
}

func (s *WeaselTestSuite) TestFind() {
	p, err := Person.Find(1)
	s.assert.Nil(err)
	s.assert.Equal("john@doe.com", p.Email)
	s.assert.Equal("John", p.FirstName)
	s.assert.Equal("Doe", p.LastName)
}

func (s *WeaselTestSuite) TestFindBy() {
	p, err := Person.FindBy("first_name", "John")
	s.assert.Nil(err)
	s.assert.Equal(1, p.Id)
	s.assert.Equal("john@doe.com", p.Email)
	s.assert.Equal("John", p.FirstName)
	s.assert.Equal("Doe", p.LastName)
}

func (s *WeaselTestSuite) TestAll() {
	p, err := Person.All().Exec()
	s.assert.Nil(err)
	s.assert.GreaterOrEqual(len(p), 2)
}

func (s *WeaselTestSuite) TestGetSet() {
	p, err := Person.Find(1)
	s.assert.Nil(err)

	s.assert.Equal("John", p.Get("first_name"))
	p.Set("first_name", "Pizza")
	s.assert.Equal("Pizza", p.FirstName)
}

func (s *WeaselTestSuite) TestSave() {
	p, err := Person.Find(1)
	s.assert.Nil(err)

	p.FirstName = "Pizza"
	err = p.Save()
	s.assert.Nil(err)
	s.assert.Equal("Pizza", p.FirstName)
}

func (s *WeaselTestSuite) TestDelete() {
	p, err := Person.Create(&PersonSchema{
		FirstName: "Somebody",
		LastName:  "Else",
		Email:     "somebodyelse@whatever.com",
		PlaceId:   1,
	})
	s.assert.Nil(err)
	err = p.Delete()
	s.assert.Nil(err)
}

func (s *WeaselTestSuite) TestInit() {
	p, err := Person.Find(1)
	s.assert.Nil(err)
	s.assert.Equal("world", p.Hello)
}

func (s *WeaselTestSuite) TestBelongsTo() {
	p, err := Person.Find(1)
	s.assert.Nil(err)

	place, err := Place.Find(1)
	s.assert.Nil(err)

	t, err := p.Place()
	s.assert.Nil(err)

	s.assert.Equal(place, t)
}

func (s *WeaselTestSuite) TestHasMany() {
	p, err := Place.Find(1)
	s.assert.Nil(err)

	person, err := Person.Find(1)
	s.assert.Nil(err)

	t, err := p.People().Where(weasel.Eq{"id": p.Id}).Exec()
	s.assert.Nil(err)

	s.assert.Equal(person.FirstName, t[0].FirstName)
	s.assert.Equal(person.LastName, t[0].LastName)
	s.assert.Equal(person.Id, t[0].Id)
	s.assert.Equal(person.Email, t[0].Email)
}

func (s *WeaselTestSuite) TestInvalidDoc() {
	p, err := Person.Create(&PersonSchema{
		FirstName: "Hello",
		LastName:  "World",
		PlaceId:   1,
	})

	s.assert.Equal(errors.New("document is invalid"), err)
	s.assert.True(contains(p.Errors, errors.New("missing email")))
}

func TestWeasel(t *testing.T) {
	suite.Run(t, new(WeaselTestSuite))
}

func contains(s []error, str error) bool {
	for _, v := range s {
		if v.Error() == str.Error() {
			return true
		}
	}

	return false
}

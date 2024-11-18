package weasel_test

import (
	"encoding/json"
	"errors"
	"regexp"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/ztcollazo/weasel"
	"github.com/ztcollazo/weasel/use"
)

var schema = `
DROP TABLE IF EXISTS person;
DROP TABLE IF EXISTS friends;
DROP TABLE IF EXISTS place;

CREATE TABLE person (
		id serial primary key,
    first_name text,
    last_name text,
    email text,
		place_id integer
);

CREATE TABLE friends (
	id serial primary key,
	friender integer,
	friended integer
);

CREATE TABLE place (
		id serial primary key,
    country text,
    city text NULL,
    telcode integer
);

INSERT INTO person (first_name, last_name, email, place_id) VALUES ('John', 'Doe', 'john@doe.com', 1);
INSERT INTO person (first_name, last_name, email, place_id) VALUES ('Jane', 'Doe', 'jane@doe.net', 1);

INSERT INTO friends (friender, friended) VALUES (1, 2);

INSERT INTO place (country, city, telcode) VALUES ('United States of America', 'Chicago', 1);`

type PersonSchema struct {
	weasel.Document[*PersonSchema]
	Id        int                            `db:"id" pk:"" type:"serial"`
	FirstName string                         `db:"first_name" type:"text"`
	LastName  string                         `db:"last_name" type:"text"`
	Email     string                         `db:"email" type:"text"`
	PlaceId   int                            `db:"place_id" type:"integer"`
	Place     weasel.BelongsTo[*PlaceSchema] `belongsto:"place" fk:"id" key:"place_id"`
	Friends   weasel.HasMany[*PersonSchema]  `hasmany:"person" through:"friends" key:"friender" fk:"friended"`
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

var Person = weasel.Create(conn, &PersonSchema{}, "person", func(m *weasel.Model[*PersonSchema]) {
	m.Set("hello", "world")
})

func (p *PersonSchema) Init() {
	p.Hello = "world"
	// Deprecated:
	//
	//	weasel.UseBelongsTo(p, Place)
	//	weasel.UseHasMany(p, Person)
	p.Use(use.HasMany[*PersonSchema](Person))
	p.Use(use.BelongsTo[*PersonSchema](Place))
	p.Use(use.ValidatePresenceOf[string]("email"))
	p.Use(use.ValidateFormatOf("email", regexp.MustCompile(`[^@ \t\r\n]+@[^@ \t\r\n]+\.[^@ \t\r\n]+`)))
	p.Use(use.ValidateUniquenessOf("email"))
	p.Use(use.ValidateUniqueCombination("first_name", "last_name"))
	p.Use(use.Validate("email", func(val string) bool {
		return val != "random@email.com"
	}))
}

func (p *PlaceSchema) Init() {
	// Deprecated:
	//
	//	weasel.UseHasMany(p, Person)
	p.Use(use.HasMany[*PlaceSchema](Person))
}

type WeaselTestSuite struct {
	suite.Suite
	assert *assert.Assertions
}

func (s *WeaselTestSuite) SetupTest() {
	s.assert = assert.New(s.T())
	conn.DB.MustExec(schema)
	Person.CreateGroup("FromUS", weasel.Eq{"place_id": 1}) // Basically the same as BelongsTo now, just a different format
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

	s.assert.Equal(place.Id, t.Id)
}

func (s *WeaselTestSuite) TestHasMany() {
	p, err := Place.Find(1)
	s.assert.Nil(err)

	person, err := Person.Find(1)
	s.assert.Nil(err)

	t, err := p.People().All().Where(weasel.Eq{"id": p.Id}).Exec()
	s.assert.Nil(err)

	s.assert.Equal(person.FirstName, t[0].FirstName)
	s.assert.Equal(person.LastName, t[0].LastName)
	s.assert.Equal(person.Id, t[0].Id)
	s.assert.Equal(person.Email, t[0].Email)
}

func (s *WeaselTestSuite) TestHasManyThrough() {
	one, err := Person.Find(1)
	s.assert.Nil(err)
	two, err := Person.Find(2)
	s.assert.Nil(err)
	friend, err := one.Friends().Find(2)
	s.assert.Nil(err)
	s.assert.Equal(two.Id, friend.Id)
}

func (s *WeaselTestSuite) TestNth() {
	one, err := Person.Find(1)
	s.assert.Nil(err)
	doc, err := Person.Nth(1)
	s.assert.Nil(err)
	s.assert.Equal(one.Id, doc.Id)
}

func (s *WeaselTestSuite) TestNthToLast() {
	last, err := Place.Find(1)
	s.assert.Nil(err)
	doc, err := Place.NthToLast(1)
	s.assert.Nil(err)
	s.assert.Equal(last.Id, doc.Id)
}

func (s *WeaselTestSuite) TestCount() {
	count, err := Place.Count()
	s.assert.Nil(err)
	s.assert.Equal(1, count)
}

func (s *WeaselTestSuite) TestExists() {
	ex, err := Person.Exists(2)
	s.assert.Nil(err)
	s.assert.True(ex)

	p, err := Place.Exists(2)
	s.assert.Nil(err)
	s.assert.False(p)
}

func (s *WeaselTestSuite) TestModelProps() {
	s.assert.Equal("world", Person.Get("hello"))
}

func (s *WeaselTestSuite) TestValidatePresence() {
	p, err := Person.Create(&PersonSchema{
		FirstName: "Hello",
		LastName:  "World",
		PlaceId:   1,
	})

	s.assert.Equal(errors.New("document is invalid"), err)
	s.assert.True(contains(p.Errors, errors.New("field email is not present in document")))
	s.assert.False(p.IsValid())
	s.assert.True(p.IsInvalid())
}

func (s *WeaselTestSuite) TestValidateFormat() {
	p, err := Person.Create(&PersonSchema{
		FirstName: "Hello",
		LastName:  "World",
		Email:     "bob",
		PlaceId:   1,
	})

	s.assert.Equal(errors.New("document is invalid"), err)
	s.assert.True(contains(p.Errors, errors.New(`field email does not match the specified pattern [^@ \t\r\n]+@[^@ \t\r\n]+\.[^@ \t\r\n]+`)))
	s.assert.False(p.IsValid())
	s.assert.True(p.IsInvalid())
}

func (s *WeaselTestSuite) TestValidateUniqueness() {
	p, err := Person.Create(&PersonSchema{
		FirstName: "Hello",
		LastName:  "World",
		Email:     "john@doe.com",
		PlaceId:   1,
	})

	s.assert.Equal(errors.New("document is invalid"), err)
	s.assert.True(contains(p.Errors, errors.New("value john@doe.com for field email is not unique")))
	s.assert.False(p.IsValid())
	s.assert.True(p.IsInvalid())
}

func (s *WeaselTestSuite) TestValidateUniqueCombination() {
	p, err := Person.Create(&PersonSchema{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "johndoe@example.com",
		PlaceId:   1,
	})

	s.assert.Equal(errors.New("document is invalid"), err)
	s.assert.True(contains(p.Errors, errors.New("combination of values [John Doe] for fields [first_name last_name] is not unique")))
	s.assert.False(p.IsValid())
	s.assert.True(p.IsInvalid())
}

func (s *WeaselTestSuite) TestValidateCustom() {
	p, err := Person.Create(&PersonSchema{
		FirstName: "Hello",
		LastName:  "World",
		Email:     "random@email.com",
		PlaceId:   1,
	})

	s.assert.Equal(errors.New("document is invalid"), err)
	s.assert.True(contains(p.Errors, errors.New("field email is not valid")))
	s.assert.False(p.IsValid())
	s.assert.True(p.IsInvalid())
}

func (s *WeaselTestSuite) TestGroup() {
	p, err := Person.FromGroup("FromUS").Find(1)

	s.assert.Nil(err)
	s.assert.Equal("John", p.FirstName)
}

func (s *WeaselTestSuite) TestJSON() {
	p, err := Person.First()
	s.assert.Nil(err)

	j, err := p.ToJSON()
	s.assert.Nil(err)

	mp := make(map[string]any)
	err = json.Unmarshal([]byte(j), &mp)
	s.assert.Nil(err)
	s.assert.Equal(1, int(mp["id"].(float64)))
	s.assert.Equal("John", mp["first_name"])
}

func (s *WeaselTestSuite) TestMap() {
	p, err := Person.First()
	s.assert.Nil(err)

	m := p.ToMap()
	s.assert.Equal(1, m["id"])
	s.assert.Equal("John", m["first_name"])
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

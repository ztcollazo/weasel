package weasel_test

import (
	"testing"

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
    email text
);

CREATE TABLE place (
		id serial primary key,
    country text,
    city text NULL,
    telcode integer
);

INSERT INTO person (first_name, last_name, email) VALUES ('John', 'Doe', 'john@doe.com');
INSERT INTO person (first_name, last_name, email) VALUES ('Jane', 'Doe', 'jane@doe.net');`

type Person struct {
	weasel.Document[*Person]
	Id        int    `db:"id" pk:"" type:"serial"`
	FirstName string `db:"first_name" type:"text"`
	LastName  string `db:"last_name" type:"text"`
	Email     string `db:"email" type:"text"`
	Hello     string
}

func (p Person) Init(d *Person) {
	d.Hello = "world"
}

type WeaselTestSuite struct {
	suite.Suite
	assert *assert.Assertions
	conn   weasel.Connection
	model  weasel.Model[*Person]
}

func (s *WeaselTestSuite) SetupTest() {
	s.assert = assert.New(s.T())
	s.conn = weasel.Connect("postgres", "user=ztcollazo dbname=postgres sslmode=disable")
	s.conn.DB.MustExec(schema)
	s.model = weasel.Create(s.conn, &Person{}, "person")
}

func (s *WeaselTestSuite) TestInsert() {
	p, err := s.model.Create(&Person{
		FirstName: "Zachary",
		LastName:  "Collazo",
		Email:     "ztcollazo08@gmail.com",
	})
	s.assert.Nil(err)
	s.assert.Equal("ztcollazo08@gmail.com", p.Email)
	s.assert.Equal("Zachary", p.FirstName)
	s.assert.Equal("Collazo", p.LastName)
}

func (s *WeaselTestSuite) TestFind() {
	p, err := s.model.Find(1)
	s.assert.Nil(err)
	s.assert.Equal("john@doe.com", p.Email)
	s.assert.Equal("John", p.FirstName)
	s.assert.Equal("Doe", p.LastName)
}

func (s *WeaselTestSuite) TestFindBy() {
	p, err := s.model.FindBy("first_name", "John")
	s.assert.Nil(err)
	s.assert.Equal(1, p.Id)
	s.assert.Equal("john@doe.com", p.Email)
	s.assert.Equal("John", p.FirstName)
	s.assert.Equal("Doe", p.LastName)
}

func (s *WeaselTestSuite) TestAll() {
	p, err := s.model.All().Exec()
	s.assert.Nil(err)
	s.assert.GreaterOrEqual(len(p), 2)
}

func (s *WeaselTestSuite) TestGetSet() {
	p, err := s.model.Find(1)
	s.assert.Nil(err)

	s.assert.Equal("John", p.Get("first_name"))
	p.Set("first_name", "Pizza")
	s.assert.Equal("Pizza", p.FirstName)
}

func (s *WeaselTestSuite) TestSave() {
	p, err := s.model.Find(1)
	s.assert.Nil(err)

	p.FirstName = "Pizza"
	err = p.Save()
	s.assert.Nil(err)
	s.assert.Equal("Pizza", p.FirstName)
}

func (s *WeaselTestSuite) TestDelete() {
	p, err := s.model.Create(&Person{
		FirstName: "Somebody",
		LastName:  "Else",
		Email:     "somebodyelse@whatever.com",
	})
	s.assert.Nil(err)
	err = p.Delete()
	s.assert.Nil(err)
}

func (s *WeaselTestSuite) TestInit() {
	p, err := s.model.Find(1)
	s.assert.Nil(err)
	s.assert.Equal("world", p.Hello)
}

func TestWeasel(t *testing.T) {
	suite.Run(t, new(WeaselTestSuite))
}

package main

import (
	"io/ioutil"
	"testing"

	"github.com/alexbostock/alder/database"
	"github.com/alexbostock/alder/schema"
)

func TestQueries(t *testing.T) {
	schemaFile, err := ioutil.ReadFile("./test.yaml")
	if err != nil {
		t.Error("Failed to load schema")
	}

	schema := schema.New(schemaFile)
	db := database.New(4, schema)

	queries := []string{
		"insert into user (forename, surname, address) values ('Alex', 'Bostock', 'nope')",
		"insert into user (forename, surname, address) values ('Alex', 'Horne', 'nope')",
		"insert into user (forename, surname, address) values ('Alex', 'Armstrong', 'nope')",
		"select * from user",
	}

	for _, query := range queries {
		db.Query(query)
	}
}

package schema

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestSchemaParse(t *testing.T) {
	schemaFile, err := ioutil.ReadFile("../test.yaml")
	if err != nil {
		t.Error(err.Error())
	}

	schema := New(schemaFile)

	expected := Schema{
		Tables: []Table{
			Table{
				Name: "order",
				Fields: []Field{
					Field{
						Name: "id",
						Type: PrimaryKey,
					},
					Field{
						Name: "items",
						Type: String,
					},
					Field{
						Name: "price",
						Type: Int,
					},
				},
			},
			Table{
				Name: "user",
				Fields: []Field{
					Field{
						Name: "id",
						Type: PrimaryKey,
					},
					Field{
						Name: "forename",
						Type: String,
					},
					Field{
						Name: "surname",
						Type: String,
					},
					Field{
						Name: "address",
						Type: String,
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(schema, expected) {
		spew.Dump(schema)
		t.Error("Schema parsed incorrectly from yaml.")
	}
}

package schema

import (
	"errors"

	"github.com/go-yaml/yaml"
)

type Datatype int

const (
	Int Datatype = iota
	String
	PrimaryKey // Primary key is always sequentially-allocated integer
)

type untypedField struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Field struct {
	Name string
	Type Datatype
}

type untypedTable struct {
	Name       string         `yaml:"name"`
	PrimaryKey string         `yaml:"key"`
	Fields     []untypedField `yaml:"fields"`
}

type Table struct {
	Name   string
	Fields []Field
}

type untypedSchema struct {
	Tables []untypedTable `yaml:"tables"`
}

type Schema struct {
	Tables []Table
}

func (ut untypedTable) typeCheck() Table {
	tab := Table{
		Name:   ut.Name,
		Fields: []Field{Field{Name: ut.PrimaryKey, Type: PrimaryKey}},
	}

	for _, f := range ut.Fields {
		var t Datatype
		switch f.Type {
		case "int":
			t = Int
		case "string":
			t = String
		default:
			panic(errors.New("Unexpected field type"))
		}

		field := Field{
			f.Name,
			t,
		}

		tab.Fields = append(tab.Fields, field)
	}

	return tab
}

func (us untypedSchema) typeCheck() Schema {
	s := Schema{make([]Table, 0, len(us.Tables))}

	for _, t := range us.Tables {
		s.Tables = append(s.Tables, t.typeCheck())
	}

	return s
}

func New(file []byte) Schema {
	untyped := untypedSchema{}
	yaml.UnmarshalStrict(file, &untyped)
	return untyped.typeCheck()
}

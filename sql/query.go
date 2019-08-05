package sql

import "github.com/alexbostock/alder/sql/parser"

// A Query is a semantic representation of a type-safe query. It should be
// instantiated by Compile.
type Query interface {
	Execute()
}

type compoundQuery struct {
	query1    selectQuery
	operation parser.Nonterminal
	query2    Query
}

func (q compoundQuery) Execute() {
}

type selectQuery struct {
	keys    []string // keys == nil => select * (universal set of keys)
	table   string
	filters []filter
}

func (q selectQuery) Execute() {
}

type val struct {
	isNum bool // true iff the value is an int (so false => value is a string)
	num   int
	str   string
}

type insertQuery struct {
	keys   []string
	values [][]val
	table  string
}

func (q insertQuery) Execute() {
}

type updateQuery struct {
	values map[string]val
	table  string
	where  whereClause
}

func (q updateQuery) Execute() {
}

type deleteQuery struct {
	table string
	where whereClause
}

func (q deleteQuery) Execute() {
}

type filter struct {
	// TODO: implement this
}

type whereClause struct {
	// TODO: implement this
}

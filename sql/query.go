package sql

import "github.com/alexbostock/alder/sql/parser"

// A Query is a semantic representation of a type-safe query. It should be
// instantiated by Compile.
type Query interface{}

type CompoundQuery struct {
	query1    SelectQuery
	operation parser.Nonterminal
	query2    Query
}

type SelectQuery struct {
	Keys    []string // keys == nil => select * (universal set of keys)
	Table   string
	Filters []Filter
}

type Val struct {
	IsNum bool // true iff the value is an int (so false => value is a string)
	Num   int
	Str   string
}

type InsertQuery struct {
	Keys   []string
	Values [][]Val
	Table  string
}

type UpdateQuery struct {
	Values map[string]Val
	Table  string
	Where  WhereClause
}

type DeleteQuery struct {
	Table string
	Where WhereClause
}

type Filter struct {
	// TODO: implement this
}

type WhereClause struct {
	// TODO: implement this
}

// Package database provides a complete database, exposing an interface to
// execute SQL queries.
package database

import (
	"fmt"

	"github.com/alexbostock/alder/schema"
	"github.com/alexbostock/alder/sql"
	"github.com/alexbostock/alder/store"
)

type Db struct {
	schema        schema.Schema
	tables        map[string]store.Store
	cachedQueries map[string]sql.Query
}

func New(branchingFactor int, schema schema.Schema) Db {
	db := Db{
		schema:        schema,
		tables:        make(map[string]store.Store),
		cachedQueries: make(map[string]sql.Query),
	}

	for _, table := range schema.Tables {
		db.tables[table.Name] = store.NewBPTree(branchingFactor)
	}

	return db
}

func (db Db) Query(q string) {
	query, ok := db.cachedQueries[q]
	if !ok {
		query = sql.Compile(db.schema, q)
		db.cachedQueries[q] = query
	}

	fmt.Println("Query execution not yet implemented.")
}

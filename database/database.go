// Package database provides a complete database, exposing an interface to
// execute SQL queries.
package database

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/alexbostock/alder/schema"
	"github.com/alexbostock/alder/sql"
	"github.com/alexbostock/alder/store"
	"github.com/davecgh/go-spew/spew"
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

	db.execute(query)
}

func (db Db) execute(q sql.Query) {
	switch q.(type) {
	case sql.CompoundQuery:
		db.compoundQuery(q.(sql.CompoundQuery))
	case sql.SelectQuery:
		spew.Dump(db.selectQuery(q.(sql.SelectQuery)))
	case sql.InsertQuery:
		db.insertQuery(q.(sql.InsertQuery))
	case sql.UpdateQuery:
		db.updateQuery(q.(sql.UpdateQuery))
	case sql.DeleteQuery:
		db.deleteQuery(q.(sql.DeleteQuery))
	default:
		panic(errors.New("Invalid query tree (which should not have passed static analysis"))
	}
}

func (db Db) compoundQuery(q sql.CompoundQuery) {
	fmt.Println("Not yet implemented")
}

func (db Db) selectQuery(q sql.SelectQuery) []map[string]sql.Val {
	store := db.tables[q.Table]
	primaryKey := db.schema.GetTable(q.Table).GetPrimaryKey()

	_ = primaryKey

	// Filters such as WHERE are not yet implemented, so return all records
	data := deserialisePlural(store.GetAllWhere(func(int, []byte) bool {
		return true
	}), primaryKey)

	// SELECT * FROM table
	if len(q.Keys) == 0 {
		return data
	}

	keys := make(map[string]bool)
	for _, key := range q.Keys {
		keys[key] = true
	}

	for _, record := range data {
		for key := range record {
			if !keys[key] {
				delete(record, key)
			}
		}
	}

	return data
}

func (db Db) insertQuery(q sql.InsertQuery) {
	fmt.Println("Not yet implemented")
}

func (db Db) updateQuery(q sql.UpdateQuery) {
	fmt.Println("Not yet implemented")
}

func (db Db) deleteQuery(q sql.DeleteQuery) {
	fmt.Println("Not yet implemented")
}

func serialise(data map[string]sql.Val) []byte {
	serialised := new(bytes.Buffer)
	e := gob.NewEncoder(serialised)

	e.Encode(data)

	return serialised.Bytes()
}

func deserialise(data []byte) map[string]sql.Val {
	serialised := bytes.NewReader(data)
	d := gob.NewDecoder(serialised)
	deserialised := make(map[string]sql.Val)
	err := d.Decode(&deserialised)
	if err != nil {
		panic(err)
	}

	return deserialised
}

func deserialisePlural(data map[int][]byte, primary string) []map[string]sql.Val {
	result := make([]map[string]sql.Val, 0, len(data))

	for primaryKey, val := range data {
		deserialised := deserialise(val)
		deserialised[primary] = sql.Val{true, primaryKey, ""}

		result = append(result, deserialised)
	}

	return result
}

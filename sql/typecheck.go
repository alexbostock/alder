package sql

import (
	"errors"
	"strconv"

	"github.com/alexbostock/alder/schema"
	"github.com/alexbostock/alder/sql/parser"
	"github.com/davecgh/go-spew/spew"
)

// Compile, given an SQL query as a string and a database schema, compiles the
// query to a Query object, which can be executed.
func Compile(s schema.Schema, query string) Query {
	schemaMap := make(map[string]map[string]schema.Datatype) // table -> key -> type

	for _, tab := range s.Tables {
		schemaMap[tab.Name] = make(map[string]schema.Datatype)

		for _, field := range tab.Fields {
			schemaMap[tab.Name][field.Name] = field.Type
		}
	}

	parser := parser.New(query)

	q := check(schemaMap, parser.Parse())
	spew.Dump(q)
	return q
}

func check(s map[string]map[string]schema.Datatype, query *parser.Node) Query {
	switch query.T {
	case parser.SelectFrom:
		sq := &selectQuery{
			keys:    checkKeyList(s, query.Args[0]),
			table:   checkTable(s, query.Args[1]),
			filters: checkFilters(s, query.Args[2]),
		}

		err := checkSelectTypes(s, sq)
		if err != nil {
			panic(err)
		}

		return sq
	case parser.InsertInto:
		is := &insertQuery{
			keys:   checkKeyList(s, query.Args[0]),
			values: checkValuesList(s, query.Args[2]),
			table:  checkTable(s, query.Args[1]),
		}

		err := checkInsertTypes(s, is)
		if err != nil {
			panic(err)
		}

		return is
	case parser.UpdateSet:
		us := &updateQuery{
			values: checkAssignments(query.Args[1]),
			table:  checkTable(s, query.Args[0]),
			where:  checkWhereClause(s, query.Args[2]),
		}

		err := checkUpdateTypes(s, us)
		if err != nil {
			panic(err)
		}

		return us
	case parser.DeleteFrom:
		ds := &deleteQuery{
			table: checkTable(s, query.Args[0]),
			where: checkWhereClause(s, query.Args[1].Args[0]),
		}

		// TODO: type-check where clause of ds

		return ds
	case parser.UnionOf:
		fallthrough
	case parser.IntersectionOf:
		fallthrough
	case parser.DifferenceOf:
		return &compoundQuery{
			query1:    check(s, query.Args[0]).(selectQuery),
			operation: query.T,
			query2:    check(s, query.Args[1]),
		}
	case parser.LiteralList, parser.Key, parser.Keys, parser.Literals,
		parser.ValueList, parser.Assignment, parser.AssignmentList,
		parser.Filters, parser.WhereExpr, parser.InnerJoin,
		parser.OuterJoin, parser.LeftJoin, parser.RightJoin,
		parser.OrderBy, parser.Literal, parser.Table:
		panic(errors.New("Invalid parse tree"))
	default:
		panic(errors.New("Semantic error: unexpected node type"))
	}

	return selectQuery{}
}

func checkKeyList(s map[string]map[string]schema.Datatype, kl *parser.Node) []string {
	spew.Dump(kl)
	if kl.T == parser.Keys {
		if len(kl.Args) != 1 {
			panic(errors.New("Invalid parse tree"))
		}

		kl = kl.Args[0]
	}

	if kl.T != parser.KeyList {
		panic(errors.New("Error: unexpected node type"))
	}

	ks := make([]string, 0, len(kl.Args))

	for _, k := range kl.Args {
		if k.T != parser.Key {
			panic(errors.New("Error: unexpected node type"))
		}

		validKey := false
		for _, tab := range s {
			if k.Val == "*" {
				return ks // * means all keys, represented by nil
			}
			if _, ok := tab[k.Val]; ok {
				ks = append(ks, k.Val)
				validKey = true
				break
			}
		}
		if !validKey {
			panic(errors.New("Invalid key"))
		}
	}

	return ks
}

func checkTable(s map[string]map[string]schema.Datatype, t *parser.Node) string {
	if t.T != parser.Table {
		panic(errors.New("Error: unexpected node type"))
	}

	if _, ok := s[t.Val]; ok {
		return t.Val
	} else {
		// TODO: invalid query should not result in panic
		panic(errors.New("Invalid table name"))
	}
}

func checkFilters(s map[string]map[string]schema.Datatype, filters *parser.Node) []filter {
	// TODO
	return nil
}

func checkValuesList(s map[string]map[string]schema.Datatype, valuesList *parser.Node) [][]val {
	valsList := make([][]val, len(valuesList.Args))

	for i, vs := range valuesList.Args {
		valsList[i] = make([]val, len(vs.Args))
		for j, v := range vs.Args {
			valsList[i][j] = checkValue(v.Args[0])
		}
	}

	return valsList
}

func checkAssignments(assignments *parser.Node) map[string]val {
	result := make(map[string]val)

	for _, assignment := range assignments.Args {
		key := assignment.Args[0].Val
		value := checkValue(assignment.Args[1])
		result[key] = value
	}

	return result
}

func checkWhereClause(s map[string]map[string]schema.Datatype, where *parser.Node) whereClause {
	// TODO
	return whereClause{}
}

func checkValue(v *parser.Node) val {
	v = v.Args[0]
	if v.T == parser.StrVal {
		return val{false, 0, v.Val}
	} else {
		i, err := strconv.Atoi(v.Val)
		if err != nil {
			panic(errors.New("Expected integer literal"))
		}

		return val{true, i, ""}
	}
}

func checkSelectTypes(s map[string]map[string]schema.Datatype, sq *selectQuery) error {
	// TODO: better error messages

	// TODO: add support for joins

	for _, key := range sq.keys {
		if _, ok := s[sq.table][key]; !ok {
			return errors.New("Invalid key")
		}
	}

	// TODO: type-check filters

	return nil
}

func checkInsertTypes(s map[string]map[string]schema.Datatype, iq *insertQuery) error {
	if iq.keys == nil {
		return errors.New("Insert query must not have no keys or *")
	}
	for _, key := range iq.keys {
		t, ok := s[iq.table][key]
		if !ok {
			return errors.New("Invalid key")
		}

		for i, valList := range iq.values {
			switch t {
			case schema.Int:
				if !valList[i].isNum {
					return errors.New("Invalid type: expected int")
				}
			case schema.String:
				if valList[i].isNum {
					return errors.New("Invalid type: expected string")
				}
			case schema.PrimaryKey:
				return errors.New("Invalid type: primary key values cannot be inserted directly")
			default:
				return errors.New("Unexpected field type, insert failed")
			}
		}
	}

	return nil
}

func checkUpdateTypes(s map[string]map[string]schema.Datatype, uq *updateQuery) error {
	for key, val := range uq.values {
		if s[uq.table][key] == schema.Int && !val.isNum {
			return errors.New("Invalid type: expected int")
		} else if s[uq.table][key] == schema.String && val.isNum {
			return errors.New("Invalid type: expected string")
		} else if s[uq.table][key] == schema.PrimaryKey {
			return errors.New("Primary key cannot be updated")
		} else {
			return errors.New("Unexpected field type, update failed")
		}
	}

	// TODO: type-check where clause

	return nil
}

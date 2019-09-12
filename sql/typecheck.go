package sql

import (
	"errors"
	"strconv"

	"github.com/alexbostock/alder/schema"
	"github.com/alexbostock/alder/sql/parser"
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
	return q
}

func check(s map[string]map[string]schema.Datatype, query *parser.Node) Query {
	switch query.T {
	case parser.SelectFrom:
		sq := &SelectQuery{
			Keys:    checkKeyList(s, query.Args[0]),
			Table:   checkTable(s, query.Args[1]),
			Filters: checkFilters(s, query.Args[2]),
		}

		err := checkSelectTypes(s, sq)
		if err != nil {
			panic(err)
		}

		return sq
	case parser.InsertInto:
		is := &InsertQuery{
			Keys:   checkKeyList(s, query.Args[0]),
			Values: checkValuesList(s, query.Args[2]),
			Table:  checkTable(s, query.Args[1]),
		}

		err := checkInsertTypes(s, is)
		if err != nil {
			panic(err)
		}

		return is
	case parser.UpdateSet:
		us := &UpdateQuery{
			Values: checkAssignments(query.Args[1]),
			Table:  checkTable(s, query.Args[0]),
			Where:  checkWhereClause(s, query.Args[2]),
		}

		err := checkUpdateTypes(s, us)
		if err != nil {
			panic(err)
		}

		return us
	case parser.DeleteFrom:
		ds := &DeleteQuery{
			Table: checkTable(s, query.Args[0]),
			Where: checkWhereClause(s, query.Args[1].Args[0]),
		}

		// TODO: type-check where clause of ds

		return ds
	case parser.UnionOf:
		fallthrough
	case parser.IntersectionOf:
		fallthrough
	case parser.DifferenceOf:
		return &CompoundQuery{
			query1:    check(s, query.Args[0]).(SelectQuery),
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

	return SelectQuery{}
}

func checkKeyList(s map[string]map[string]schema.Datatype, kl *parser.Node) []string {
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
			panic(errors.New("Invalid key " + k.Val))
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

func checkFilters(s map[string]map[string]schema.Datatype, filters *parser.Node) []Filter {
	// TODO
	return nil
}

func checkValuesList(s map[string]map[string]schema.Datatype, valuesList *parser.Node) [][]Val {
	valsList := make([][]Val, len(valuesList.Args))

	for i, vs := range valuesList.Args[0].Args {
		valsList[i] = make([]Val, len(vs.Args))
		for j, v := range vs.Args {
			valsList[i][j] = checkValue(v)
		}
	}

	return valsList
}

func checkAssignments(assignments *parser.Node) map[string]Val {
	result := make(map[string]Val)

	for _, assignment := range assignments.Args {
		key := assignment.Args[0].Val
		value := checkValue(assignment.Args[1])
		result[key] = value
	}

	return result
}

func checkWhereClause(s map[string]map[string]schema.Datatype, where *parser.Node) WhereClause {
	// TODO
	return WhereClause{}
}

func checkValue(v *parser.Node) Val {
	v = v.Args[0]
	if v.T == parser.StrVal {
		return Val{false, 0, v.Val}
	} else {
		i, err := strconv.Atoi(v.Val)
		if err != nil {
			panic(errors.New("Expected integer literal"))
		}

		return Val{true, i, ""}
	}
}

func checkSelectTypes(s map[string]map[string]schema.Datatype, sq *SelectQuery) error {
	// TODO: better error messages

	// TODO: add support for joins

	for _, key := range sq.Keys {
		if _, ok := s[sq.Table][key]; !ok {
			return errors.New("Invalid key")
		}
	}

	// TODO: type-check filters

	return nil
}

func checkInsertTypes(s map[string]map[string]schema.Datatype, iq *InsertQuery) error {
	if iq.Keys == nil {
		return errors.New("Insert query must not have no keys or *")
	}
	for _, key := range iq.Keys {
		t, ok := s[iq.Table][key]
		if !ok {
			return errors.New("Invalid key")
		}

		for i, valList := range iq.Values {
			switch t {
			case schema.Int:
				if !valList[i].IsNum {
					return errors.New("Invalid type: expected int")
				}
			case schema.String:
				if valList[i].IsNum {
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

func checkUpdateTypes(s map[string]map[string]schema.Datatype, uq *UpdateQuery) error {
	for key, val := range uq.Values {
		if s[uq.Table][key] == schema.Int && !val.IsNum {
			return errors.New("Invalid type: expected int")
		} else if s[uq.Table][key] == schema.String && val.IsNum {
			return errors.New("Invalid type: expected string")
		} else if s[uq.Table][key] == schema.PrimaryKey {
			return errors.New("Primary key cannot be updated")
		}
	}

	// TODO: type-check where clause

	return nil
}

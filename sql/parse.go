package sql

import (
	"errors"

	"github.com/davecgh/go-spew/spew"
)

type nonterminal int

const (
	statement nonterminal = iota
	selectFrom
	insertInto
	updateSet
	deleteFrom
	unionOf
	intersectionOf
	differenceOf
	keyList
	literalList
	key
	keys
	literals
	valueList
	assignment
	assignmentList
	filters
	whereExpr
	smaller
	larger
	equals
	innerJoin
	outerJoin
	leftJoin
	rightJoin
	orderBy
	literal
	table
	integer
	strVal
)

type node struct {
	t    nonterminal
	args []*node
	val  string
}

// A parser parses an SQL query, assumes the input does not end in a semicolon.
// It should be instantiated using newParser.
type parser struct {
	lookahead token
	l         *lexer
}

func newParser(input string) *parser {
	p := &parser{
		l: newLexer(input),
	}

	p.lookahead = p.l.lex()

	return p
}

func (p *parser) consume(t tokentype) string {
	if p.lookahead.kind == t {
		s := p.lookahead.str
		p.lookahead = p.l.lex()

		return s
	} else {
		spew.Dump(t)
		spew.Dump(p.lookahead)
		panic(errors.New("Parse error: unexpected token"))
	}
}

func (p *parser) parse() *node {
	n := &node{statement, nil, ""}

	n.args = []*node{p.query()}

	for p.lookahead.kind != eof {
		switch p.lookahead.kind {
		case union:
			p.consume(union)
			n = &node{unionOf, []*node{n, p.query()}, ""}
		case intersect:
			p.consume(intersect)
			n = &node{intersectionOf, []*node{n, p.query()}, ""}
		case minus:
			p.consume(minus)
			n = &node{differenceOf, []*node{n, p.query()}, ""}
		default:
			panic(errors.New("Parse error: expected UNION, INTERSECT or MINUS"))
		}
	}

	return n
}

func (p *parser) query() *node {
	switch p.lookahead.kind {
	case slct:
		return p.selectFrom()
	case insert:
		return p.insertInto()
	case update:
		return p.updateSet()
	case del:
		return p.del()
	default:
		panic(errors.New("Parse error: expected SELECT, INSERT, UPDATE or DELETE"))
	}
}

func (p *parser) selectFrom() *node {
	p.consume(slct)
	keys := p.keyList()
	p.consume(from)
	table := p.table()
	filters := p.filters()

	return &node{selectFrom, []*node{keys, table, filters}, ""}
}

func (p *parser) insertInto() *node {
	p.consume(insert)
	keys := p.keys()
	p.consume(into)
	table := p.table()
	p.consume(values)
	valuesList := p.valuesList()
	filters := p.filters()

	return &node{insertInto, []*node{keys, table, valuesList, filters}, ""}
}

func (p *parser) updateSet() *node {
	p.consume(update)
	table := p.table()
	p.consume(set)
	assignments := p.assignmentList()
	filters := p.filters()

	return &node{updateSet, []*node{table, assignments, filters}, ""}
}

func (p *parser) del() *node {
	p.consume(del)
	p.consume(from)
	table := p.table()
	filters := p.filters()

	return &node{deleteFrom, []*node{table, filters}, ""}
}

func (p *parser) keyList() *node {
	n := &node{keyList, make([]*node, 0, 1), ""}
	n.args = append(n.args, p.key())

	for p.lookahead.kind == comma {
		p.consume(comma)
		n.args = append(n.args, p.key())
	}

	return n
}

func (p *parser) literalList() *node {
	n := &node{literalList, make([]*node, 0, 1), ""}
	n.args = append(n.args, p.value())

	for p.lookahead.kind == comma {
		p.consume(comma)
		n.args = append(n.args, p.value())
	}

	return n
}

func (p *parser) keys() *node {
	p.consume(lparen)
	kl := p.keyList()
	p.consume(rparen)

	return &node{keys, []*node{kl}, ""}
}

func (p *parser) values() *node {
	p.consume(lparen)
	ll := p.literalList()
	p.consume(rparen)

	return &node{literals, []*node{ll}, ""}
}

func (p *parser) valuesList() *node {
	n := &node{valueList, make([]*node, 0, 1), ""}
	n.args = append(n.args, p.values())

	for p.lookahead.kind == comma {
		p.consume(comma)
		n.args = append(n.args, p.values())
	}

	return n
}

func (p *parser) assignment() *node {
	key := p.key()
	p.consume(equal)
	val := p.value()

	return &node{assignment, []*node{key, val}, ""}
}

func (p *parser) assignmentList() *node {
	n := &node{assignmentList, make([]*node, 0, 1), ""}
	n.args = append(n.args, p.assignment())

	for p.lookahead.kind == comma {
		p.consume(comma)
		n.args = append(n.args, p.assignment())
	}

	return n
}

func (p *parser) filters() *node {
	n := &node{filters, make([]*node, 0), ""}
	for {
		switch p.lookahead.kind {
		case where:
			n.args = append(n.args, p.where())
		case orderby:
			n.args = append(n.args, p.order())
		case inner, outer, left, right, join:
			n.args = append(n.args, p.join())
		case eof, union, intersect, minus:
			return n
		default:
			panic(errors.New("Parse error: expected WHERE, ORDER or [INNER|OUTER|LEFT|RIGHT] JOIN"))
		}
	}
}

func (p *parser) where() *node {
	p.consume(where)
	a := p.value()
	comp := p.comparator()
	b := p.value()

	if a.t == key {
		if b.t != literal {
			panic(errors.New("Parse error: expected key = literal"))
		}
	} else {
		if b.t != key {
			panic(errors.New("Parse error: expected literal = key"))
		}
		a, b = b, a
	}

	return &node{whereExpr, []*node{a, comp, b}, ""}
}

func (p *parser) comparator() *node {
	switch p.lookahead.kind {
	case less:
		p.consume(less)
		return &node{smaller, nil, ""}
	case greater:
		p.consume(greater)
		return &node{larger, nil, ""}
	case equal:
		p.consume(equal)
		return &node{equals, nil, ""}
	default:
		panic(errors.New("Parse error: expected <, > or ="))
	}
}
func (p *parser) order() *node {
	p.consume(orderby)
	k := p.key()

	return &node{orderBy, []*node{k}, ""}
}

func (p *parser) join() *node {
	n := &node{}

	switch p.lookahead.kind {
	case inner:
		n.t = innerJoin
		p.consume(inner)
	case outer:
		n.t = outerJoin
		p.consume(outer)
	case left:
		n.t = leftJoin
		p.consume(left)
	case right:
		n.t = rightJoin
		p.consume(right)
	case join:
		n.t = innerJoin
	default:
		panic(errors.New("Parse error: expected a JOIN statement"))
	}

	p.consume(join)

	t := p.table()
	p.consume(on)
	a := p.key()
	c := p.comparator()
	b := p.key()

	n.args = []*node{t, a, c, b}

	return n
}

func (p *parser) key() *node {
	var s string
	if p.lookahead.kind == str {
		s = p.consume(str)
	} else {
		s = p.lookahead.str
		p.consume(star)
	}

	// TODO: check s against schema

	return &node{key, nil, s}
}

func (p *parser) table() *node {
	s := p.consume(str)

	// TODO: check s against schema

	return &node{table, nil, s}
}

func (p *parser) value() *node {
	switch p.lookahead.kind {
	case str:
		return p.key()
	case num:
		return &node{literal, []*node{
			&node{integer, nil, p.consume(num)},
		}, ""}
	case quote:
		p.consume(quote)
		s := p.consume(str)
		p.consume(quote)
		return &node{literal, []*node{
			&node{strVal, nil, s},
		}, ""}
	default:
		panic(errors.New("Parse error: expected a key, \"string\" or number"))
	}
}

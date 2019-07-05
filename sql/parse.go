package sql

import "errors"

type nonterminal int

const (
	statement nonterminal = iota
	query
	selectFrom
	insertInto
	updateSet
	deleteFrom
	unionOf
	intersectionOf
	differenceOf
)

type node struct {
	t    nonterminal
	args []*node
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
		panic(errors.New("Parse error"))
	}
}

func (p *parser) parse() *node {
	n := &node{statement, nil}

	n.args = []*node{p.query()}

	for p.lookahead.kind != eof {
		switch p.lookahead.kind {
		case union:
			p.consume(union)
			n = &node{unionOf, []*node{n, p.query()}}
		case intersect:
			p.consume(intersect)
			n = &node{intersectionOf, []*node{n, p.query()}}
		case minus:
			p.consume(minus)
			n = &node{differenceOf, []*node{n, p.query()}}
		default:
			panic(errors.New("Parse error"))
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
	}
}

func (p *parser) selectFrom() *node {
	p.consume(slct)
	keys := p.keyList()
	p.consume(from)
	table := p.table()
	filters := p.filters()

	return &node{selectFrom, []*node{keys, table, filters}}
}

func (p *parser) insertInto() *node {
	p.consume(insert)
	keys := p.keys()
	p.consume(into)
	table := p.table()
	p.consume(values)
	valuesList := p.valuesList()
	filters := p.filters()

	return &node{insertInto, []*node{keys, table, valuesList, filters}}
}

func (p *parser) updateSet() *node {
	p.consume(update)
	table := p.table()
	p.consume(set)
	assignments := p.assignmentList()
	filters := p.filters()

	return &node{updateSet, table, assignments, filters}
}

func (p *parser) del() *node {
	p.consume(del)
	p.consume(from)
	table := p.table()
	filters := p.filters()

	return &node{deleteFrom, table, filters}
}

func (p *parser) keyList() *node {
	n := &node{keyList, make([]*node)}
	n.args = append(n.args, p.key())

	for p.lookahead.kind == comma {
		p.consume(comma)
		n.args = append(n.args, p.key())
	}

	return n
}

func (p *parser) literalList() *node {
	n := &node{literalList, make([]*node)}
	n.args = append(n.args, p.literal())

	for p.lookahead.kind == comma {
		p.consume(comma)
		n.args = append(n.args, p.literal())
	}
}

func (p *parser) keys() *node {
	p.consume(lparen)
	kl := p.keyList()
	p.consume(rparen)

	return &node{keys, []*node{kl}}
}

func (p *parser) values() *node {
	p.consume(lparen)
	ll := p.literalList()
	p.consume(rparen)

	return &node{literals, []*node{ll}}
}

func (p *parser) valuesList() *node {
	n := &node{valuesList, make([]*node)}
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
	val := p.literal()

	return &node{assignment, []*node{key, val}}
}

func (p *parser) assignmentList() *node {
	n := &node{assignmentList, make([]*node)}
	n.args = append(n.args, p.assignment())

	for p.lookahead.kind == comma {
		p.consume(comma)
		n.args = append(n.args, p.assignment())
	}
}

func (p *parser) filters() *node {
	n := &node{filters, make([]*node)}
	for {
		switch p.lookahead.kind {
		case where:
			n.arsg = append(n.args, p.where())
		case order:
			n.arsg = append(n.args, p.order())
		case inner:
		case outer:
		case left:
		case right:
		case join:
			n.arsg = append(n.args, p.join())
		case eof:
		case union:
		case intersect:
		case minus:
			break
		default:
			panic(errors.New("Parse error"))
		}
	}

	return n
}

func (p *parser) where() *node {
	p.consume(where)
	a := p.consume(str) // key or literal
	comp := p.comparator()
	b := p.consume(str) // key or literal

	if a.nonterminal == key {
		if b.nonterminal != literal {
			panic(errors.New("Parse error"))
		}
	} else {
		if b.nonterminal != key {
			panic(errors.New("Parse error"))
		}
		a, b = b, a
	}

	return &node{where, []*node{a, comp, b}}
}

func (p *parser) comparator() *node {
	switch p.lookahead.kind {
	case less:
		p.consume(less)
		return &node{smaller, nil}
	case greater:
		p.consume(greater)
		return &node{larger, nil}
	case equal:
		p.consume(equal)
		return &node{equals, nil}
	default:
		panic(errors.New("Parse error"))
	}
}
func (p *parser) order() *node {
	p.consume(order)
	p.consume(by)
	k := p.key()

	return &node{orderBy, k}
}

func (p *parser) join() {
	n := &node{}

	switch p.lookahead.kind {
	case inner:
		n.kind = innerJoin
		p.consume(inner)
	case outer:
		n.kind = outerJoin
		p.consume(outer)
	case left:
		n.kind = leftJoin
		p.consume(left)
	case right:
		n.kind = rightJoin
		p.consume(right)
	case join:
		n.kind = innerJoin
	default:
		panic(errors.New("Parse error"))
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
	if p.lookahead.kind == str {
		s := p.lookahead.str
		p.consume(str)
	} else {
		p.consume(star)
	}

	// TODO: check s against schema
}

func (p *parser) table() {
	s := p.consume(str)
	_ = s

	// TODO: check s against schema
}

func (p *parser) literal() {
	// All values stored are integers
	n := p.consume(num)
	_ = n
}

func (p *parser) value() {
	switch p.lookahead.kind {
	case str:
		p.key()
	case num:
		p.literal()
	default:
		panic(errors.New("Parse error"))
	}
}

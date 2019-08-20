package parser

import (
	"errors"

	"github.com/alexbostock/alder/sql/lexer"
	"github.com/davecgh/go-spew/spew"
)

type Nonterminal int

const (
	_ Nonterminal = iota
	SelectFrom
	InsertInto
	UpdateSet
	DeleteFrom
	UnionOf
	IntersectionOf
	DifferenceOf
	KeyList
	LiteralList
	Key
	Keys
	Literals
	ValueList
	Assignment
	AssignmentList
	Filters
	WhereExpr
	Smaller
	Larger
	Equals
	InnerJoin
	OuterJoin
	LeftJoin
	RightJoin
	OrderBy
	Literal
	Table
	Integer
	StrVal
)

type Node struct {
	T    Nonterminal
	Args []*Node
	Val  string
}

// A parser parses an SQL query, assumes the input does not end in a semicolon.
// It should be instantiated using newParser.
type Parser struct {
	lookahead lexer.Token
	l         *lexer.Lexer
}

func New(input string) *Parser {
	p := &Parser{
		l: lexer.New(input),
	}

	p.lookahead = p.l.Lex()

	return p
}

func (p *Parser) consume(t lexer.TokenType) string {
	if p.lookahead.Kind == t {
		s := p.lookahead.Str
		p.lookahead = p.l.Lex()

		return s
	} else {
		spew.Dump(t)
		spew.Dump(p.lookahead)
		panic(errors.New("Parse error: unexpected token"))
	}
}

func (p *Parser) Parse() *Node {
	n := p.query()

	for p.lookahead.Kind != lexer.Eof {
		switch p.lookahead.Kind {
		case lexer.Union:
			p.consume(lexer.Union)
			n = &Node{UnionOf, []*Node{n, p.query()}, ""}
		case lexer.Intersect:
			p.consume(lexer.Intersect)
			n = &Node{IntersectionOf, []*Node{n, p.query()}, ""}
		case lexer.Minus:
			p.consume(lexer.Minus)
			n = &Node{DifferenceOf, []*Node{n, p.query()}, ""}
		default:
			panic(errors.New("Parse error: expected UNION, INTERSECT or MINUS"))
		}
	}

	return n
}

func (p *Parser) query() *Node {
	switch p.lookahead.Kind {
	case lexer.Slct:
		return p.selectFrom()
	case lexer.Insert:
		return p.insertInto()
	case lexer.Update:
		return p.updateSet()
	case lexer.Del:
		return p.del()
	default:
		panic(errors.New("Parse error: expected SELECT, INSERT, UPDATE or DELETE"))
	}
}

func (p *Parser) selectFrom() *Node {
	p.consume(lexer.Slct)
	keys := p.keyList()
	p.consume(lexer.From)
	table := p.table()
	filters := p.filters()

	return &Node{SelectFrom, []*Node{keys, table, filters}, ""}
}

func (p *Parser) insertInto() *Node {
	p.consume(lexer.Insert)
	keys := p.keys()
	p.consume(lexer.Into)
	table := p.table()
	p.consume(lexer.Values)
	valuesList := p.valuesList()

	return &Node{InsertInto, []*Node{keys, table, valuesList}, ""}
}

func (p *Parser) updateSet() *Node {
	p.consume(lexer.Update)
	table := p.table()
	p.consume(lexer.Set)
	assignments := p.assignmentList()
	filters := p.filters()

	return &Node{UpdateSet, []*Node{table, assignments, filters}, ""}
}

func (p *Parser) del() *Node {
	p.consume(lexer.Del)
	p.consume(lexer.From)
	table := p.table()
	filters := p.filters()

	return &Node{DeleteFrom, []*Node{table, filters}, ""}
}

func (p *Parser) keyList() *Node {
	n := &Node{KeyList, make([]*Node, 0, 1), ""}
	n.Args = append(n.Args, p.key())

	for p.lookahead.Kind == lexer.Comma {
		p.consume(lexer.Comma)
		n.Args = append(n.Args, p.key())
	}

	return n
}

func (p *Parser) literalList() *Node {
	n := &Node{LiteralList, make([]*Node, 0, 1), ""}
	n.Args = append(n.Args, p.value())

	for p.lookahead.Kind == lexer.Comma {
		p.consume(lexer.Comma)
		n.Args = append(n.Args, p.value())
	}

	return n
}

func (p *Parser) keys() *Node {
	p.consume(lexer.Lparen)
	kl := p.keyList()
	p.consume(lexer.Rparen)

	return &Node{Keys, []*Node{kl}, ""}
}

func (p *Parser) values() *Node {
	p.consume(lexer.Lparen)
	ll := p.literalList()
	p.consume(lexer.Rparen)

	return &Node{Literals, []*Node{ll}, ""}
}

func (p *Parser) valuesList() *Node {
	n := &Node{ValueList, make([]*Node, 0, 1), ""}
	n.Args = append(n.Args, p.values())

	for p.lookahead.Kind == lexer.Comma {
		p.consume(lexer.Comma)
		n.Args = append(n.Args, p.values())
	}

	return n
}

func (p *Parser) assignment() *Node {
	key := p.key()
	p.consume(lexer.Equal)
	val := p.value()

	return &Node{Assignment, []*Node{key, val}, ""}
}

func (p *Parser) assignmentList() *Node {
	n := &Node{AssignmentList, make([]*Node, 0, 1), ""}
	n.Args = append(n.Args, p.assignment())

	for p.lookahead.Kind == lexer.Comma {
		p.consume(lexer.Comma)
		n.Args = append(n.Args, p.assignment())
	}

	return n
}

func (p *Parser) filters() *Node {
	n := &Node{Filters, make([]*Node, 0), ""}
	for {
		switch p.lookahead.Kind {
		case lexer.Where:
			n.Args = append(n.Args, p.where())
			for p.lookahead.Kind == lexer.And {
				n.Args = append(n.Args, p.and())
			}
		case lexer.Orderby:
			n.Args = append(n.Args, p.order())
		case lexer.Inner, lexer.Outer, lexer.Left, lexer.Right, lexer.Join:
			n.Args = append(n.Args, p.join())
		case lexer.Eof, lexer.Union, lexer.Intersect, lexer.Minus:
			return n
		default:
			panic(errors.New("Parse error: expected WHERE, ORDER or [INNER|OUTER|LEFT|RIGHT] JOIN"))
		}
	}
}

func (p *Parser) where() *Node {
	p.consume(lexer.Where)
	a := p.value()
	comp := p.comparator()
	b := p.value()

	return &Node{WhereExpr, []*Node{a, comp, b}, ""}
}

func (p *Parser) and() *Node {
	p.consume(lexer.And)
	a := p.value()
	comp := p.comparator()
	b := p.value()

	return &Node{WhereExpr, []*Node{a, comp, b}, ""}
}

func (p *Parser) comparator() *Node {
	switch p.lookahead.Kind {
	case lexer.Less:
		p.consume(lexer.Less)
		return &Node{Smaller, nil, ""}
	case lexer.Greater:
		p.consume(lexer.Greater)
		return &Node{Larger, nil, ""}
	case lexer.Equal:
		p.consume(lexer.Equal)
		return &Node{Equals, nil, ""}
	default:
		panic(errors.New("Parse error: expected <, > or ="))
	}
}
func (p *Parser) order() *Node {
	p.consume(lexer.Orderby)
	k := p.key()

	return &Node{OrderBy, []*Node{k}, ""}
}

func (p *Parser) join() *Node {
	n := &Node{}

	switch p.lookahead.Kind {
	case lexer.Inner:
		n.T = InnerJoin
		p.consume(lexer.Inner)
	case lexer.Outer:
		n.T = OuterJoin
		p.consume(lexer.Outer)
	case lexer.Left:
		n.T = LeftJoin
		p.consume(lexer.Left)
	case lexer.Right:
		n.T = RightJoin
		p.consume(lexer.Right)
	case lexer.Join:
		n.T = InnerJoin
	default:
		panic(errors.New("Parse error: expected a JOIN statement"))
	}

	p.consume(lexer.Join)

	t := p.table()
	p.consume(lexer.On)
	a := p.key()
	c := p.comparator()
	b := p.key()

	n.Args = []*Node{t, a, c, b}

	return n
}

func (p *Parser) key() *Node {
	var s string
	if p.lookahead.Kind == lexer.Str {
		s = p.consume(lexer.Str)
	} else {
		s = "*"
		p.consume(lexer.Star)
	}

	return &Node{Key, nil, s}
}

func (p *Parser) table() *Node {
	s := p.consume(lexer.Str)

	return &Node{Table, nil, s}
}

func (p *Parser) value() *Node {
	switch p.lookahead.Kind {
	case lexer.Str:
		return p.key()
	case lexer.Num:
		return &Node{Literal, []*Node{
			&Node{Integer, nil, p.consume(lexer.Num)},
		}, ""}
	case lexer.StringLit:
		s := p.consume(lexer.StringLit)
		return &Node{Literal, []*Node{
			&Node{StrVal, nil, s},
		}, ""}
	default:
		panic(errors.New("Parse error: expected a key, \"string\" or number"))
	}
}

package sql

import (
	"strings"
	"unicode"
)

type tokentype int

const (
	_ tokentype = iota
	slct
	from
	where
	order
	by
	inner
	outer
	left
	right
	join
	union
	intersect
	minus
	insert
	into
	values
	update
	set
	del
	comma
	lparen
	rparen
	equal
	less
	greater
	star
	str
	num
)

type token struct {
	kind tokentype
	str  string
}

type lexer struct {
	str        string
	dictionary map[string]tokentype
}

func newLexer(input string) *lexer {
	return &lexer{str: strings.ToLower(input),
		dictionary: map[string]tokentype{
			"select":    slct,
			"from":      from,
			"where":     where,
			"order":     order,
			"by":        by,
			"inner":     inner,
			"outer":     outer,
			"left":      left,
			"right":     right,
			"join":      join,
			"union":     union,
			"intersect": intersect,
			"minus":     minus,
			"insert":    insert,
			"into":      into,
			"values":    values,
			"update":    update,
			"set":       set,
			"delete":    del},
	}
}

func (l *lexer) lex() token {
startloop:
	for {
		if len(l.str) == 0 {
			break startloop
		}

		switch rune(l.str[0]) {
		case ' ':
			l.str = l.str[1:]
		case '\t':
			l.str = l.str[1:]
		case '\n':
			l.str = l.str[1:]
		case '(':
			l.str = l.str[1:]
			return token{lparen, ""}
		case ')':
			l.str = l.str[1:]
			return token{rparen, ""}
		case ',':
			l.str = l.str[1:]
			return token{comma, ""}
		case '=':
			l.str = l.str[1:]
			return token{equal, ""}
		case '>':
			l.str = l.str[1:]
			return token{greater, ""}
		case '<':
			l.str = l.str[1:]
			return token{less, ""}
		case '*':
			l.str = l.str[1:]
			return token{star, ""}
		default:
			break startloop
		}
	}

	for str, tok := range l.dictionary {
		if strings.HasPrefix(l.str, str) {
			l.str = l.str[len(str):]
			return token{tok, ""}
		}
	}

	var s strings.Builder
	isNumber := true

builderloop:
	for {
		if len(l.str) == 0 {
			break builderloop
		}

		switch l.str[0] {
		case ' ':
			break builderloop
		case '\t':
			break builderloop
		case '\n':
			break builderloop
		case '(':
			break builderloop
		case ')':
			break builderloop
		case ',':
			break builderloop
		case '=':
			break builderloop
		case '>':
			break builderloop
		case '<':
			break builderloop
		case '*':
			break builderloop
		default:
			if !unicode.IsDigit(rune(l.str[0])) {
				isNumber = false
			}

			s.Write([]byte{l.str[0]})
			l.str = l.str[1:]
		}
	}

	if isNumber {
		return token{num, s.String()}
	} else {
		return token{str, s.String()}
	}
}

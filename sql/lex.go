package sql

import (
	"regexp"
	"strings"
	"unicode"
)

type tokentype int

const (
	eof tokentype = iota
	slct
	from
	where
	orderby
	inner
	outer
	left
	right
	join
	on
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
	stringLit
)

type token struct {
	kind tokentype
	str  string
}

type lexer struct {
	str           string
	dictionary    map[string]tokentype
	strPattern    *regexp.Regexp
	strLitPattern *regexp.Regexp
	numPattern    *regexp.Regexp
}

func newLexer(input string) *lexer {
	return &lexer{str: strings.ToLower(input),
		dictionary: map[string]tokentype{
			"select":    slct,
			"from":      from,
			"where":     where,
			"order by":  orderby,
			"inner":     inner,
			"outer":     outer,
			"left":      left,
			"right":     right,
			"join":      join,
			"on":        on,
			"union":     union,
			"intersect": intersect,
			"minus":     minus,
			"insert":    insert,
			"into":      into,
			"values":    values,
			"update":    update,
			"set":       set,
			"delete":    del,
			"(":         lparen,
			")":         rparen,
			",":         comma,
			"=":         equal,
			">":         greater,
			"<":         less,
			"*":         star,
		},
		strPattern:    regexp.MustCompile("[a-zA-Z0-9_\\.]+"),
		strLitPattern: regexp.MustCompile("'.*'"),
		numPattern:    regexp.MustCompile("[0-9]+"),
	}
}

func (l *lexer) lex() token {
	for {
		if len(l.str) == 0 {
			return token{eof, ""}
		}

		if unicode.IsSpace(rune(l.str[0])) {
			l.str = l.str[1:]
		}

		for str, tok := range l.dictionary {
			if strings.HasPrefix(l.str, str) {
				l.str = l.str[len(str):]
				return token{tok, ""}
			}
		}

		if match := l.strPattern.FindString(l.str); match != "" && strings.HasPrefix(l.str, match) {
			l.str = l.str[len(match):]
			return token{str, strings.TrimSpace(match)}
		}

		if match := l.strLitPattern.FindString(l.str); match != "" && strings.HasPrefix(l.str, match) {
			l.str = l.str[len(match):]
			return token{stringLit, match[1 : len(match)-1]}
		}

		if match := l.numPattern.FindString(l.str); match != "" && strings.HasPrefix(l.str, match) {
			l.str = l.str[len(match):]
			return token{num, match}
		}
	}
}

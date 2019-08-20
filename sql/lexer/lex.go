package lexer

import (
	"regexp"
	"strings"
	"unicode"
)

type TokenType int

const (
	Eof TokenType = iota
	Slct
	From
	Where
	And
	Orderby
	Inner
	Outer
	Left
	Right
	Join
	On
	Union
	Intersect
	Minus
	Insert
	Into
	Values
	Update
	Set
	Del
	Comma
	Lparen
	Rparen
	Equal
	Less
	Greater
	Star
	Str
	Num
	StringLit
)

type Token struct {
	Kind TokenType
	Str  string
}

type Lexer struct {
	str           string
	dictionary    map[string]TokenType
	strPattern    *regexp.Regexp
	strLitPattern *regexp.Regexp
	numPattern    *regexp.Regexp
}

func New(input string) *Lexer {
	return &Lexer{str: strings.ToLower(input),
		dictionary: map[string]TokenType{
			"select":    Slct,
			"from":      From,
			"where":     Where,
			"and":       And,
			"order by":  Orderby,
			"inner":     Inner,
			"outer":     Outer,
			"left":      Left,
			"right":     Right,
			"join":      Join,
			"on":        On,
			"union":     Union,
			"intersect": Intersect,
			"minus":     Minus,
			"insert":    Insert,
			"into":      Into,
			"values":    Values,
			"update":    Update,
			"set":       Set,
			"delete":    Del,
			"(":         Lparen,
			")":         Rparen,
			",":         Comma,
			"=":         Equal,
			">":         Greater,
			"<":         Less,
			"*":         Star,
		},
		strPattern:    regexp.MustCompile("[a-zA-Z0-9_\\.]+"),
		strLitPattern: regexp.MustCompile("'.*'"),
		numPattern:    regexp.MustCompile("[0-9]+"),
	}
}

func (l *Lexer) Lex() Token {
	for {
		if len(l.str) == 0 {
			return Token{Eof, ""}
		}

		if unicode.IsSpace(rune(l.str[0])) {
			l.str = l.str[1:]
		}

		for str, tok := range l.dictionary {
			if strings.HasPrefix(l.str, str) {
				l.str = l.str[len(str):]
				return Token{tok, ""}
			}
		}

		if match := l.strPattern.FindString(l.str); match != "" && strings.HasPrefix(l.str, match) {
			l.str = l.str[len(match):]
			return Token{Str, strings.TrimSpace(match)}
		}

		if match := l.strLitPattern.FindString(l.str); match != "" && strings.HasPrefix(l.str, match) {
			l.str = l.str[len(match):]
			return Token{StringLit, match[1 : len(match)-1]}
		}

		if match := l.numPattern.FindString(l.str); match != "" && strings.HasPrefix(l.str, match) {
			l.str = l.str[len(match):]
			return Token{Num, match}
		}
	}
}

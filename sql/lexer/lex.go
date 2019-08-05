package lexer

import (
	"strings"
	"unicode"
)

type TokenType int

const (
	Eof TokenType = iota
	Slct
	From
	Where
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
	Quote
)

type Token struct {
	Kind TokenType
	Str  string
}

type Lexer struct {
	str        string
	dictionary map[string]TokenType
}

func New(input string) *Lexer {
	return &Lexer{str: strings.ToLower(input),
		dictionary: map[string]TokenType{
			"select":    Slct,
			"from":      From,
			"where":     Where,
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
			"delete":    Del},
	}
}

func (l *Lexer) Lex() Token {
	if len(l.str) == 0 {
		return Token{Eof, ""}
	}

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
			return Token{Lparen, ""}
		case ')':
			l.str = l.str[1:]
			return Token{Rparen, ""}
		case ',':
			l.str = l.str[1:]
			return Token{Comma, ""}
		case '=':
			l.str = l.str[1:]
			return Token{Equal, ""}
		case '>':
			l.str = l.str[1:]
			return Token{Greater, ""}
		case '<':
			l.str = l.str[1:]
			return Token{Less, ""}
		case '*':
			l.str = l.str[1:]
			return Token{Star, ""}
		case '"':
			l.str = l.str[1:]
			return Token{Quote, ""}
		default:
			break startloop
		}
	}

	for str, tok := range l.dictionary {
		if strings.HasPrefix(l.str, str) {
			l.str = l.str[len(str):]
			return Token{tok, ""}
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
		case ' ', '\t', '\n', '(', ')', ',', '=', '>', '<', '*', '"':
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
		return Token{Num, s.String()}
	} else {
		return Token{Str, s.String()}
	}
}

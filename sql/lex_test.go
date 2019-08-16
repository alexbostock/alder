package sql

import "testing"

func TestLexer(t *testing.T) {
	l := newLexer("SELECT price FROM products WHERE name = 'apples and pears'")

	tokens := []token{
		token{slct, ""},
		token{str, "price"},
		token{from, ""},
		token{str, "products"},
		token{where, ""},
		token{str, "name"},
		token{equal, ""},
		token{stringLit, "apples and pears"},
		token{eof, ""},
	}

	for _, token := range tokens {
		if lexed := l.lex(); lexed != token {
			t.Errorf("Expected %v, got %v", token, lexed)
		}
	}
}

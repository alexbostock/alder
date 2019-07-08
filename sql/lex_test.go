package sql

import "testing"

func TestLexer(t *testing.T) {
	l := newLexer("SELECT price FROM products WHERE name = \"apples\"")

	tokens := []token{
		token{slct, ""},
		token{str, "price"},
		token{from, ""},
		token{str, "products"},
		token{where, ""},
		token{str, "name"},
		token{equal, ""},
		token{quote, ""},
		token{str, "apples"},
		token{quote, ""},
		token{eof, ""},
	}

	for _, token := range tokens {
		if lexed := l.lex(); lexed != token {
			t.Errorf("Expected %v, got %v", token, lexed)
		}
	}
}

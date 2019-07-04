package sql

import "testing"

func TestLexer(t *testing.T) {
	l := newLexer("SELECT name, price FROM products WHERE id < 5")

	tokens := []token{
		token{slct, ""},
		token{str, "name"},
		token{comma, ""},
		token{str, "price"},
		token{from, ""},
		token{str, "products"},
		token{where, ""},
		token{str, "id"},
		token{less, ""},
		token{num, "5"},
	}

	for _, token := range tokens {
		if lexed := l.lex(); lexed != token {
			t.Errorf("Expected %v, got %v", token, lexed)
		}
	}
}

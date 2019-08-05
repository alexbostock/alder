package lexer

import "testing"

func TestLexer(t *testing.T) {
	l := New("SELECT price FROM products WHERE name = \"apples\"")

	tokens := []Token{
		Token{Slct, ""},
		Token{Str, "price"},
		Token{From, ""},
		Token{Str, "products"},
		Token{Where, ""},
		Token{Str, "name"},
		Token{Equal, ""},
		Token{Quote, ""},
		Token{Str, "apples"},
		Token{Quote, ""},
		Token{Eof, ""},
	}

	for _, token := range tokens {
		if lexed := l.Lex(); lexed != token {
			t.Errorf("Expected %v, got %v", token, lexed)
		}
	}
}

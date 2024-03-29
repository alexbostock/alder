package parser

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestParser(t *testing.T) {
	queries := []string{
		"SELECT price FROM order WHERE user_id = 1",
		"select surname, price from order join user on user.id = order.user_id intersect select surname from user where forename = 'Alex' and surname = 'Bostock'",
		"INSERT INTO user (forename, surname) VALUES ('Alex', 'Bostock')",
	}

	for _, q := range queries {
		l := New(q)
		spew.Dump(l.Parse())
	}
}

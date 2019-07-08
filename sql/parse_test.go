package sql

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestParser(t *testing.T) {
	queries := []string{
		"SELECT price FROM order WHERE user_id = 1",
		"select surname, price from order join user on user.id = order.user_id intersect select surname from user where forename = \"Alex\"",
	}

	for _, q := range queries {
		l := newParser(q)
		spew.Dump(l.parse())
	}
}

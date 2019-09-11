package database

import (
	"reflect"
	"testing"

	"github.com/alexbostock/alder/sql"
)

func TestSerialize(t *testing.T) {
	data := make(map[string]sql.Val)

	data["items"] = sql.Val{IsNum: false, Str: "apples"}
	data["price"] = sql.Val{IsNum: true, Num: 100}
	data["user_id"] = sql.Val{IsNum: true, Num: 5}

	if !reflect.DeepEqual(deserialise(serialise(data)), data) {
		t.Error("Serialisation/deserialisation failed")
	}
}

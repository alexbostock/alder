package store

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestBPTree(t *testing.T) {
	store := NewBPTree(3)

	store.Insert(0, []byte{1})
	store.Insert(5, []byte{2})

	r := store.Get(5)
	if !bytes.Equal(r, []byte{2}) {
		t.Error("Incorrect value read.")
	}
	r = store.Get(3)
	if r != nil {
		t.Error("Value returned where no value should exist.")
	}

	store.Insert(3, []byte{3})

	r = store.Get(5)
	if !bytes.Equal(r, []byte{2}) {
		t.Error("Incorrect value read.")
	}
	ok := store.Insert(3, []byte{4})
	if ok {
		t.Error("Duplicate key should not be successfully inserted.")
	}

	store.Insert(4, []byte{4})

	r = store.Get(5)
	if !bytes.Equal(r, []byte{2}) {
		t.Error("Incorrect value read.")
	}

	// Test many operations by applying the same operations to a map and comparing

	store = NewBPTree(4)
	reference := make(map[int][]byte)

	succ := func(x []byte) []byte {
		y := make([]byte, len(x))

		if len(y) > 0 {
			y[len(y)-1]++
		}

		return y
	}

	for i := 0; i < 1000; i++ {
		key := rand.Intn(100)

		r := rand.Float32()
		if r < 0.25 {
			get(t, store, reference, key)
		} else if r < 0.5 {
			val := make([]byte, 10)
			rand.Read(val)
			insert(t, store, reference, key, val)
		} else if r < 0.75 {
			update(t, store, reference, key, succ)
		} else {
			//del(t, store, reference, key)
		}
	}
}

func get(t *testing.T, store Store, reference map[int][]byte, key int) {
	t.Log("get", key)

	val := store.Get(key)
	if !bytes.Equal(val, reference[key]) {
		t.Errorf("Incorrect value read from store. Expected %v. Got %v.", reference[key], val)
	}
}

func insert(t *testing.T, store Store, reference map[int][]byte, key int, val []byte) {
	t.Log("insert", key)

	ok := store.Insert(key, val)
	if !ok && reference[key] == nil {
		t.Error("Insert failed.")
	}
	if ok {
		reference[key] = val

		r := store.Get(key)
		if !bytes.Equal(r, val) {
			t.Error("Inserted value could not be retrieved.")
		}
	}
}

func update(t *testing.T, store Store, reference map[int][]byte, key int, f func([]byte) []byte) {
	t.Log("update", key)

	ok := store.Update(key, f)
	if !ok && reference[key] != nil {
		t.Error("Update failed.")
	}
	if ok {
		reference[key] = f(reference[key])
	}
}

func del(t *testing.T, store Store, reference map[int][]byte, key int) {
	t.Log("del", key)

	present := reference[key] != nil
	deleted := store.Delete(key)
	if present != deleted {
		t.Error("Delete failed.")
	}
	if deleted {
		delete(reference, key)
	}
}

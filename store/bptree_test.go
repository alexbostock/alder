package store

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestBPTree(t *testing.T) {
	// Test many operations by applying the same operations to a map and comparing

	store := NewBPTree(7)
	reference := make(map[int][]byte)

	succ := func(x []byte) []byte {
		y := make([]byte, len(x))

		if len(y) > 0 {
			y[len(y)-1]++
		}

		return y
	}

	for i := 0; i < 1000; i++ {
		key := rand.Int()

		r := rand.Float32()
		if r < 0.25 {
			get(t, store, reference, key)
		} else if r < 0.5 {
			val := make([]byte, rand.Intn(100))
			rand.Read(val)
			insert(t, store, reference, key, val)
		} else if r < 0.75 {
			update(t, store, reference, key, succ)
		} else {
			del(t, store, reference, key)
		}
	}
}

func get(t *testing.T, store Store, reference map[int][]byte, key int) {
	val := store.Get(key)
	if !bytes.Equal(val, reference[key]) {
		t.Errorf("Incorrect value read from store. Expected %v. Got %v.", reference[key], val)
	}
}

func insert(t *testing.T, store Store, reference map[int][]byte, key int, val []byte) {
	ok := store.Insert(key, val)
	if !ok {
		t.Error("Insert failed.")
	}
	reference[key] = val
}

func update(t *testing.T, store Store, reference map[int][]byte, key int, f func([]byte) []byte) {
	ok := store.Update(key, f)
	if !ok {
		t.Error("Update failed.")
	}
	reference[key] = f(reference[key])
}

func del(t *testing.T, store Store, reference map[int][]byte, key int) {
	present := reference[key] != nil
	deleted := store.Delete(key)
	if present != deleted {
		t.Error("Delete failed.")
	}
	delete(reference, key)
}

package store

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"reflect"
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
		copy(y, x)

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
			del(t, store, reference, key)
		}

		errs := store.root.verifyInvariants(store.b)
		for _, err := range errs {
			t.Errorf(err.Error())
		}
	}
}

func TestGetRange(t *testing.T) {
	store := NewBPTree(4)

	for i := 0; i < 100; i++ {
		if ok := store.Insert(i, []byte{byte(i)}); !ok {
			t.Error("Insertion failed")
		}
	}

	errs := store.root.verifyInvariants(store.b)
	for _, err := range errs {
		t.Errorf(err.Error())
	}

	res := store.GetRange(5, 23)
	if len(res) != 23-5+1 {
		t.Error("Incorrect range returned")
	}
	for i := 5; i <= 23; i++ {
		if !bytes.Equal(res[i], []byte{byte(i)}) {
			t.Errorf("Value missing from range: %v", i)
		}
	}
}

func TestGetAllWhere(t *testing.T) {
	store := NewBPTree(4)

	for i := 0; i < 100; i++ {
		store.Insert(i, []byte{byte(i)})
	}

	res := store.GetAllWhere(func(key int, val []byte) bool {
		return 5 <= key && key <= 23
	})
	if !reflect.DeepEqual(res, store.GetRange(5, 23)) {
		t.Error("TestGetAllWhere output does not match equivalent GetRange")
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
		store.Update(key, f)
	}
	if ok {
		reference[key] = f(reference[key])
	}
}

func del(t *testing.T, store Store, reference map[int][]byte, key int) {
	present := reference[key] != nil
	deleted := store.Delete(key)
	if present != deleted {
		t.Error("Delete failed.")
	}
	if deleted {
		t.Log("del", key)

		delete(reference, key)
	}
}

func (n *nonleafnode) verifyInvariants(b int) []error {
	errs := make([]error, 0)

	if n.parent == nil {
		if len(n.children) < 2 {
			errs = append(errs, fmt.Errorf("Root has too few children %v", n))
		}
		if len(n.children) > b {
			errs = append(errs, fmt.Errorf("Root has too many children %v", n))
		}
	} else {
		if len(n.children) < int(math.Ceil(float64(b)/2)) {
			errs = append(errs, fmt.Errorf("Internal node has too few children %v", n))
		}
		if len(n.children) > b {
			errs = append(errs, fmt.Errorf("Internal node has too many children %v", n))
		}
	}

	if len(n.children) != len(n.keys)+1 {
		errs = append(errs, fmt.Errorf("Node does not satisfy #children == #keys + 1 %v", n))
	}

	for _, child := range n.children {
		errs = append(errs, child.verifyInvariants(b)...)

		if child.getParent() != n {
			errs = append(errs, fmt.Errorf("Incorrect parent pointer %+v %p %p", child, n, child.getParent()))
		}
	}

	for i, key := range n.keys {
		if key != n.children[i+1].firstKey() {
			errs = append(errs, fmt.Errorf("Node key does match smallest leaf key %v", n))
		}
	}

	return errs
}

func (n *leafnode) verifyInvariants(b int) []error {
	errs := make([]error, 0)

	if len(n.children) < int(math.Ceil(float64(b)/2)) && n.parent != nil {
		errs = append(errs, fmt.Errorf("Leaf node has too few values %v", n))
	}
	if len(n.children) > b {
		errs = append(errs, fmt.Errorf("Leaf node has too many values %v", n))
	}

	if len(n.children) != len(n.keys) {
		errs = append(errs, fmt.Errorf("#keys in leaf does not match #values %v", n))
	}

	return errs
}

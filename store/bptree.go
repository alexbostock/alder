package store

import (
	"errors"
	"math"
)

// A bptree is an in-memory B+ tree implementation of Store.
type bptree struct {
	b    int
	root treenode
}

// Get searches for a value in the B+ tree.
func (t *bptree) Get(key int) []byte {
	return t.root.get(key)
}

// Insert adds a new key-value pair to the tree.
func (t *bptree) Insert(key int, val []byte) bool {
	newKey, newChild, err := t.root.insert(key, val, t.b)

	if err != nil {
		return false
	}

	if newChild != nil {
		newNode := newNonLeaf(t.b)
		newNode.keys = []int{newKey}
		newNode.children = []treenode{t.root, newChild}

		t.root.setParent(newNode)
		newChild.setParent(newNode)
		t.root = newNode
	}

	return true
}

// Update applies the given function to an existing value, or returns false if
// no value is present with the given key.
func (t *bptree) Update(key int, f func([]byte) []byte) bool {
	return t.root.update(key, f)
}

// Delete deletes the record associated with a given key, if such a record exists.
// It returns true if a record was deleted.
func (t *bptree) Delete(key int) bool {
	return t.root.del(key, t.b, nil, nil)
}

// NewBPTree instantitates a B+ tree, with a given branching factor.
func NewBPTree(b int) *bptree {
	return &bptree{
		b,
		newLeaf(b),
	}
}

// A treenode can be either a leafnode or a nonleafnode.
type treenode interface {
	del(key, b int, leftSibling, rightSibling treenode) bool
	get(key int) []byte
	insert(key int, val []byte, b int) (int, treenode, error)
	update(key int, f func([]byte) []byte) bool
	getParent() *nonleafnode
	setParent(p *nonleafnode)
	firstKey() int
	concat(t treenode)
	verifyInvariants(b int) []error
}

type nonleafnode struct {
	keys     []int      // List of keys (length b-1)
	children []treenode // Pointers to children (length b)
	parent   *nonleafnode
}

func newNonLeaf(b int) *nonleafnode {
	return &nonleafnode{
		make([]int, 0, b-1),
		make([]treenode, 0, b),
		nil,
	}
}

// A leafnode is a B+ tree leaf. Instead of treenodes as children, it has record
// nodes. Record nodes are just byte slices.
type leafnode struct {
	keys     []int
	children [][]byte
	nextLeaf *leafnode
	prevLeaf *leafnode
	parent   *nonleafnode
}

func newLeaf(b int) *leafnode {
	return &leafnode{
		make([]int, 0, b),
		make([][]byte, 0, b),
		nil,
		nil,
		nil,
	}
}

func (n *nonleafnode) get(key int) []byte {
	for i, k := range n.keys {
		if k > key {
			return n.children[i].get(key)
		}
	}
	return n.children[len(n.keys)].get(key)
}

func (n *leafnode) get(key int) []byte {
	for i, k := range n.keys {
		if k == key {
			return n.children[i]
		}
		if k > key {
			return nil
		}
	}
	return nil
}

func (n *nonleafnode) insert(key int, val []byte, b int) (int, treenode, error) {
	for i, k := range n.keys {
		if k == key {
			return -1, nil, errors.New("Key already present")
		}
		if k > key {
			newKey, newChild, err := n.children[i].insert(key, val, b)
			return n.insertChild(newKey, newChild, err, b)
		}
	}
	newKey, newChild, err := n.children[len(n.keys)].insert(key, val, b)
	return n.insertChild(newKey, newChild, err, b)
}

func (n *nonleafnode) insertChild(k int, c treenode, err error, b int) (int, treenode, error) {
	if err != nil {
		return k, c, err
	}
	if c == nil {
		return -1, nil, nil
	}
	if len(n.children) < b {
		n.insChild(k, c)
		return -1, nil, nil
	} else {
		i := int(math.Ceil(float64(len(n.keys)+1)/2)) - 1
		median := n.keys[i]

		newNode := newNonLeaf(b)
		newNode.setParent(n.parent)

		newNode.keys = make([]int, b-i-2)
		newNode.children = make([]treenode, b-i-1)
		copy(newNode.keys, n.keys[i+1:])
		copy(newNode.children, n.children[i+1:])

		for _, child := range newNode.children {
			child.setParent(newNode)
		}

		n.keys = n.keys[:i]
		n.children = n.children[:i+1]

		if k >= median {
			newNode.insChild(k, c)
		} else {
			n.insChild(k, c)
		}

		return median, newNode, nil
	}
}

func (n *nonleafnode) insChild(key int, val treenode) {
	// Insert the given key-child pair into a node (assumes the node has space)

	var index int
	for i, k := range n.keys {
		index = i
		if k > key {
			break
		}
	}
	if len(n.keys) > 0 && key > n.keys[len(n.keys)-1] {
		index++
	}

	keys := append(n.keys, 0)
	copy(keys[index+1:], n.keys[index:])
	keys[index] = key
	n.keys = keys

	index++
	vals := append(n.children, nil)
	copy(vals[index+1:], n.children[index:])
	vals[index] = val
	n.children = vals

	val.setParent(n)
}

func (n *leafnode) insert(key int, val []byte, b int) (int, treenode, error) {
	for _, k := range n.keys {
		if k == key {
			return -1, nil, errors.New("Key already present")
		}
	}

	if len(n.keys) < b {
		n.insRecord(key, val)

		return -1, nil, nil
	} else {
		// Split
		i := int(math.Ceil(float64(len(n.keys)+1)/2)) - 1
		median := n.keys[i]

		newNode := newLeaf(b)

		newNode.keys = make([]int, b-i)
		newNode.children = make([][]byte, b-i)
		copy(newNode.keys, n.keys[i:])
		copy(newNode.children, n.children[i:])

		newNode.nextLeaf = n.nextLeaf
		newNode.prevLeaf = n
		if newNode.nextLeaf != nil {
			newNode.nextLeaf.prevLeaf = newNode
		}
		newNode.setParent(n.parent)

		n.keys = n.keys[:i]
		n.children = n.children[:i]
		n.nextLeaf = newNode

		if key >= median {
			newNode.insRecord(key, val)
		} else {
			n.insRecord(key, val)
		}

		return median, newNode, nil
	}
}

func (n *leafnode) insRecord(key int, val []byte) {
	// Insert the given key-value pair into a node (assumes the node has space)

	var index int
	for i, k := range n.keys {
		index = i
		if k > key {
			break
		}
	}
	if len(n.keys) > 0 && key > n.keys[len(n.keys)-1] {
		index++
	}

	keys := append(n.keys, 0)
	copy(keys[index+1:], n.keys[index:])
	keys[index] = key
	n.keys = keys

	vals := append(n.children, nil)
	copy(vals[index+1:], n.children[index:])
	vals[index] = val
	n.children = vals
}

func (n *nonleafnode) update(key int, f func([]byte) []byte) bool {
	for i, k := range n.keys {
		if k > key {
			return n.children[i].update(key, f)
		}
	}
	return n.children[len(n.keys)].update(key, f)
}

func (n *leafnode) update(key int, f func([]byte) []byte) bool {
	for i, k := range n.keys {
		if k == key {
			oldVal := n.children[i]
			n.children[i] = f(oldVal)
			return true
		}
		if k > key {
			return false
		}
	}
	return false
}

func (n *nonleafnode) del(key, b int, leftSib, rightSib treenode) bool {
	var ok bool
	for i, k := range n.keys {
		if k > key {
			var leftSib treenode
			if i > 0 {
				leftSib = n.children[i-1]
			}
			ok = n.children[i].del(key, b, leftSib, n.children[i+1])
			goto redistribute
		}
	}
	if len(n.children) > 1 {
		ok = n.children[len(n.keys)].del(key, b, n.children[len(n.keys)-1], nil)
		goto redistribute
	} else {
		ok = n.children[len(n.keys)].del(key, b, nil, nil)
		goto redistribute
	}

redistribute:
	n.updateKeys()

	minChildren := int(math.Ceil(float64(b) / 2))
	if !ok || len(n.children) >= minChildren {
		return ok
	}

	// Try to redistribute
	if leftSib != nil {
		leftSibling := leftSib.(*nonleafnode)
		_, v := leftSibling.borrowValue(false, b)
		if v != nil {
			v.setParent(n)

			n.children = append(n.children, nil)
			copy(n.children[1:], n.children)
			n.children[0] = v

			n.keys = append(n.keys, 0)
			copy(n.keys[1:], n.keys)
			n.keys[0] = n.children[1].firstKey()

			for i, k := range n.parent.keys {
				if k > n.keys[0] {
					n.parent.keys[i] = n.keys[0]
					break
				}
			}

			return true
		}
	}
	if rightSib != nil {
		rightSibling := rightSib.(*nonleafnode)
		_, v := rightSibling.borrowValue(true, b)
		if v != nil {
			v.setParent(n)

			n.keys = append(n.keys, v.firstKey())
			n.children = append(n.children, v)

			for i, k := range n.parent.keys {
				if k == v.firstKey() {
					n.parent.keys[i] = rightSib.firstKey()
				}
			}

			return true
		}
	}

	// If redistribution is not possible, merge
	if leftSib != nil {
		leftSib.concat(n)

		for i, k := range n.parent.keys {
			if k == n.firstKey() {
				n.parent.keys = append(n.parent.keys[:i], n.parent.keys[i+1:]...)
				n.parent.children = append(n.parent.children[:i+1], n.parent.children[i+2:]...)
				break
			}
		}

		return true
	}
	if rightSib != nil {
		n.concat(rightSib)

		for i, k := range n.parent.keys {
			if k == rightSib.firstKey() {
				n.parent.keys = append(n.parent.keys[:i], n.parent.keys[i+1:]...)
				n.parent.children = append(n.parent.children[:i+1], n.parent.children[i+2:]...)
				break
			}
		}

		return true
	}

	// If we haven't returned before here, this is the root
	return true
}

func (n *leafnode) del(key, b int, leftSib, rightSib treenode) bool {
	// If key not present, fail
	for _, k := range n.keys {
		if k == key {
			goto present
		}
		if k > key {
			return false
		}
	}
	return false

present:
	i := 0
	for _, k := range n.keys {
		if k == key {
			break
		}
		i++
	}

	n.keys = append(n.keys[:i], n.keys[i+1:]...)
	n.children = append(n.children[:i], n.children[i+1:]...)

	minChildren := int(math.Ceil(float64(b) / 2))
	if len(n.children) >= minChildren {
		return true
	}

	// Try to redistribute
	if leftSib != nil {
		leftSibling := leftSib.(*leafnode)
		k, v := leftSibling.borrowValue(false, b)
		if v != nil {
			n.keys = append(n.keys, 0)
			copy(n.keys[1:], n.keys)
			n.keys[0] = k

			n.children = append(n.children, nil)
			copy(n.children[1:], n.children)
			n.children[0] = v

			return true
		}
	}
	if rightSib != nil {
		rightSibling := rightSib.(*leafnode)
		k, v := rightSibling.borrowValue(true, b)
		if v != nil {
			n.keys = append(n.keys, k)
			n.children = append(n.children, v)

			return true
		}
	}

	// If redistribution is not possible, merge
	if leftSib != nil {
		leftSib.concat(n)

		for i, k := range n.parent.keys {
			if k == n.firstKey() || k == key {
				n.parent.keys = append(n.parent.keys[:i], n.parent.keys[i+1:]...)
				n.parent.children = append(n.parent.children[:i+1], n.parent.children[i+2:]...)
				break
			}
		}

		return true
	}
	if rightSib != nil {
		n.concat(rightSib)

		for i, k := range n.parent.keys {
			if k == rightSib.firstKey() {
				n.parent.keys = append(n.parent.keys[:i], n.parent.keys[i+1:]...)
				n.parent.children = append(n.parent.children[:i+1], n.parent.children[i+2:]...)
				break
			}
		}

		return true
	}

	// If we haven't returned before here, this is the root
	return true
}

func (n *leafnode) getParent() *nonleafnode {
	return n.parent
}

func (n *nonleafnode) getParent() *nonleafnode {
	return n.parent
}

func (n *leafnode) setParent(p *nonleafnode) {
	n.parent = p
}

func (n *nonleafnode) setParent(p *nonleafnode) {
	n.parent = p
}

func (n *leafnode) borrowValue(left bool, b int) (key int, val []byte) {
	minChildren := int(math.Ceil(float64(b) / 2))
	if len(n.children) > minChildren {
		if left {
			key = n.keys[0]
			val = n.children[0]
			n.keys = n.keys[1:]
			n.children = n.children[1:]
			return
		} else {
			i := len(n.keys) - 1

			key = n.keys[i]
			val = n.children[i]
			n.keys = n.keys[:i]
			n.children = n.children[:i]
			return
		}
	} else {
		return -1, nil
	}
}

func (n *nonleafnode) borrowValue(left bool, b int) (key int, val treenode) {
	minChildren := int(math.Ceil(float64(b) / 2))
	if len(n.children) > minChildren {
		if left {
			key = n.keys[0]
			val = n.children[0]
			n.keys = n.keys[1:]
			n.children = n.children[1:]
			return
		} else {
			i := len(n.keys) - 1

			key = n.keys[i]
			val = n.children[i+1]
			n.keys = n.keys[:i]
			n.children = n.children[:i+1]
			return
		}
	} else {
		return -1, nil
	}
}

func (n *leafnode) firstKey() int {
	return n.keys[0]
}

func (n *nonleafnode) firstKey() int {
	return n.children[0].firstKey()
}

func (l *leafnode) concat(t treenode) {
	r := t.(*leafnode)

	l.keys = append(l.keys, r.keys...)
	l.children = append(l.children, r.children...)

	if r.nextLeaf != nil {
		r.nextLeaf.prevLeaf = l
	}
	l.nextLeaf = r.nextLeaf
}

func (l *nonleafnode) concat(t treenode) {
	r := t.(*nonleafnode)

	l.keys = append(l.keys, r.firstKey())
	l.keys = append(l.keys, r.keys...)
	l.children = append(l.children, r.children...)

	for _, child := range l.children {
		child.setParent(l)
	}
}

func (n *nonleafnode) updateKeys() {
	for i, c := range n.children[1:] {
		n.keys[i] = c.firstKey()
	}
}

package store

// A bptree is a B+ tree implementation of Store.
type bptree struct {
	b    int
	root treenode
}

// Get searches for a value in the B+ tree.
func (t *bptree) Get(key int) []byte {
	return nil
}

// Insert adds a new key-value pair to the tree.
func (t *bptree) Insert(key int, val []byte) bool {
	return false
}

// Update applies the given function to an existing value, or returns false if
// no value is present with the given key.
func (t *bptree) Update(key int, f func([]byte) []byte) bool {
	return false
}

func (t *bptree) Delete(key int) bool {
	return false
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
	get(key int) []byte
}

type nonleafnode struct {
	keys     []int      // List of keys (length b-1)
	children []treenode // Pointers to children (length b)
}

func newNonLeaf(b int) treenode {
	return &nonleafnode{
		make([]int, b-1),
		make([]treenode, b),
	}
}

// A leafnode is a B+ tree leaf. Instead of treenodes as children, it has record
// nodes. Record nodes are just byte slices.
type leafnode struct {
	keys     []int
	children [][]byte
	nextLeaf *leafnode
}

func newLeaf(b int) treenode {
	return &leafnode{
		make([]int, b-1),
		make([][]byte, b-1),
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
		if k < key {
			return nil
		}
	}
	return nil
}

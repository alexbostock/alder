// Package store provides an interface for storing data on disk.
package store

// A Store is a B+ tree structure for storing data on disk.
// There should be one Store for each database table.
type Store interface {
	Get(key int) []byte                                     // Returns a record, or nil to indicate value not present
	Insert(key int, val []byte) bool                        // Insert a new record and return true if successful
	Update(key int, f func([]byte) []byte) bool             // Update an existing record and return true if successful
	Delete(key int) bool                                    // Delete an existing record and return true if successful
	GetRange(minKey, maxKey int) map[int][]byte             // Returns all key-value pairs with keys in inclusive range [minKey,maxKey]
	GetAllWhere(pred func(int, []byte) bool) map[int][]byte // Returns all key-values pairs for which pred is true
}

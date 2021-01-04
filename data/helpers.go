package data

import (
	"fmt"

	"github.com/dgraph-io/badger/v2"
)

// getPrevviousValue returns the existing value for key and parsed the value with fn.
// Returns nil if the key doesn't exist.
func getPreviousValue(txn *badger.Txn, key []byte, fn func(val []byte) error) error {
	item, err := txn.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get previous value for %v: %w", string(key), err)
	}
	if err := item.Value(fn); err != nil {
		return fmt.Errorf("failed to read previous value for %v: %w", string(key), err)
	}
	return nil
}

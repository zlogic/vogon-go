package data

import (
	"fmt"

	"github.com/dgraph-io/badger"
)

func getPreviousValue(txn *badger.Txn, key []byte, fn func(val []byte) error) error {
	item, err := txn.Get(key)
	if err != nil {
		return fmt.Errorf("Failed to get previous value for %v because of %w", string(key), err)
	}
	if err := item.Value(fn); err != nil {
		return fmt.Errorf("Failed to read previous value for %v because of %w", string(key), err)
	}
	return nil
}

package data

import (
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
)

func getPreviousValue(txn *badger.Txn, key []byte, fn func(val []byte) error) error {
	item, err := txn.Get(key)
	if err != nil {
		return errors.Wrapf(err, "Failed to get previous value for %v", string(key))
	}
	if err := item.Value(fn); err != nil {
		return errors.Wrapf(err, "Failed to read previous value for %v %v", string(key), err)
	}
	return nil
}

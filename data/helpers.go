package data

import (
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
)

func getPreviousValue(txn *badger.Txn, key []byte) ([]byte, error) {
	item, err := txn.Get(key)
	if err != nil && err != badger.ErrKeyNotFound {
		return nil, errors.Wrapf(err, "Failed to get previous value for %v", string(key))
	}
	if err == nil {
		value, err := item.Value()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to read previous value for %v %v", string(key), err)
		}
		return value, nil
	}
	return nil, nil
}

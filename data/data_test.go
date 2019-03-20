package data

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
)

var testUser = User{ID: 11}

var testAccount1 = Account{
	ID:             0,
	Name:           "Test 1",
	Currency:       "USD",
	IncludeInTotal: false,
	ShowInList:     true,
}
var testAccount2 = Account{
	ID:             1,
	Name:           "Test 2",
	Currency:       "EUR",
	IncludeInTotal: false,
	ShowInList:     true,
}

func createDb() (dbService *DBService, cleanupFunc func(), err error) {
	dir, err := ioutil.TempDir("", "vogon")
	if err != nil {
		return nil, func() {}, err
	}

	var opts = badger.DefaultOptions
	opts.ValueLogFileSize = 1 << 20
	opts.SyncWrites = false
	opts.Dir = dir
	opts.ValueDir = dir

	dbService, err = Open(opts)
	if err != nil {
		return nil, func() {}, err
	}
	return dbService, func() {
		dbService.Close()
		os.RemoveAll(opts.Dir)
	}, nil
}

func getAllUsers(s *DBService) ([]*User, error) {
	users := make([]*User, 0)
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(UserKeyPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			k := item.Key()

			username, err := DecodeUserKey(k)
			if err != nil {
				return errors.Wrap(err, "Failed to decode username of user")
			}

			v, err := item.Value()
			if err != nil {
				return errors.Wrap(err, "Failed to read value of user")
			}

			user := &User{username: *username}
			err = gob.NewDecoder(bytes.NewBuffer(v)).Decode(&user)
			if err != nil {
				return errors.Wrap(err, "Failed to unmarshal value of user")
			}
			users = append(users, user)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot read users")
	}
	return users, nil
}

func createTestAccounts(s *DBService) error {
	saveUser := testUser
	saveAccount := testAccount1
	if err := s.CreateAccount(&saveUser, &saveAccount); err != nil {
		return err
	}
	saveAccount = testAccount2
	return s.CreateAccount(&saveUser, &saveAccount)
}

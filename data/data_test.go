package data

import (
	"fmt"

	"github.com/dgraph-io/badger/v2"
	log "github.com/sirupsen/logrus"
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

var dbService *DBService

func resetDb() (err error) {
	var opts = badger.DefaultOptions("")
	opts.Logger = log.New()
	opts.ValueLogFileSize = 1 << 20
	opts.InMemory = true

	dbService, err = Open(opts)
	return
}

func getAllUsers(s *DBService) ([]*User, error) {
	users := make([]*User, 0)
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(userKeyPrefix)
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			k := item.Key()

			username, err := decodeUserKey(k)
			if err != nil {
				return fmt.Errorf("failed to decode username from key %v: %w", string(k), err)
			}

			user := &User{}

			if err := item.Value(user.decode); err != nil {
				return fmt.Errorf("failed to read value of user %v: %w", username, err)
			}

			user.username = *username
			users = append(users, user)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot read users: %w", err)
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

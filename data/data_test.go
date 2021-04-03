package data

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/akrylysov/pogreb"
	"github.com/akrylysov/pogreb/fs"
	"github.com/stretchr/testify/assert"
)

var testUser = User{UUID: "uuid11"}

var testAccount1 = Account{
	Name:           "Test 1",
	Currency:       "USD",
	IncludeInTotal: false,
	ShowInList:     true,
}
var testAccount2 = Account{
	Name:           "Test 2",
	Currency:       "EUR",
	IncludeInTotal: false,
	ShowInList:     true,
}

var dbService *DBService

func resetDb() (err error) {
	if dbService != nil {
		it := dbService.db.Items()
		for {
			k, _, err := it.Next()
			if err == pogreb.ErrIterationDone {
				break
			} else if err != nil {
				return err
			}
			err = dbService.db.Delete(k)
			if err != nil {
				return err
			}
		}
		return
	}
	opts := pogreb.Options{FileSystem: fs.Mem}

	dbService, err = Open(opts)
	return
}

func getAllUsers(s *DBService) ([]*User, error) {
	users := make([]*User, 0)
	it := s.db.Items()
	for {
		k, value, err := it.Next()
		if err == pogreb.ErrIterationDone {
			break
		} else if err != nil {
			return nil, err
		}
		if !bytes.HasPrefix(k, []byte(userKeyPrefix)) {
			continue
		}

		username, err := decodeUserKey(k)
		if err != nil {
			return nil, fmt.Errorf("failed to decode username from key %v: %w", string(k), err)
		}

		user := &User{}

		if err := user.decode(value); err != nil {
			return nil, fmt.Errorf("failed to read value of user %v: %w", username, err)
		}

		user.username = *username
		users = append(users, user)
	}
	return users, nil
}

func createTestAccounts(s *DBService) error {
	saveUser := testUser
	saveAccount := testAccount1
	if err := s.CreateAccount(&saveUser, &saveAccount); err != nil {
		return err
	}
	testAccount1.UUID = saveAccount.UUID
	saveAccount = testAccount2
	if err := s.CreateAccount(&saveUser, &saveAccount); err != nil {
		return err
	}
	testAccount2.UUID = saveAccount.UUID
	return nil
}

func assertIndexEquals(t *testing.T, prefix string, expectKeys ...string) {
	index, err := dbService.getReferencedKeys([]byte(testUser.createAccountKeyPrefix()))
	assert.NoError(t, err)
	indexValues := make([]string, len(index))
	for i := range index {
		indexValues[i] = string(index[i])
	}
	assert.ElementsMatch(t, expectKeys, indexValues)
}

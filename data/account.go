package data

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/dgraph-io/badger/v2"
	log "github.com/sirupsen/logrus"
)

// Account keeps the balance and other details for an account.
type Account struct {
	ID             uint64
	Name           string
	Balance        int64
	Currency       string
	IncludeInTotal bool
	ShowInList     bool
}

// encode serializes an Account.
func (account *Account) encode() ([]byte, error) {
	var value bytes.Buffer
	if err := gob.NewEncoder(&value).Encode(account); err != nil {
		return nil, err
	}
	return value.Bytes(), nil
}

// decode deserializes an Account.
func (account *Account) decode(val []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(val)).Decode(account)
}

// createAccount creates and saves the specified account.
// The account ID is not generated here and should be generated before
// calling this method.
func (*DBService) createAccount(user *User, account *Account) func(*badger.Txn) error {
	return func(txn *badger.Txn) error {
		key := user.createAccountKey(account)
		value, err := account.encode()

		if err != nil {
			return fmt.Errorf("cannot encode account: %w", err)
		}

		return txn.Set(key, value)
	}
}

// CreateAccount creates and saves the specified account.
// It generates sets the ID to the generated account ID.
func (s *DBService) CreateAccount(user *User, account *Account) error {
	seq, err := s.db.GetSequence([]byte(user.createSequenceAccountKey()), 1)
	defer seq.Release()
	if err != nil {
		return fmt.Errorf("cannot create account sequence object: %w", err)
	}
	id, err := seq.Next()
	if err != nil {
		return fmt.Errorf("cannot generate id for account: %w", err)
	}

	account.ID = id
	account.Balance = 0

	return s.db.Update(func(txn *badger.Txn) error {
		return s.createAccount(user, account)(txn)
	})
}

// UpdateAccount saves an already existing account.
// If the account doesn't exist, it returns an error.
func (s *DBService) UpdateAccount(user *User, account *Account) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := user.createAccountKey(account)

		previousAccount := &Account{}
		if err := getPreviousValue(txn, key, previousAccount.decode); err != nil {
			if err == badger.ErrKeyNotFound {
				return fmt.Errorf("cannot update account %v if it doesn't exist", string(key))
			}
			return fmt.Errorf("cannot get previous value for account %v: %w", string(key), err)
		}
		account.Balance = previousAccount.Balance
		if account == previousAccount {
			log.WithField("key", string(key)).Debug("Account is unchanged")
			return nil
		}

		value, err := account.encode()
		if err != nil {
			return fmt.Errorf("cannot encode account: %w", err)
		}
		return txn.Set(key, value)
	})
}

// updateAccountBalance updates the account balance by delta.
func (s *DBService) updateAccountBalance(user *User, accountID uint64, deltaBalance int64) func(*badger.Txn) error {
	return func(txn *badger.Txn) error {
		if deltaBalance == 0 {
			return nil
		}
		key := user.createAccountKeyFromID(accountID)

		account := &Account{}
		if err := getPreviousValue(txn, key, account.decode); err != nil {

			if err == badger.ErrKeyNotFound {
				return fmt.Errorf("cannot update account %v if it doesn't exist", string(key))
			}
			return fmt.Errorf("cannot get previous value for account %v: %w", string(key), err)
		}

		account.Balance += deltaBalance

		value, err := account.encode()
		if err != nil {
			return fmt.Errorf("cannot encode account: %w", err)
		}
		return txn.Set(key, value)
	}
}

// getAccounts returns all accounts for user.
func (s *DBService) getAccounts(user *User) func(*badger.Txn) ([]*Account, error) {
	return func(txn *badger.Txn) ([]*Account, error) {
		accounts := make([]*Account, 0)

		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(user.createAccountKeyPrefix())
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			k := item.Key()

			account := &Account{}
			if err := item.Value(account.decode); err != nil {
				return nil, fmt.Errorf("failed to read value for account %v: %w", string(k), err)
			}
			accounts = append(accounts, account)
		}
		return accounts, nil
	}
}

// GetAccount returns an Account by its ID.
// If the Account doesn't exist, it returns an error.
func (s *DBService) GetAccount(user *User, accountID uint64) (*Account, error) {
	account := &Account{}

	key := user.createAccountKeyFromID(accountID)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return fmt.Errorf("failed to get account %v: %w", string(key), err)
		}

		if err := item.Value(account.decode); err != nil {
			return fmt.Errorf("failed to read value for account %v: %w", string(key), err)
		}

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get account %v: %w", accountID, err)
	}
	return account, nil
}

// GetAccounts returns all accounts for user.
func (s *DBService) GetAccounts(user *User) ([]*Account, error) {
	var accounts []*Account
	err := s.db.View(func(txn *badger.Txn) error {
		var err error
		accounts, err = s.getAccounts(user)(txn)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	return accounts, nil
}

// DeleteAccount deletes an account by its ID.
// If the account doesn't exist, it returns an error.
func (s *DBService) DeleteAccount(user *User, accountID uint64) error {
	key := user.createAccountKeyFromID(accountID)
	return s.db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(key); err != nil {
			return fmt.Errorf("cannot delete account %v because it doesn't exist: %w", accountID, err)
		}

		return txn.Delete(key)
	})
}

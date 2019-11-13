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

// Encode serializes an Account.
func (account *Account) Encode() ([]byte, error) {
	var value bytes.Buffer
	if err := gob.NewEncoder(&value).Encode(account); err != nil {
		return nil, err
	}
	return value.Bytes(), nil
}

// Decode deserializes an Account.
func (account *Account) Decode(val []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(val)).Decode(account)
}

func (*DBService) createAccount(user *User, account *Account) func(*badger.Txn) error {
	return func(txn *badger.Txn) error {
		key := user.CreateAccountKey(account)
		value, err := account.Encode()

		if err != nil {
			return fmt.Errorf("Cannot encode account because of %w", err)
		}

		return txn.Set(key, value)
	}
}

// CreateAccount creates and saves the specified account.
// It generates sets the ID to the generated account ID.
func (s *DBService) CreateAccount(user *User, account *Account) error {
	seq, err := s.db.GetSequence([]byte(user.CreateSequenceAccountKey()), 1)
	defer seq.Release()
	if err != nil {
		return fmt.Errorf("Cannot create account sequence object because of %w", err)
	}
	id, err := seq.Next()
	if err != nil {
		return fmt.Errorf("Cannot generate id for account because of %w", err)
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
		key := user.CreateAccountKey(account)

		previousAccount := &Account{}
		if err := getPreviousValue(txn, key, previousAccount.Decode); err != nil {
			if err == badger.ErrKeyNotFound {
				log.WithField("key", key).Error("Cannot update account if it doesn't exist")
				return fmt.Errorf("Cannot update account if it doesn't exist")
			}
			log.WithField("key", key).WithError(err).Error("Cannot get previous value for account")
			return err
		}
		account.Balance = previousAccount.Balance
		if account == previousAccount {
			log.WithField("key", key).Debug("Account is unchanged")
			return nil
		}

		value, err := account.Encode()
		if err != nil {
			return fmt.Errorf("Cannot encode account because of %w", err)
		}
		return txn.Set(key, value)
	})
}

func (s *DBService) updateAccountBalance(user *User, accountID uint64, deltaBalance int64) func(*badger.Txn) error {
	return func(txn *badger.Txn) error {
		if deltaBalance == 0 {
			return nil
		}
		key := user.CreateAccountKeyFromID(accountID)

		account := &Account{}
		if err := getPreviousValue(txn, key, account.Decode); err != nil {

			if err == badger.ErrKeyNotFound {
				log.WithField("key", key).Error("Cannot update account if it doesn't exist")
			}
			log.WithField("key", key).WithError(err).Error("Cannot get previous value for account")
			return err
		}

		account.Balance += deltaBalance

		value, err := account.Encode()
		if err != nil {
			return fmt.Errorf("Cannot encode account because of %w", err)
		}
		return txn.Set(key, value)
	}
}

func (s *DBService) getAccounts(user *User) func(*badger.Txn) ([]*Account, error) {
	return func(txn *badger.Txn) ([]*Account, error) {
		accounts := make([]*Account, 0)

		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(user.CreateAccountKeyPrefix())
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			k := item.Key()

			account := &Account{}
			if err := item.Value(account.Decode); err != nil {
				log.WithField("key", k).WithError(err).Error("Failed to read value of account")
				continue
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

	key := user.CreateAccountKeyFromID(accountID)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return fmt.Errorf("Failed to get account %v because of %w", string(key), err)
		}

		if err := item.Value(account.Decode); err != nil {
			return fmt.Errorf("Failed to read value for account %v because of %w", string(key), err)
		}

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to get account %v because of %w", accountID, err)
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
		return nil, fmt.Errorf("Failed to get accounts because of %w", err)
	}
	return accounts, nil
}

// DeleteAccount deletes an account by its ID.
// If the account doesn't exist, it returns an error.
func (s *DBService) DeleteAccount(user *User, accountID uint64) error {
	key := user.CreateAccountKeyFromID(accountID)
	return s.db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(key); err != nil {
			log.WithField("key", key).WithError(err).Error("Cannot delete account if it doesn't exist")
			return err
		}

		return txn.Delete(key)
	})
}

package data

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Account struct {
	ID             uint64
	Name           string
	Balance        int64
	Currency       string
	IncludeInTotal bool
	ShowInList     bool
}

func (account *Account) Encode() ([]byte, error) {
	var value bytes.Buffer
	if err := gob.NewEncoder(&value).Encode(account); err != nil {
		return nil, err
	}
	return value.Bytes(), nil
}

func (*DBService) createAccount(user *User, account *Account) func(*badger.Txn) error {
	return func(txn *badger.Txn) error {
		key := user.CreateAccountKey(account)
		value, err := account.Encode()

		if err != nil {
			return errors.Wrap(err, "Cannot encode account")
		}

		return txn.Set(key, value)
	}
}

func (s *DBService) CreateAccount(user *User, account *Account) error {
	seq, err := s.db.GetSequence([]byte(user.CreateSequenceAccountKey()), 1)
	defer seq.Release()
	if err != nil {
		return errors.Wrap(err, "Cannot create account sequence object")
	}
	id, err := seq.Next()
	if err != nil {
		return errors.Wrap(err, "Cannot generate id for account")
	}

	account.ID = id
	account.Balance = 0

	return s.db.Update(func(txn *badger.Txn) error {
		return s.createAccount(user, account)(txn)
	})
}

func (s *DBService) UpdateAccount(user *User, account *Account) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := user.CreateAccountKey(account)

		previousValue, err := getPreviousValue(txn, key)
		if err != nil {
			log.WithField("key", key).WithError(err).Error("Cannot get previous value for account")
			return err
		}

		if previousValue == nil {
			log.WithField("key", key).Error("Cannot update account if it doesn't exist")
			return fmt.Errorf("Cannot update account if it doesn't exist")
		}

		previousAccount := &Account{}
		if err := gob.NewDecoder(bytes.NewBuffer(previousValue)).Decode(previousAccount); err != nil {
			log.WithField("key", key).WithError(err).Error("Failed to read previous value of account")
			return err
		}
		account.Balance = previousAccount.Balance
		if account == previousAccount {
			log.WithField("key", key).Debug("Account is unchanged")
			return nil
		}

		value, err := account.Encode()
		if err != nil {
			return errors.Wrap(err, "Cannot encode account")
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

		previousValue, err := getPreviousValue(txn, key)
		if err != nil {
			log.WithField("key", key).WithError(err).Error("Cannot get previous value for account")
			return err
		}

		if previousValue == nil {
			log.WithField("key", key).Error("Cannot update account if it doesn't exist")
			return nil
		}

		account := &Account{}
		if err := gob.NewDecoder(bytes.NewBuffer(previousValue)).Decode(account); err != nil {
			log.WithField("key", key).WithError(err).Error("Failed to read previous value of account")
			return err
		}

		account.Balance += deltaBalance

		value, err := account.Encode()
		if err != nil {
			return errors.Wrap(err, "Cannot encode account")
		}
		return txn.Set(key, value)
	}
}

func (s *DBService) getAccounts(user *User) func(*badger.Txn) ([]*Account, error) {
	return func(txn *badger.Txn) ([]*Account, error) {
		accounts := make([]*Account, 0)

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(user.CreateAccountKeyPrefix())
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			k := item.Key()
			v, err := item.Value()
			if err != nil {
				log.WithField("key", k).WithError(err).Error("Failed to read value of account")
				continue
			}
			account := &Account{}

			if err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(account); err != nil {
				log.WithField("key", k).WithError(err).Error("Failed to decode value of account")
				return nil, err
			}
			accounts = append(accounts, account)
		}
		return accounts, nil
	}
}

func (s *DBService) GetAccount(user *User, accountID uint64) (*Account, error) {
	var account *Account

	key := user.CreateAccountKeyFromID(accountID)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return errors.Wrapf(err, "Failed to get account %v", string(key))
		}
		v, err := item.Value()
		if err != nil {
			return errors.Wrapf(err, "Failed to get value for account %v", string(key))
		}

		account = &Account{}
		if err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(account); err != nil {
			return errors.Wrapf(err, "Failed to decode value for account %v", string(key))
		}
		return err
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get account %v", accountID)
	}
	return account, nil
}

func (s *DBService) GetAccounts(user *User) ([]*Account, error) {
	var accounts []*Account
	err := s.db.View(func(txn *badger.Txn) error {
		var err error
		accounts, err = s.getAccounts(user)(txn)
		return err
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get accounts")
	}
	return accounts, nil
}

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

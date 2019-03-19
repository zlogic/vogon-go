package data

import (
	"bytes"
	"encoding/gob"

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

func (s *DBService) CreateAccount(user *User, account *Account) error {
	seq, err := s.db.GetSequence([]byte(CreateSequenceAccountKey(user)), 1)
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

	accountKey := user.CreateAccountKey(account)
	value, err := account.Encode()
	if err != nil {
		return errors.Wrap(err, "Cannot encode account")
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(accountKey, value)
	})
}

func (s *DBService) UpdateAccount(user *User, account *Account) error {
	return s.db.Update(func(txn *badger.Txn) error {
		accountKey := user.CreateAccountKey(account)

		previousValue, err := getPreviousValue(txn, accountKey)
		if err != nil {
			log.WithField("key", accountKey).WithError(err).Error("Cannot get previous value for account")
			return err
		}

		if previousValue == nil {
			log.WithField("key", accountKey).WithError(err).Error("Cannot update account if it doesn't exist")
		}

		previousAccount := &Account{}
		if err := gob.NewDecoder(bytes.NewBuffer(previousValue)).Decode(previousAccount); err != nil {
			log.WithField("key", accountKey).WithError(err).Error("Failed to read previous value of account")
			return err
		}
		account.Balance = previousAccount.Balance
		if account == previousAccount {
			log.WithField("key", accountKey).Debug("Account is unchanged")
			return nil
		}

		value, err := account.Encode()
		if err != nil {
			return errors.Wrap(err, "Cannot encode account")
		}
		return txn.Set(accountKey, value)
	})
}

func (s *DBService) GetAccounts(user *User) ([]*Account, error) {
	accounts := make([]*Account, 0)
	err := s.db.View(func(txn *badger.Txn) error {
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
				return err
			}
			accounts = append(accounts, account)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get accounts")
	}
	return accounts, nil
}

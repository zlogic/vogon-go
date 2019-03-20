package data

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	TransactionTypeExpenseIncome = iota
	TransactionTypeTransfer      = iota
)

type Transaction struct {
	ID          uint64
	Description string
	Type        int
	Tags        []string
	Date        string
	Components  []TransactionComponent
}

type TransactionComponent struct {
	Amount    int64
	AccountID uint64
}

func (transaction *Transaction) Encode() ([]byte, error) {
	var value bytes.Buffer
	if err := gob.NewEncoder(&value).Encode(transaction); err != nil {
		return nil, err
	}
	return value.Bytes(), nil
}

func (s *DBService) CreateTransaction(user *User, transaction *Transaction) error {
	seq, err := s.db.GetSequence([]byte(CreateSequenceTransactionKey(user)), 1)
	defer seq.Release()
	if err != nil {
		return errors.Wrap(err, "Cannot create transaction sequence object")
	}
	id, err := seq.Next()
	if err != nil {
		return errors.Wrap(err, "Cannot generate id for transaction")
	}
	transaction.ID = id

	key := user.CreateTransactionKey(transaction)
	value, err := transaction.Encode()
	if err != nil {
		return errors.Wrap(err, "Cannot encode transaction")
	}

	return s.db.Update(func(txn *badger.Txn) error {
		if err := s.updateAccountsBalance(user, nil, &transaction.Components)(txn); err != nil {
			return errors.Wrap(err, "Cannot update account balance")
		}

		return txn.Set(key, value)
	})
}

func (s *DBService) UpdateTransaction(user *User, transaction *Transaction) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := user.CreateTransactionKey(transaction)

		previousValue, err := getPreviousValue(txn, key)
		if err != nil {
			log.WithField("key", key).WithError(err).Error("Cannot get previous value for transaction")
			return err
		}

		if previousValue == nil {
			log.WithField("key", key).Error("Cannot update transaction if it doesn't exist")
			return fmt.Errorf("Cannot update transaction if it doesn't exist")
		}

		previousTransaction := &Transaction{}
		if err := gob.NewDecoder(bytes.NewBuffer(previousValue)).Decode(previousTransaction); err != nil {
			log.WithField("key", key).WithError(err).Error("Failed to read previous value of transaction")
			return err
		}
		if transaction == previousTransaction {
			log.WithField("key", key).Debug("Transaction is unchanged")
			return nil
		}

		if err := s.updateAccountsBalance(user, &previousTransaction.Components, &transaction.Components)(txn); err != nil {
			return errors.Wrap(err, "Cannot update account balance")
		}

		value, err := transaction.Encode()
		if err != nil {
			return errors.Wrap(err, "Cannot encode transaction")
		}
		return txn.Set(key, value)
	})
}

func (s *DBService) updateAccountsBalance(user *User, previousComponents *[]TransactionComponent, newComponents *[]TransactionComponent) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		accountDeltas := make(map[uint64]int64)
		if previousComponents != nil {
			for _, component := range *previousComponents {
				accountDeltas[component.AccountID] = accountDeltas[component.AccountID] - component.Amount
			}
		}
		if newComponents != nil {
			for _, component := range *newComponents {
				accountDeltas[component.AccountID] = accountDeltas[component.AccountID] + component.Amount
			}
		}
		for accountID, deltaAmount := range accountDeltas {
			if err := s.updateAccountBalance(user, accountID, deltaAmount)(txn); err != nil {
				return errors.Wrap(err, "Cannot update account balance")
			}
		}
		return nil
	}
}

func (s *DBService) GetTransactions(user *User) ([]*Transaction, error) {
	transactions := make([]*Transaction, 0)
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(user.CreateTransactionKeyPrefix())
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			k := item.Key()
			v, err := item.Value()
			if err != nil {
				log.WithField("key", k).WithError(err).Error("Failed to read value of transaction")
				continue
			}
			transaction := &Transaction{}

			if err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(transaction); err != nil {
				log.WithField("key", k).WithError(err).Error("Failed to decode value of transaction")
				return err
			}
			transactions = append(transactions, transaction)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get transactions")
	}
	return transactions, nil
}

func (s *DBService) DeleteTransaction(user *User, transaction *Transaction) error {
	key := user.CreateTransactionKey(transaction)
	return s.db.Update(func(txn *badger.Txn) error {
		previousValue, err := getPreviousValue(txn, key)
		if err != nil {
			log.WithField("key", key).WithError(err).Error("Cannot get value for deleted transaction")
			return err
		}

		if previousValue == nil {
			log.WithField("key", key).Error("Cannot delete transaction if it doesn't exist")
			return fmt.Errorf("Cannot delete non-existing transaction")
		}

		deleteTransaction := &Transaction{}
		if err := gob.NewDecoder(bytes.NewBuffer(previousValue)).Decode(deleteTransaction); err != nil {
			log.WithField("key", key).WithError(err).Error("Failed to read value of deleted transaction")
			return err
		}

		if err := s.updateAccountsBalance(user, &deleteTransaction.Components, nil)(txn); err != nil {
			return errors.Wrap(err, "Cannot update account balance")
		}

		return txn.Delete(key)
	})
}

package data

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"
	"sync"
	"time"

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

const DateFormat = "2006-01-02"
const inputDateFormat = "2006-1-2"

func (transaction *Transaction) Normalize() error {
	date, err := time.Parse(inputDateFormat, transaction.Date)
	if err != nil {
		return errors.Wrapf(err, "Cannot parse date "+transaction.Date)
	}

	transaction.Date = date.Format(DateFormat)

	if len(transaction.Tags) > 0 {
		filteredTags := make([]string, 0, len(transaction.Tags))
		for _, tag := range transaction.Tags {
			duplicate := false
			if tag == "" {
				continue
			}
			for _, filteredTag := range filteredTags {
				if filteredTag == tag {
					duplicate = true
					break
				}
			}
			if !duplicate {
				filteredTags = append(filteredTags, tag)
			}
		}
		sort.Strings(filteredTags)
		transaction.Tags = filteredTags
	}
	return nil
}

func sortTransactionsAsc(transactions []*Transaction) {
	sort.Slice(transactions, func(i, j int) bool {
		if transactions[i].Date != transactions[j].Date {
			return transactions[i].Date < transactions[j].Date
		} else {
			return transactions[i].ID < transactions[j].ID
		}
	})
}

func (s *DBService) createTransaction(user *User, transaction *Transaction) func(*badger.Txn) error {
	return func(txn *badger.Txn) error {
		key := user.CreateTransactionKey(transaction)
		value, err := transaction.Encode()
		if err != nil {
			return errors.Wrap(err, "Cannot encode transaction")
		}

		if err := s.updateAccountsBalance(user, nil, &transaction.Components)(txn); err != nil {
			return errors.Wrap(err, "Cannot update account balance")
		}

		return txn.Set(key, value)
	}
}
func (s *DBService) CreateTransaction(user *User, transaction *Transaction) error {
	seq, err := s.db.GetSequence([]byte(user.CreateSequenceTransactionKey()), 1)
	defer seq.Release()
	if err != nil {
		return errors.Wrap(err, "Cannot create transaction sequence object")
	}
	id, err := seq.Next()
	if err != nil {
		return errors.Wrap(err, "Cannot generate id for transaction")
	}
	transaction.ID = id

	return s.db.Update(func(txn *badger.Txn) error {
		return s.createTransaction(user, transaction)(txn)
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

func (s *DBService) updateAccountsBalance(user *User, previousComponents *[]TransactionComponent, newComponents *[]TransactionComponent) func(*badger.Txn) error {
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

func (s *DBService) getTransactions(user *User) func(*badger.Txn) ([]*Transaction, error) {
	return func(txn *badger.Txn) ([]*Transaction, error) {
		transactions := make([]*Transaction, 0)

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(user.CreateTransactionKeyPrefix())

		type kv struct {
			k []byte
			v []byte
		}

		kvc := make(chan *kv, 16)
		tc := make(chan *Transaction, 16)

		var wg sync.WaitGroup
		var failedErr error

		for i := 0; i < 16; i++ {
			go func() {
				for kv := range kvc {
					transaction := &Transaction{}
					if err := gob.NewDecoder(bytes.NewBuffer(kv.v)).Decode(transaction); err != nil {
						log.WithField("key", kv.k).WithError(err).Error("Failed to decode value of transaction")
						failedErr = err
						tc <- nil
						return
					}
					tc <- transaction
				}
			}()
		}

		go func() {
			for transaction := range tc {
				transactions = append(transactions, transaction)
				wg.Done()
			}
			close(tc)
			close(kvc)
		}()

		clone := func(data []byte) []byte {
			clone := make([]byte, len(data))
			copy(clone, data)
			return clone
		}

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if failedErr != nil {
				break
			}

			item := it.Item()

			k := item.Key()
			v, err := item.Value()
			if err != nil {
				log.WithField("key", k).WithError(err).Error("Failed to read value of transaction")
				continue
			}

			wg.Add(1)
			kvc <- &kv{k: clone(k), v: clone(v)}
		}

		wg.Wait()
		if failedErr != nil {
			return nil, failedErr
		}

		return transactions, nil
	}
}

func (s *DBService) GetTransactions(user *User) ([]*Transaction, error) {
	var transactions []*Transaction

	err := s.db.View(func(txn *badger.Txn) error {
		var err error
		transactions, err = s.getTransactions(user)(txn)
		return err
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

package data

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"
	"strings"
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

type TransactionFilterOptions struct {
	FilterDescription    string
	FilterFromDate       string
	FilterToDate         string
	FilterTags           []string
	FilterAccounts       []uint64
	ExcludeExpenseIncome bool
	ExcludeTransfer      bool
}

type GetTransactionOptions struct {
	Offset uint64
	Limit  uint64
	TransactionFilterOptions
}

var GetAllTransactionsOptions = GetTransactionOptions{Offset: 0, Limit: ^uint64(0)}

func IteratorDoNotPrefetchOptions() badger.IteratorOptions {
	options := badger.DefaultIteratorOptions
	options.PrefetchValues = false
	return options
}

func IteratorIndexOptions() badger.IteratorOptions {
	options := badger.DefaultIteratorOptions
	options.PrefetchValues = false
	options.Reverse = true
	return options
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

		index := user.CreateTransactionIndexKey(transaction)
		if err := txn.Set(index, nil); err != nil {
			return errors.Wrap(err, "Cannot create index for transaction")
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
		index := user.CreateTransactionIndexKey(transaction)
		if err := txn.Set(index, nil); err != nil {
			return errors.Wrap(err, "Cannot create index for transaction")
		}

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

		previousTransactionIndexKey := user.CreateTransactionIndexKey(previousTransaction)
		transactionIndexKey := user.CreateTransactionIndexKey(transaction)
		if !bytes.Equal(previousTransactionIndexKey, transactionIndexKey) {
			if err := txn.Delete(previousTransactionIndexKey); err != nil {
				return errors.Wrap(err, "Cannot delete previous index for transaction")
			}
			if err := txn.Set(transactionIndexKey, nil); err != nil {
				return errors.Wrap(err, "Cannot create index for transaction")
			}
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

func (options *TransactionFilterOptions) IsEmpty() bool {
	return options.FilterDescription == "" &&
		options.FilterFromDate == "" &&
		options.FilterToDate == "" &&
		len(options.FilterTags) == 0 &&
		len(options.FilterAccounts) == 0 &&
		options.ExcludeExpenseIncome == false &&
		options.ExcludeTransfer == false
}

func (options *TransactionFilterOptions) Matches(transaction *Transaction) bool {
	containsTag := func(searchIn []string, searchFor []string) bool {
		for _, a := range searchIn {
			for _, b := range searchFor {
				if a == b {
					return true
				}
			}
		}
		return false
	}
	containsAccount := func(searchIn *Transaction, searchFor []uint64) bool {
		for _, a := range searchIn.Components {
			for _, b := range searchFor {
				if a.AccountID == b {
					return true
				}
			}
		}
		return false
	}
	matchesDescription := options.FilterDescription == "" || strings.Contains(strings.ToLower(transaction.Description), strings.ToLower(options.FilterDescription))
	matchedFromDate := options.FilterFromDate == "" || options.FilterFromDate <= transaction.Date
	matchedToDate := options.FilterToDate == "" || transaction.Date <= options.FilterToDate
	matchesTags := len(options.FilterTags) == 0 || containsTag(transaction.Tags, options.FilterTags)
	matchesAccounts := len(options.FilterAccounts) == 0 || containsAccount(transaction, options.FilterAccounts)
	matchesType := (transaction.Type == TransactionTypeExpenseIncome && !options.ExcludeExpenseIncome) ||
		(transaction.Type == TransactionTypeTransfer && !options.ExcludeTransfer)
	return matchesDescription &&
		matchedFromDate && matchedToDate &&
		matchesTags &&
		matchesAccounts &&
		matchesType
}

func (s *DBService) getTransaction(user *User, transactionID uint64) func(*badger.Txn) (*Transaction, error) {
	return func(txn *badger.Txn) (*Transaction, error) {
		transactionKey := user.CreateTransactionKeyFromID(transactionID)
		item, err := txn.Get(transactionKey)
		if err != nil {
			log.WithField("key", transactionKey).WithError(err).Error("Failed to get transaction")
			return nil, err
		}

		k := item.Key()
		v, err := item.Value()
		if err != nil {
			log.WithField("key", k).WithError(err).Error("Failed to read value of transaction")
			return nil, err
		}

		transaction := &Transaction{}
		if err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(transaction); err != nil {
			log.WithField("key", k).WithError(err).Error("Failed to decode value of transaction")
			return nil, err
		}
		return transaction, nil
	}
}

func (s *DBService) getTransactions(user *User, options GetTransactionOptions) func(*badger.Txn) ([]*Transaction, error) {
	return func(txn *badger.Txn) ([]*Transaction, error) {
		transactions := make([]*Transaction, 0)

		it := txn.NewIterator(IteratorIndexOptions())
		defer it.Close()
		prefix := []byte(user.CreateTransactionIndexKeyPrefix())
		reversePrefix := append([]byte(user.CreateTransactionIndexKeyPrefix()), 0xff)

		var currentItem uint64
		skipItem := func() bool {
			currentItem++
			return currentItem < (options.Offset + 1)
		}
		emptyFilter := options.TransactionFilterOptions.IsEmpty()
		for it.Seek(reversePrefix); it.ValidForPrefix(prefix); it.Next() {
			if emptyFilter && skipItem() {
				continue
			}
			transactionID, err := user.DecodeTransactionIndexKey(it.Item().Key())
			if err != nil {
				log.WithError(err).Error("Failed parse transaction index")
				return nil, err
			}

			transaction, err := s.getTransaction(user, transactionID)(txn)
			if err != nil {
				return nil, err
			}

			if !emptyFilter {
				if !options.TransactionFilterOptions.Matches(transaction) {
					continue
				}
				if skipItem() {
					continue
				}
			}

			transactions = append(transactions, transaction)
			if uint64(len(transactions)) >= options.Limit {
				break
			}
		}

		return transactions, nil
	}
}

func (s *DBService) GetTransaction(user *User, transactionID uint64) (*Transaction, error) {
	var transaction *Transaction

	err := s.db.View(func(txn *badger.Txn) error {
		var err error
		transaction, err = s.getTransaction(user, transactionID)(txn)
		return err
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get transaction %v", transactionID)
	}
	return transaction, nil
}

func (s *DBService) GetTransactions(user *User, options GetTransactionOptions) ([]*Transaction, error) {
	var transactions []*Transaction

	err := s.db.View(func(txn *badger.Txn) error {
		var err error
		transactions, err = s.getTransactions(user, options)(txn)
		return err
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get transactions")
	}
	return transactions, nil
}

func (s *DBService) CountTransactions(user *User, options TransactionFilterOptions) (uint64, error) {
	var count uint64

	emptyFilter := options.IsEmpty()
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(IteratorDoNotPrefetchOptions())
		defer it.Close()
		prefix := []byte(user.CreateTransactionIndexKeyPrefix())

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if !emptyFilter {
				transactionID, err := user.DecodeTransactionIndexKey(it.Item().Key())
				if err != nil {
					log.WithError(err).Error("Failed to parse transaction index")
					return err
				}

				transaction, err := s.getTransaction(user, transactionID)(txn)
				if err != nil {
					return err
				}

				if !options.Matches(transaction) {
					continue
				}
			}
			count++
		}
		return nil
	})
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to count transactions")
	}
	return count, nil
}

func (s *DBService) DeleteTransaction(user *User, transactionID uint64) error {
	key := user.CreateTransactionKeyFromID(transactionID)
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

		transactionIndexKey := user.CreateTransactionIndexKey(deleteTransaction)
		if err := txn.Delete(transactionIndexKey); err != nil {
			return errors.Wrap(err, "Failed to delete transaction index")
		}

		return txn.Delete(key)
	})
}

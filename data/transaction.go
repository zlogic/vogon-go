package data

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v2"
	log "github.com/sirupsen/logrus"
)

const (
	// TransactionTypeExpenseIncome is an Expense/Income transaction (totals do not have to be equal).
	TransactionTypeExpenseIncome = iota
	// TransactionTypeTransfer is a Transfer transaction (totals for each currency should be zero).
	TransactionTypeTransfer
)

// Transaction saves details for one expense item.
type Transaction struct {
	ID          uint64
	Description string
	Type        int
	Tags        []string
	Date        string
	Components  []TransactionComponent
}

// TransactionComponent contains details for a part of a Transaction.
type TransactionComponent struct {
	Amount    int64
	AccountID uint64
}

// TransactionFilterOptions specifies filter parameters for Transactions.
type TransactionFilterOptions struct {
	FilterDescription    string
	FilterFromDate       string
	FilterToDate         string
	FilterTags           []string
	FilterAccounts       []uint64
	ExcludeExpenseIncome bool
	ExcludeTransfer      bool
}

// GetTransactionOptions specifies paging and filtering options for retrieving transactions.
type GetTransactionOptions struct {
	Offset uint64
	Limit  uint64
	TransactionFilterOptions
}

// GetAllTransactionsOptions is a GetTransactionOptions which returns all transactions in one page.
var GetAllTransactionsOptions = GetTransactionOptions{Offset: 0, Limit: ^uint64(0)}

// encode serializes a Transaction.
func (transaction *Transaction) encode() ([]byte, error) {
	var value bytes.Buffer
	if err := gob.NewEncoder(&value).Encode(transaction); err != nil {
		return nil, err
	}
	return value.Bytes(), nil
}

// decode deserializes a Transaction.
func (transaction *Transaction) decode(val []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(val)).Decode(transaction)
}

// dateFormat is the format used to serialize the Transaction Date.
const dateFormat = "2006-01-02"

// inputDateFormat is the format used to decode the Transaction Date.
const inputDateFormat = "2006-1-2"

// normalize reformats the date to a common format and sorts/deduplicates transaction tags.
func (transaction *Transaction) normalize() error {
	date, err := time.Parse(inputDateFormat, transaction.Date)
	if err != nil {
		return fmt.Errorf("cannot parse date %v: %w", transaction.Date, err)
	}

	transaction.Date = date.Format(dateFormat)

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

// sortTransactionsAsc sorts a Transactions slice by date and id ascending.
func sortTransactionsAsc(transactions []*Transaction) {
	sort.Slice(transactions, func(i, j int) bool {
		if transactions[i].Date != transactions[j].Date {
			return transactions[i].Date < transactions[j].Date
		}
		return transactions[i].ID < transactions[j].ID
	})
}

// createTransaction returns a function to create a transaction for user.
func (s *DBService) createTransaction(user *User, transaction *Transaction) func(*badger.Txn) error {
	return func(txn *badger.Txn) error {
		key := user.createTransactionKey(transaction)
		value, err := transaction.encode()
		if err != nil {
			return fmt.Errorf("cannot encode transaction: %w", err)
		}

		if err := s.updateAccountsBalance(user, nil, &transaction.Components)(txn); err != nil {
			return fmt.Errorf("cannot update account balance: %w", err)
		}

		index := user.createTransactionIndexKey(transaction)
		if err := txn.Set(index, nil); err != nil {
			return fmt.Errorf("cannot create index for transaction: %w", err)
		}

		return txn.Set(key, value)
	}
}

// CreateTransaction saves a new Transaction into the database.
func (s *DBService) CreateTransaction(user *User, transaction *Transaction) error {
	seq, err := s.db.GetSequence([]byte(user.createSequenceTransactionKey()), 1)
	defer seq.Release()
	if err != nil {
		return fmt.Errorf("cannot create transaction sequence object: %w", err)
	}
	id, err := seq.Next()
	if err != nil {
		return fmt.Errorf("cannot generate id for transaction: %w", err)
	}
	transaction.ID = id

	return s.db.Update(func(txn *badger.Txn) error {
		index := user.createTransactionIndexKey(transaction)
		if err := txn.Set(index, nil); err != nil {
			return fmt.Errorf("cannot create index for transaction: %w", err)
		}

		return s.createTransaction(user, transaction)(txn)
	})
}

// UpdateTransaction updates an existing Transaction in the database.
func (s *DBService) UpdateTransaction(user *User, transaction *Transaction) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := user.createTransactionKey(transaction)

		previousTransaction := &Transaction{}
		if err := getPreviousValue(txn, key, previousTransaction.decode); err != nil {
			if err == badger.ErrKeyNotFound {
				return fmt.Errorf("cannot update transaction %v if it doesn't exist", string(key))
			}
			return fmt.Errorf("cannot get previous value for transaction %v: %w", string(key), err)
		}
		if transaction == previousTransaction {
			log.WithField("key", string(key)).Debug("Transaction is unchanged")
			return nil
		}

		if err := s.updateAccountsBalance(user, &previousTransaction.Components, &transaction.Components)(txn); err != nil {
			return fmt.Errorf("cannot update account balance: %w", err)
		}

		previousTransactionIndexKey := user.createTransactionIndexKey(previousTransaction)
		transactionIndexKey := user.createTransactionIndexKey(transaction)
		if !bytes.Equal(previousTransactionIndexKey, transactionIndexKey) {
			if err := txn.Delete(previousTransactionIndexKey); err != nil {
				return fmt.Errorf("cannot delete previous index for transaction %v: %w", string(previousTransactionIndexKey), err)
			}
			if err := txn.Set(transactionIndexKey, nil); err != nil {
				return fmt.Errorf("cannot create index for transaction %v: %w", string(transactionIndexKey), err)
			}
		}

		value, err := transaction.encode()
		if err != nil {
			return fmt.Errorf("cannot encode transaction: %w", err)
		}
		return txn.Set(key, value)
	})
}

// updateAccountsBalance returns a function to update account balance.
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
				return fmt.Errorf("cannot update account balance: %w", err)
			}
		}
		return nil
	}
}

// IsEmpty returns if options do not apply any filtering (all transactions match this filter).
func (options *TransactionFilterOptions) IsEmpty() bool {
	return options.FilterDescription == "" &&
		options.FilterFromDate == "" &&
		options.FilterToDate == "" &&
		len(options.FilterTags) == 0 &&
		len(options.FilterAccounts) == 0 &&
		options.ExcludeExpenseIncome == false &&
		options.ExcludeTransfer == false
}

// Matches returns true if transaction is accepted by the filter.
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

// getTransaction returns a function to get a transaction by its ID.
func (s *DBService) getTransaction(user *User, transactionID uint64) func(*badger.Txn) (*Transaction, error) {
	return func(txn *badger.Txn) (*Transaction, error) {
		transactionKey := user.createTransactionKeyFromID(transactionID)
		item, err := txn.Get(transactionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get transaction %v: %w", string(transactionKey), err)
		}

		k := item.Key()
		transaction := &Transaction{}
		if err := item.Value(transaction.decode); err != nil {
			return nil, fmt.Errorf("failed to read value of transaction %v: %w", string(k), err)
		}

		return transaction, nil
	}
}

// getTransaction returns a function to get transactions with the specified options.
func (s *DBService) getTransactions(user *User, options GetTransactionOptions) func(*badger.Txn) ([]*Transaction, error) {
	return func(txn *badger.Txn) ([]*Transaction, error) {
		transactions := make([]*Transaction, 0)

		opts := iteratorIndexOptions()
		opts.Prefix = []byte(user.createTransactionIndexKeyPrefix())
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()
		//TODO: remove this workaround for Badger and just use it.Rewind()
		reversePrefix := append([]byte(user.createTransactionIndexKeyPrefix()), 0xff)

		var currentItem uint64
		skipItem := func() bool {
			currentItem++
			return currentItem < (options.Offset + 1)
		}
		emptyFilter := options.TransactionFilterOptions.IsEmpty()
		for it.Seek(reversePrefix); it.Valid(); it.Next() {
			if emptyFilter && skipItem() {
				continue
			}
			transactionID, err := user.decodeTransactionIndexKey(it.Item().Key())
			if err != nil {
				return nil, fmt.Errorf("failed parse transaction index: %w", err)
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

// GetTransaction returns a Transaction by its ID.
// If the Transaction doesn't exist, it returns an error.
func (s *DBService) GetTransaction(user *User, transactionID uint64) (*Transaction, error) {
	var transaction *Transaction

	err := s.db.View(func(txn *badger.Txn) error {
		var err error
		transaction, err = s.getTransaction(user, transactionID)(txn)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction %v: %w", transactionID, err)
	}
	return transaction, nil
}

// GetTransactions returns transactions for user matching the filter and paging options.
// Returns an empty list if no transactions match the options.
func (s *DBService) GetTransactions(user *User, options GetTransactionOptions) ([]*Transaction, error) {
	var transactions []*Transaction

	err := s.db.View(func(txn *badger.Txn) error {
		var err error
		transactions, err = s.getTransactions(user, options)(txn)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	return transactions, nil
}

// CountTransactions returns the number of transactions matching the filter options.
func (s *DBService) CountTransactions(user *User, options TransactionFilterOptions) (uint64, error) {
	var count uint64

	emptyFilter := options.IsEmpty()
	err := s.db.View(func(txn *badger.Txn) error {
		opts := iteratorDoNotPrefetchOptions()
		opts.Prefix = []byte(user.createTransactionIndexKeyPrefix())
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			if !emptyFilter {
				transactionID, err := user.decodeTransactionIndexKey(it.Item().Key())
				if err != nil {
					return fmt.Errorf("failed to parse transaction index: %w", err)
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
		return 0, fmt.Errorf("failed to count transactions: %w", err)
	}
	return count, nil
}

// DeleteTransaction deletes a Transaction and its sort index key by its ID.
// Deleting a transaction also updates the affected Account balance.
// If transaction doesn't exist, returns an error.
func (s *DBService) DeleteTransaction(user *User, transactionID uint64) error {
	key := user.createTransactionKeyFromID(transactionID)
	return s.db.Update(func(txn *badger.Txn) error {
		deleteTransaction := &Transaction{}
		if err := getPreviousValue(txn, key, deleteTransaction.decode); err != nil {
			if err == badger.ErrKeyNotFound {
				return fmt.Errorf("cannot delete transaction %v if it doesn't exist", string(key))
			}
			return fmt.Errorf("failed to read value of deleted transaction %v: %w", string(key), err)
		}

		if err := s.updateAccountsBalance(user, &deleteTransaction.Components, nil)(txn); err != nil {
			return fmt.Errorf("cannot update accounts balance: %w", err)
		}

		transactionIndexKey := user.createTransactionIndexKey(deleteTransaction)
		if err := txn.Delete(transactionIndexKey); err != nil {
			return fmt.Errorf("failed to delete transaction index: %w", err)
		}

		return txn.Delete(key)
	})
}

package data

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
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
	UUID        string
	Description string
	Type        int
	Tags        []string
	Date        string
	Components  []TransactionComponent
}

// TransactionComponent contains details for a part of a Transaction.
type TransactionComponent struct {
	Amount      int64
	AccountUUID string
}

// TransactionFilterOptions specifies filter parameters for Transactions.
type TransactionFilterOptions struct {
	FilterDescription    string
	FilterFromDate       string
	FilterToDate         string
	FilterTags           []string
	FilterAccounts       []string
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

// createTransactionIndexKey creates an index key for transaction.
func (s *DBService) createTransactionIndexKey(user *User, transaction *Transaction) error {
	date, err := time.Parse(inputDateFormat, transaction.Date)
	if err != nil {
		return fmt.Errorf("cannot parse date %v: %w", transaction.Date, err)
	}

	// Add the year index key.
	indexKey := []byte(user.createTransactionKeyPrefix())
	yearKey := make([]byte, 2)
	binary.BigEndian.PutUint16(yearKey, uint16(date.Year()))
	if err := s.addReferencedKey(indexKey, yearKey, true); err != nil {
		return fmt.Errorf("cannot add add year index %v: %w", transaction.Date, err)
	}

	// Add the month index key.
	indexKey = append(indexKey, yearKey...)
	monthKey := []byte{uint8(date.Month())}
	if err := s.addReferencedKey(indexKey, monthKey, true); err != nil {
		return fmt.Errorf("cannot add add month index %v: %w", transaction.Date, err)
	}

	// Add the day index key.
	indexKey = append(indexKey, monthKey...)
	dayKey := []byte{uint8(date.Day())}
	if err := s.addReferencedKey(indexKey, dayKey, true); err != nil {
		return fmt.Errorf("cannot add add day index %v: %w", transaction.Date, err)
	}

	// Add transaction to the day index.
	// Transaction will be added as the last item for that day.
	indexKey = append(indexKey, dayKey...)
	if err := s.addReferencedKey(indexKey, []byte(transaction.UUID), false); err != nil {
		return fmt.Errorf("cannot add transaction% v to day index %v: %w", transaction.UUID, transaction.Date, err)
	}

	return nil
}

// deleteTransactionIndexKey deletes an index key for transaction.
func (s *DBService) deleteTransactionIndexKey(user *User, transaction *Transaction) error {
	date, err := time.Parse(inputDateFormat, transaction.Date)
	if err != nil {
		return fmt.Errorf("cannot parse date %v: %w", transaction.Date, err)
	}

	// Build the transaction index key.
	yearIndexKey := []byte(user.createTransactionKeyPrefix())
	yearKey := make([]byte, 2)
	binary.BigEndian.PutUint16(yearKey, uint16(date.Year()))

	monthIndexKey := append(yearIndexKey, yearKey...)
	monthKey := []byte{uint8(date.Month())}

	dayIndexKey := append(monthIndexKey, monthKey...)
	dayKey := []byte{uint8(date.Day())}
	indexKey := append(dayIndexKey, dayKey...)

	if err := s.deleteReferencedKey(indexKey, []byte(transaction.UUID)); err != nil {
		return err
	}

	// Cleanup empty parent keys.
	parentIndexKeys := [][]byte{indexKey, dayIndexKey, monthIndexKey, yearIndexKey}
	for i := range parentIndexKeys {
		parentIndexKey := parentIndexKeys[i]

		indexKeys, err := s.getReferencedKeys(parentIndexKey)
		if err != nil {
			return err
		}
		if len(indexKeys) > 0 {
			// Items still remaining in index.
			break
		}
		// No transactions remaining for this index - delete it.
		if err := s.db.Delete(parentIndexKey); err != nil {
			return err
		}
	}
	return nil
}

// createTransaction creates transaction for user.
// The transaction UUID is not generated here and should be generated before
// calling this method.
func (s *DBService) createTransaction(user *User, transaction *Transaction) error {
	key := user.createTransactionKey(transaction)
	value, err := transaction.encode()
	if err != nil {
		return fmt.Errorf("cannot encode transaction: %w", err)
	}

	if err := s.createTransactionIndexKey(user, transaction); err != nil {
		return fmt.Errorf("cannot create index for transaction: %w", err)
	}

	if err := s.updateAccountsBalance(user, nil, &transaction.Components); err != nil {
		return fmt.Errorf("cannot update account balance: %w", err)
	}

	return s.db.Put(key, value)
}

// CreateTransaction saves a new Transaction into the database.
func (s *DBService) CreateTransaction(user *User, transaction *Transaction) error {
	transaction.UUID = uuid.NewString()

	return s.update(func() error {
		return s.createTransaction(user, transaction)
	})
}

// UpdateTransaction updates an existing Transaction in the database.
func (s *DBService) UpdateTransaction(user *User, transaction *Transaction) error {
	return s.update(func() error {
		key := user.createTransactionKey(transaction)

		previousTransaction := &Transaction{}
		value, err := s.db.Get(key)
		if err != nil {
			return fmt.Errorf("cannot get previous value for transaction %v: %w", string(key), err)
		}
		if value == nil {
			return fmt.Errorf("cannot update transaction %v if it doesn't exist", string(key))
		}
		if err := previousTransaction.decode(value); err != nil {
			return fmt.Errorf("cannot decode previous value for transaction %v: %w", string(key), err)
		}
		if transaction == previousTransaction {
			log.WithField("key", string(key)).Debug("Transaction is unchanged")
			return nil
		}

		if err := s.updateAccountsBalance(user, &previousTransaction.Components, &transaction.Components); err != nil {
			return fmt.Errorf("cannot update account balance: %w", err)
		}

		if transaction.Date != previousTransaction.Date {
			if err := s.createTransactionIndexKey(user, transaction); err != nil {
				return fmt.Errorf("cannot create index for transaction %v: %w", string(key), err)
			}
			if err := s.deleteTransactionIndexKey(user, previousTransaction); err != nil {
				return fmt.Errorf("cannot delete previous index for transaction %v: %w", string(key), err)
			}
		}

		value, err = transaction.encode()
		if err != nil {
			return fmt.Errorf("cannot encode transaction: %w", err)
		}
		return s.db.Put(key, value)
	})
}

// updateAccountsBalance updates account balance for a transaction.
func (s *DBService) updateAccountsBalance(user *User, previousComponents *[]TransactionComponent, newComponents *[]TransactionComponent) error {
	accountDeltas := make(map[string]int64)
	if previousComponents != nil {
		for _, component := range *previousComponents {
			accountDeltas[component.AccountUUID] = accountDeltas[component.AccountUUID] - component.Amount
		}
	}
	if newComponents != nil {
		for _, component := range *newComponents {
			accountDeltas[component.AccountUUID] = accountDeltas[component.AccountUUID] + component.Amount
		}
	}
	for accountID, deltaAmount := range accountDeltas {
		if err := s.updateAccountBalance(user, accountID, deltaAmount); err != nil {
			return fmt.Errorf("cannot update account balance: %w", err)
		}
	}
	return nil
}

// IsEmpty returns if options do not apply any filtering (all transactions match this filter).
func (options *TransactionFilterOptions) IsEmpty() bool {
	return options.FilterDescription == "" &&
		options.FilterFromDate == "" &&
		options.FilterToDate == "" &&
		len(options.FilterTags) == 0 &&
		len(options.FilterAccounts) == 0 &&
		!options.ExcludeExpenseIncome &&
		!options.ExcludeTransfer
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
	containsAccount := func(searchIn *Transaction, searchFor []string) bool {
		for _, a := range searchIn.Components {
			for _, b := range searchFor {
				if a.AccountUUID == b {
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

// getTransaction gets a transaction by its UUID.
func (s *DBService) getTransaction(user *User, transactionUUID string) (*Transaction, error) {
	transactionKey := user.createTransactionKeyFromUUID(transactionUUID)
	value, err := s.db.Get(transactionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction %v: %w", string(transactionKey), err)
	}
	if value == nil {
		return nil, nil
	}

	transaction := &Transaction{}
	if err := transaction.decode(value); err != nil {
		return nil, fmt.Errorf("failed to read value of transaction %v: %w", string(transactionKey), err)
	}

	return transaction, nil
}

// iterateTransactions will iterate transactions, following their sort order.
// For each transaction, it will  call handleFn.
// If doneFn returns true (or handleFn returns an error), iteration will stop.
func (s *DBService) iterateTransactions(user *User,
	handleFn func(transactionUUID string) error,
	doneFn func() bool) error {
	indexKey := []byte(user.createTransactionKeyPrefix())
	years, err := s.getReferencedKeys(indexKey)
	if err != nil {
		return fmt.Errorf("failed get transactions years index: %w", err)
	}
	for i := len(years) - 1; i >= 0; i-- {
		year := years[i]
		monthsIndexKey := append(indexKey, year...)
		months, err := s.getReferencedKeys(monthsIndexKey)
		if err != nil {
			return fmt.Errorf("failed get transactions months index: %w", err)
		}

		for j := len(months) - 1; j >= 0; j-- {
			month := months[j]
			daysIndexKey := append(monthsIndexKey, month...)
			days, err := s.getReferencedKeys(daysIndexKey)
			if err != nil {
				return fmt.Errorf("failed get transactions days index: %w", err)
			}
			for k := len(days) - 1; k >= 0; k-- {
				day := days[k]
				transactionsIndexKey := append(daysIndexKey, day...)
				transactionKeys, err := s.getReferencedKeys(transactionsIndexKey)
				if err != nil {
					return fmt.Errorf("failed get transactions index: %w", err)
				}
				for l := len(transactionKeys) - 1; l >= 0; l-- {
					transactionUUID := transactionKeys[l]
					// TODO: if transaction date matches the index date, schedule a cleanup for this user.
					if err := handleFn(string(transactionUUID)); err != nil {
						return err
					}
					if doneFn() {
						return nil
					}
				}
			}
		}
	}
	return nil
}

// getTransaction gets transactions with the specified options.
func (s *DBService) getTransactions(user *User, options GetTransactionOptions) ([]*Transaction, error) {
	transactions := make([]*Transaction, 0)

	var currentItem uint64
	skipItem := func() bool {
		currentItem++
		return currentItem < (options.Offset + 1)
	}
	emptyFilter := options.TransactionFilterOptions.IsEmpty()

	handleFn := func(transactionUUID string) error {
		transaction, err := s.getTransaction(user, transactionUUID)
		if err != nil {
			return err
		}
		if transaction == nil {
			// TODO: schedule a cleanup for this user.
			return nil
		}

		if !emptyFilter {
			if !options.TransactionFilterOptions.Matches(transaction) {
				return nil
			}
		}
		if !skipItem() {
			transactions = append(transactions, transaction)
		}

		return nil
	}
	doneFn := func() bool { return uint64(len(transactions)) >= options.Limit }

	if err := s.iterateTransactions(user, handleFn, doneFn); err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetTransaction returns a Transaction by its UUID.
// If the Transaction doesn't exist, it returns nil.
func (s *DBService) GetTransaction(user *User, transactionUUID string) (*Transaction, error) {
	var transaction *Transaction

	err := s.view(func() error {
		var err error
		transaction, err = s.getTransaction(user, transactionUUID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction %v: %w", transactionUUID, err)
	}
	return transaction, nil
}

// GetTransactions returns transactions for user matching the filter and paging options.
// Returns an empty list if no transactions match the options.
func (s *DBService) GetTransactions(user *User, options GetTransactionOptions) ([]*Transaction, error) {
	var transactions []*Transaction

	err := s.view(func() error {
		var err error
		transactions, err = s.getTransactions(user, options)
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
	err := s.view(func() error {
		handleFn := func(transactionUUID string) error {
			if emptyFilter {
				transactionKey := user.createTransactionKeyFromUUID(transactionUUID)
				exists, err := s.db.Has(transactionKey)
				if err != nil {
					return err
				}
				if !exists {
					// TODO: schedule a cleanup for this user.
					return nil
				}
			} else {
				transaction, err := s.getTransaction(user, transactionUUID)
				if err != nil {
					return err
				}
				if transaction == nil {
					// TODO: schedule a cleanup for this user.
					return nil
				}
				if !options.Matches(transaction) {
					return nil
				}
			}

			count++
			return nil
		}
		doneFn := func() bool { return false }
		return s.iterateTransactions(user, handleFn, doneFn)
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count transactions: %w", err)
	}
	return count, nil
}

// deleteTransactions deletes all transactions for user.
func (s *DBService) deleteTransactions(user *User) error {
	transactions, err := s.getTransactions(user, GetAllTransactionsOptions)
	if err != nil {
		return fmt.Errorf("failed to get transactions to delete: %w", err)
	}

	for _, transaction := range transactions {
		key := user.createTransactionKeyFromUUID(transaction.UUID)
		if err := s.db.Delete(key); err != nil {
			return err
		}

		if err := s.deleteTransactionIndexKey(user, transaction); err != nil {
			return err
		}
	}
	return nil
}

// DeleteTransaction deletes a Transaction and its sort index key by its UUID.
// Deleting a transaction also updates the affected Account balance.
// If transaction doesn't exist, returns an error.
func (s *DBService) DeleteTransaction(user *User, transactionUUID string) error {
	key := user.createTransactionKeyFromUUID(transactionUUID)
	return s.update(func() error {
		value, err := s.db.Get(key)
		if err != nil {
			return fmt.Errorf("cannot get transaction to delete %v: %w", transactionUUID, err)
		} else if value == nil {
			return fmt.Errorf("cannot delete transaction %v because it doesn't exist: %w", transactionUUID, err)
		}

		deleteTransaction := &Transaction{}
		if err := deleteTransaction.decode(value); err != nil {
			return fmt.Errorf("cannot decode transaction %v to delete: %w", transactionUUID, err)
		}

		if err := s.updateAccountsBalance(user, &deleteTransaction.Components, nil); err != nil {
			return fmt.Errorf("cannot update accounts balance: %w", err)
		}

		if err := s.deleteTransactionIndexKey(user, deleteTransaction); err != nil {
			return fmt.Errorf("failed to delete transaction index: %w", err)
		}

		return s.db.Delete(key)
	})
}

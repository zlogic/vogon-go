package data

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v3"
)

// backupData is the toplevel structure exported in a backup.
type backupData struct {
	Accounts     []*Account
	Transactions []*Transaction
}

// Backup returns a serialized copy of all data for user.
func (s *DBService) Backup(user *User) (string, error) {
	data := backupData{}

	err := s.db.View(func(txn *badger.Txn) error {
		var err error
		accounts, err := s.getAccounts(user)(txn)
		if err != nil {
			return fmt.Errorf("failed to get accounts: %w", err)
		}

		transactions, err := s.getTransactions(user, GetAllTransactionsOptions)(txn)
		if err != nil {
			return fmt.Errorf("failed to get transactions: %w", err)
		}

		sortTransactionsAsc(transactions)
		for i, transaction := range transactions {
			transaction.ID = uint64(i)
		}

		data.Accounts = accounts
		data.Transactions = transactions
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to get data to back up: %w", err)
	}

	value, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling json: %w", err)
	}

	return string(value), nil
}

func deletePrefix(prefix []byte) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		opts := iteratorDoNotPrefetchOptions()
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			if err := txn.Delete(key); err != nil {
				return fmt.Errorf("error deleting key %v: %w", key, err)
			}
		}
		return nil
	}
}

// Restore replaces all data for user with the provided serialized backup.
func (s *DBService) Restore(user *User, value string) error {
	data := backupData{}
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return fmt.Errorf("error unmarshaling json: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		// Delete previous values
		if err := deletePrefix([]byte(user.createAccountKeyPrefix()))(txn); err != nil {
			return fmt.Errorf("failed to cleanup previous accounts: %w", err)
		}
		if err := deletePrefix([]byte(user.createTransactionKeyPrefix()))(txn); err != nil {
			return fmt.Errorf("failed to cleanup previous transactions: %w", err)
		}
		if err := deletePrefix([]byte(user.createTransactionIndexKeyPrefix()))(txn); err != nil {
			return fmt.Errorf("failed to cleanup previous transactions index: %w", err)
		}

		accountIDs := make(map[uint64]uint64)

		seq, err := s.db.GetSequence([]byte(user.createSequenceAccountKey()), 100)
		defer seq.Release()
		if err != nil {
			return fmt.Errorf("cannot create account sequence object: %w", err)
		}
		for _, account := range data.Accounts {
			id, err := seq.Next()
			if err != nil {
				return fmt.Errorf("cannot generate id for account: %w", err)
			}
			accountIDs[account.ID] = id

			account.ID = id
			account.Balance = 0

			if err := s.createAccount(user, account)(txn); err != nil {
				return fmt.Errorf("failed to create account %v: %w", account, err)
			}
		}

		seq, err = s.db.GetSequence([]byte(user.createSequenceTransactionKey()), 1000)
		defer seq.Release()
		if err != nil {
			return fmt.Errorf("cannot create transaction sequence object: %w", err)
		}
		for _, transaction := range data.Transactions {
			id, err := seq.Next()
			if err != nil {
				return fmt.Errorf("cannot generate id for transaction: %w", err)
			}
			transaction.ID = id

			transaction.normalize()

			for i, component := range transaction.Components {
				accountID, ok := accountIDs[component.AccountID]
				if !ok {
					return fmt.Errorf("cannot remap account id for component %v", component)
				}
				transaction.Components[i].AccountID = accountID
			}

			if err := s.createTransaction(user, transaction)(txn); err != nil {
				return fmt.Errorf("failed to create transaction %v: %w", transaction, err)
			}
		}
		return nil
	})
}

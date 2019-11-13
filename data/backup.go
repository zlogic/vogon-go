package data

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v2"
)

// BackupData is the toplevel structure exported in a backup.
type BackupData struct {
	Accounts     []*Account
	Transactions []*Transaction
}

// Backup returns a serialized copy of all data for user.
func (s *DBService) Backup(user *User) (string, error) {
	data := BackupData{}

	err := s.db.View(func(txn *badger.Txn) error {
		var err error
		accounts, err := s.getAccounts(user)(txn)
		if err != nil {
			return fmt.Errorf("Failed to get accounts when backing up data because of %w", err)
		}

		transactions, err := s.getTransactions(user, GetAllTransactionsOptions)(txn)
		if err != nil {
			return fmt.Errorf("Failed to get transactions when backing up data because of %w", err)
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
		return "", fmt.Errorf("Failed to get data to back up because of %w", err)
	}

	value, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("Error marshaling json (%w)", err)
	}

	return string(value), nil
}

func deletePrefix(prefix []byte) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		opts := IteratorDoNotPrefetchOptions()
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			if err := txn.Delete(key); err != nil {
				return fmt.Errorf("Error deleting key %v because of %w", key, err)
			}
		}
		return nil
	}
}

// Restore replaces all data for user with the provided serialized backup.
func (s *DBService) Restore(user *User, value string) error {
	data := BackupData{}
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return fmt.Errorf("Error unmarshaling json (%w)", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		// Delete previous values
		if err := deletePrefix([]byte(user.CreateAccountKeyPrefix()))(txn); err != nil {
			return fmt.Errorf("Failed to cleanup previous accounts because of %w", err)
		}
		if err := deletePrefix([]byte(user.CreateTransactionKeyPrefix()))(txn); err != nil {
			return fmt.Errorf("Failed to cleanup previous transactions because of %w", err)
		}
		if err := deletePrefix([]byte(user.CreateTransactionIndexKeyPrefix()))(txn); err != nil {
			return fmt.Errorf("Failed to cleanup previous transactions index because of %w", err)
		}

		accountIDs := make(map[uint64]uint64)

		seq, err := s.db.GetSequence([]byte(user.CreateSequenceAccountKey()), 100)
		defer seq.Release()
		if err != nil {
			return fmt.Errorf("Cannot create account sequence object because of %w", err)
		}
		for _, account := range data.Accounts {
			id, err := seq.Next()
			if err != nil {
				return fmt.Errorf("Cannot generate id for account because of %w", err)
			}
			accountIDs[account.ID] = id

			account.ID = id
			account.Balance = 0

			if err := s.createAccount(user, account)(txn); err != nil {
				return fmt.Errorf("Failed to create account %v because of %w", account, err)
			}
		}

		seq, err = s.db.GetSequence([]byte(user.CreateSequenceTransactionKey()), 1000)
		defer seq.Release()
		if err != nil {
			return fmt.Errorf("Cannot create transaction sequence object because of %w", err)
		}
		for _, transaction := range data.Transactions {
			id, err := seq.Next()
			if err != nil {
				return fmt.Errorf("Cannot generate id for transaction because of %w", err)
			}
			transaction.ID = id

			transaction.Normalize()

			for i, component := range transaction.Components {
				accountID, ok := accountIDs[component.AccountID]
				if !ok {
					return fmt.Errorf("Cannot remap account id for component %v", component)
				}
				transaction.Components[i].AccountID = accountID
			}

			if err := s.createTransaction(user, transaction)(txn); err != nil {
				return fmt.Errorf("Failed to create transaction %v because of %w", transaction, err)
			}
		}
		return nil
	})
}

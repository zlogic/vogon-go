package data

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
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
			return errors.Wrap(err, "Failed to get accounts when backing up data")
		}

		transactions, err := s.getTransactions(user, GetAllTransactionsOptions)(txn)
		if err != nil {
			return errors.Wrap(err, "Failed to get transactions when backing up data")
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
		return "", errors.Wrapf(err, "Failed to get data to back up")
	}

	value, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", errors.Wrap(err, "Error marshaling json")
	}

	return string(value), nil
}

func deletePrefix(prefix []byte) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		it := txn.NewIterator(IteratorDoNotPrefetchOptions())
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := make([]byte, len(item.Key()))
			copy(key, item.Key())
			if err := txn.Delete(key); err != nil {
				return errors.Wrapf(err, "Error deleting key %v", key)
			}
		}
		return nil
	}
}

// Restore replaces all data for user with the provided serialized backup.
func (s *DBService) Restore(user *User, value string) error {
	data := BackupData{}
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return errors.Wrap(err, "Error unmarshaling json")
	}

	return s.db.Update(func(txn *badger.Txn) error {
		// Delete previous values
		if err := deletePrefix([]byte(user.CreateAccountKeyPrefix()))(txn); err != nil {
			return errors.Wrap(err, "Failed to cleanup previous accounts")
		}
		if err := deletePrefix([]byte(user.CreateTransactionKeyPrefix()))(txn); err != nil {
			return errors.Wrap(err, "Failed to cleanup previous transactions")
		}
		if err := deletePrefix([]byte(user.CreateTransactionIndexKeyPrefix()))(txn); err != nil {
			return errors.Wrap(err, "Failed to cleanup previous transactions index")
		}

		accountIDs := make(map[uint64]uint64)

		seq, err := s.db.GetSequence([]byte(user.CreateSequenceAccountKey()), 100)
		defer seq.Release()
		if err != nil {
			return errors.Wrap(err, "Cannot create account sequence object")
		}
		for _, account := range data.Accounts {
			id, err := seq.Next()
			if err != nil {
				return errors.Wrap(err, "Cannot generate id for account")
			}
			accountIDs[account.ID] = id

			account.ID = id
			account.Balance = 0

			if err := s.createAccount(user, account)(txn); err != nil {
				return errors.Wrapf(err, "Failed to create account %v", account)
			}
		}

		seq, err = s.db.GetSequence([]byte(user.CreateSequenceTransactionKey()), 1000)
		defer seq.Release()
		if err != nil {
			return errors.Wrap(err, "Cannot create transaction sequence object")
		}
		for _, transaction := range data.Transactions {
			id, err := seq.Next()
			if err != nil {
				return errors.Wrap(err, "Cannot generate id for transaction")
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
				return errors.Wrapf(err, "Failed to create transaction %v", transaction)
			}
		}
		return nil
	})
}

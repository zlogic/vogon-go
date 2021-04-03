package data

import (
	"encoding/json"
	"fmt"
)

// backupData is the toplevel structure exported in a backup.
type backupData struct {
	Accounts     []*Account
	Transactions []*Transaction
}

// Backup returns a serialized copy of all data for user.
func (s *DBService) Backup(user *User) (string, error) {
	data := backupData{}

	err := s.view(func() error {
		var err error
		accounts, err := s.getAccounts(user)
		if err != nil {
			return fmt.Errorf("failed to get accounts: %w", err)
		}

		transactions, err := s.getTransactions(user, GetAllTransactionsOptions)
		if err != nil {
			return fmt.Errorf("failed to get transactions: %w", err)
		}

		// Reverse the order.
		for i, j := 0, len(transactions)-1; i < j; i, j = i+1, j-1 {
			transactions[i], transactions[j] = transactions[j], transactions[i]
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

// Restore replaces all data for user with the provided serialized backup.
func (s *DBService) Restore(user *User, value string) error {
	data := backupData{}
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return fmt.Errorf("error unmarshaling json: %w", err)
	}

	return s.update(func() error {
		// Delete previous values.
		if err := s.deleteAccounts(user); err != nil {
			return fmt.Errorf("failed to cleanup previous accounts: %w", err)
		}
		if err := s.deleteTransactions(user); err != nil {
			return fmt.Errorf("failed to cleanup previous transactions: %w", err)
		}
		for _, account := range data.Accounts {
			account.Balance = 0

			if err := s.createAccount(user, account); err != nil {
				return fmt.Errorf("failed to create account %v: %w", account, err)
			}
		}

		for _, transaction := range data.Transactions {
			transaction.normalize()

			if err := s.createTransaction(user, transaction); err != nil {
				return fmt.Errorf("failed to create transaction %v: %w", transaction, err)
			}
		}
		return nil
	})
}

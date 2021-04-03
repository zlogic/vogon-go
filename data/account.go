package data

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Account keeps the balance and other details for an account.
type Account struct {
	UUID           string
	Name           string
	Balance        int64
	Currency       string
	IncludeInTotal bool
	ShowInList     bool
}

// encode serializes an Account.
func (account *Account) encode() ([]byte, error) {
	var value bytes.Buffer
	if err := gob.NewEncoder(&value).Encode(account); err != nil {
		return nil, err
	}
	return value.Bytes(), nil
}

// decode deserializes an Account.
func (account *Account) decode(val []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(val)).Decode(account)
}

// createAccount creates and saves the specified account.
// The account UUID is not generated here and should be generated before
// calling this method.
func (s *DBService) createAccount(user *User, account *Account) error {
	key := user.createAccountKey(account)
	value, err := account.encode()

	if err != nil {
		return fmt.Errorf("cannot encode account: %w", err)
	}

	if err := s.addReferencedKey([]byte(user.createAccountKeyPrefix()), []byte(account.UUID), false); err != nil {
		return fmt.Errorf("cannot add account to index: %w", err)
	}

	return s.db.Put(key, value)
}

// CreateAccount creates and saves the specified account.
// It generates sets the ID to the generated account ID.
func (s *DBService) CreateAccount(user *User, account *Account) error {
	account.UUID = uuid.NewString()
	account.Balance = 0

	return s.update(func() error {
		return s.createAccount(user, account)
	})
}

// UpdateAccount saves an already existing account.
// If the account doesn't exist, it returns an error.
func (s *DBService) UpdateAccount(user *User, account *Account) error {
	return s.update(func() error {
		key := user.createAccountKey(account)

		previousAccount, err := s.getAccount(user, account.UUID)
		if err != nil {
			return fmt.Errorf("cannot get previous value for account %v: %w", string(key), err)
		}

		account.Balance = previousAccount.Balance
		if account == previousAccount {
			log.WithField("key", string(key)).Debug("Account is unchanged")
			return nil
		}

		value, err := account.encode()
		if err != nil {
			return fmt.Errorf("cannot encode account: %w", err)
		}
		return s.db.Put(key, value)
	})
}

// updateAccountBalance updates the account balance by delta.
func (s *DBService) updateAccountBalance(user *User, accountUUID string, deltaBalance int64) error {
	if deltaBalance == 0 {
		return nil
	}
	key := user.createAccountKeyFromUUID(accountUUID)

	account := &Account{}
	value, err := s.db.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get previous value for %v: %w", string(key), err)
	}
	if value == nil {
		return fmt.Errorf("cannot update account %v if it doesn't exist", string(key))
	}
	if err := account.decode(value); err != nil {
		return fmt.Errorf("cannot get previous value for account %v: %w", string(key), err)
	}

	account.Balance += deltaBalance

	value, err = account.encode()
	if err != nil {
		return fmt.Errorf("cannot encode account: %w", err)
	}
	return s.db.Put(key, value)
}

// getAccounts returns all accounts for user.
func (s *DBService) getAccounts(user *User) ([]*Account, error) {
	accountsPrefix := []byte(user.createAccountKeyPrefix())
	accountsUUIDs, err := s.getReferencedKeys(accountsPrefix)
	if err != nil {
		return nil, fmt.Errorf("cannot get accounts UUIDs for user: %w", err)
	}

	accounts := make([]*Account, 0, len(accountsUUIDs))
	for _, accountUUID := range accountsUUIDs {
		accountKey := user.createAccountKeyFromUUID(string(accountUUID))
		accountValue, err := s.db.Get(accountKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get account %v: %w", string(accountKey), err)
		}
		if accountValue == nil {
			// TODO: schedule a cleanup for this user.
			continue
		}

		account := &Account{}
		if err := account.decode(accountValue); err != nil {
			return nil, fmt.Errorf("failed to read value for account %v: %w", string(accountKey), err)
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// GetAccount returns an Account by its UUID.
// If the Account doesn't exist, it returns nil.
func (s *DBService) getAccount(user *User, accountUUID string) (*Account, error) {
	key := user.createAccountKeyFromUUID(accountUUID)

	value, err := s.db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get account %v: %w", string(key), err)
	}
	if value == nil {
		return nil, nil
	}

	account := &Account{}
	if err := account.decode(value); err != nil {
		return nil, fmt.Errorf("failed to read value for account %v: %w", string(key), err)
	}
	return account, nil
}

// GetAccount returns an Account by its UUID.
// If the Account doesn't exist, it returns nil.
func (s *DBService) GetAccount(user *User, accountUUID string) (*Account, error) {
	var account *Account
	err := s.view(func() error {
		var err error
		account, err = s.getAccount(user, accountUUID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

// GetAccounts returns all accounts for user.
func (s *DBService) GetAccounts(user *User) ([]*Account, error) {
	var accounts []*Account
	err := s.view(func() error {
		var err error
		accounts, err = s.getAccounts(user)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	return accounts, nil
}

// deleteAccounts deletes all accounts for user.
func (s *DBService) deleteAccounts(user *User) error {
	accountsPrefix := []byte(user.createAccountKeyPrefix())
	accountsUUIDs, err := s.getReferencedKeys(accountsPrefix)
	if err != nil {
		return fmt.Errorf("cannot get accounts UUIDs for user: %w", err)
	}

	for _, accountUUID := range accountsUUIDs {
		accountKey := user.createAccountKeyFromUUID(string(accountUUID))
		exists, err := s.db.Has(accountKey)
		if err != nil {
			return fmt.Errorf("failed to get account %v: %w", string(accountKey), err)
		}
		if !exists {
			continue
		}

		if err := s.db.Delete(accountKey); err != nil {
			return err
		}
	}

	return s.db.Delete(accountsPrefix)
}

// DeleteAccount deletes an account by its UUID.
// If the account doesn't exist, it returns an error.
func (s *DBService) DeleteAccount(user *User, accountUUID string) error {
	key := user.createAccountKeyFromUUID(accountUUID)
	return s.update(func() error {
		exists, err := s.db.Has(key)
		if err != nil {
			return fmt.Errorf("cannot check if account exists %v: %w", accountUUID, err)
		} else if !exists {
			return fmt.Errorf("cannot delete account %v because it doesn't exist: %w", accountUUID, err)
		}

		if err := s.db.Delete(key); err != nil {
			return fmt.Errorf("cannot delete account %v: %w", accountUUID, err)
		}

		accountsPrefix := []byte(user.createAccountKeyPrefix())
		return s.deleteReferencedKey(accountsPrefix, []byte(accountUUID))
	})
}

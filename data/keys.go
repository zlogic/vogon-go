package data

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// separatator is used to separate parts of an item key.
const separator = "/"

// encodePart encodes a part of the key into a string so that it can be safely joined with separator.
func encodePart(part string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(part))
}

// decodePart decodes a part of the key into a string.
func decodePart(part string) (string, error) {
	res, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

// userKeyPrefix is the key prefix for User entries.
const userKeyPrefix = "user" + separator

// createUserKey creates a key for user.
func createUserKey(username string) []byte {
	return []byte(userKeyPrefix + encodePart(username))
}

// createKey creates a key for user.
func (user *User) createKey() []byte {
	return createUserKey(user.username)
}

// decodeUserKey decodes the username from a user key.
func decodeUserKey(key []byte) (*string, error) {
	keyString := string(key)
	if !strings.HasPrefix(keyString, userKeyPrefix) {
		return nil, fmt.Errorf("not a user key: %v", keyString)
	}
	parts := strings.Split(keyString, separator)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format of user key: %v", keyString)
	}
	username, err := decodePart(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode username %v: %w", keyString, err)
	}
	return &username, nil
}

// accountKeyPrefix is the key prefix for Account.
const accountKeyPrefix = "account" + separator

// createAccountKeyPrefix creates an Account key prefix for user.
func (user *User) createAccountKeyPrefix() string {
	return accountKeyPrefix + user.UUID
}

// createAccountKeyFromUUID creates a key for an Account based on its UUID.
func (user *User) createAccountKeyFromUUID(accountUUID string) []byte {
	return []byte(user.createAccountKeyPrefix() + separator + accountUUID)
}

// createAccountKey creates a key for an Account entry.
func (user *User) createAccountKey(account *Account) []byte {
	return user.createAccountKeyFromUUID(account.UUID)
}

// transactionKeyPrefix is the key prefix for Transaction.
const transactionKeyPrefix = "transaction" + separator

// createTransactionKeyPrefix creates a Transaction key prefix for user.
func (user *User) createTransactionKeyPrefix() string {
	return transactionKeyPrefix + user.UUID
}

// createTransactionKeyFromUUID creates a key for a Transaction based on its UUID.
func (user *User) createTransactionKeyFromUUID(transactionUUID string) []byte {
	return []byte(user.createTransactionKeyPrefix() + separator + transactionUUID)
}

// createTransactionKey creates a key for a Transaction entry.
func (user *User) createTransactionKey(transaction *Transaction) []byte {
	return user.createTransactionKeyFromUUID(transaction.UUID)
}

// serverConfigKeyPrefix is the key prefix for a ServerConfig item.
const serverConfigKeyPrefix = "serverconfig" + separator

// createServerConfigKey creates a key for a ServerConfig item.
func createServerConfigKey(varName string) []byte {
	return []byte(serverConfigKeyPrefix + encodePart(varName))
}

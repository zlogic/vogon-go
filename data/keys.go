package data

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// separatator is used to separate parts of an item key.
const separator = "/"

// indexPrefix is the key prefix for indexes.
const indexPrefix = "index"

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
	return accountKeyPrefix + strconv.FormatUint(user.ID, 10) + separator
}

// createAccountKeyFromID creates a key for an Account based on its ID.
func (user *User) createAccountKeyFromID(accountID uint64) []byte {
	id := make([]byte, 8)
	binary.BigEndian.PutUint64(id, accountID)
	return append([]byte(user.createAccountKeyPrefix()), id...)
}

// createAccountKey creates a key for an Account entry.
func (user *User) createAccountKey(account *Account) []byte {
	return user.createAccountKeyFromID(account.ID)
}

// transactionKeyPrefix is the key prefix for Transaction.
const transactionKeyPrefix = "transaction" + separator

// createTransactionKeyPrefix creates a Transaction key prefix for user.
func (user *User) createTransactionKeyPrefix() string {
	return transactionKeyPrefix + strconv.FormatUint(user.ID, 10) + separator
}

// createTransactionKeyFromID creates a key for a Transaction based on its ID.
func (user *User) createTransactionKeyFromID(transactionID uint64) []byte {
	id := make([]byte, 8)
	binary.BigEndian.PutUint64(id, transactionID)
	return append([]byte(user.createTransactionKeyPrefix()), id...)
}

// createTransactionKey creates a key for a Transaction entry.
func (user *User) createTransactionKey(transaction *Transaction) []byte {
	return user.createTransactionKeyFromID(transaction.ID)
}

// transactionIndexPrefix is the key prefix for a Transaction sort index.
const transactionIndexPrefix = indexPrefix + separator + "transaction" + separator

// createTransactionIndexKeyPrefix creates a Transaction sort index key prefix for user.
func (user *User) createTransactionIndexKeyPrefix() string {
	return transactionIndexPrefix + strconv.FormatUint(user.ID, 10) + separator
}

// createTransactionIndexKey creates a Transaction sort index key a Transaciton.
func (user *User) createTransactionIndexKey(transaction *Transaction) []byte {
	id := make([]byte, 8)
	binary.BigEndian.PutUint64(id, transaction.ID)
	return append([]byte(user.createTransactionIndexKeyPrefix()+transaction.Date+separator), id...)
}

// decodeTransactionIndexKey decodes the Transaction ID from a transaction sort index key.
func (user *User) decodeTransactionIndexKey(key []byte) (uint64, error) {
	currentSeparator := 0
	start := 0
	for i := 0; i < len(key); i++ {
		c := key[i]
		if byte(separator[0]) == c {
			currentSeparator++
		}
		if currentSeparator == 4 {
			start = i + 1
			break
		}
	}
	if start == 0 || len(key) != start+8 {
		return 0, fmt.Errorf("invalid format of transaction index key: %v", string(key))
	}

	return binary.BigEndian.Uint64(key[len(key)-8:]), nil
}

// serverConfigKeyPrefix is the key prefix for a ServerConfig item.
const serverConfigKeyPrefix = "serverconfig" + separator

// createServerConfigKey creates a key for a ServerConfig item.
func createServerConfigKey(varName string) []byte {
	return []byte(serverConfigKeyPrefix + encodePart(varName))
}

// decodeServerConfigKey decodes the name of a ServerConfig key.
func decodeServerConfigKey(key []byte) (string, error) {
	keyString := string(key)
	if !strings.HasPrefix(keyString, serverConfigKeyPrefix) {
		return "", fmt.Errorf("not a config item key: %v", keyString)
	}
	parts := strings.Split(keyString, separator)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid format of config item key: %v", keyString)
	}
	value, err := decodePart(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to config item value %v: %w", keyString, err)
	}
	return value, nil
}

// sequencePrefix is the key prefix for Sequence items.
const sequencePrefix = "sequence" + separator

// sequenceUserKey is the key prefix for the User Sequence.
const sequenceUserKey = sequencePrefix + "user"

// sequenceAccountPrefix is the key prefix for the Account Sequence.
const sequenceAccountPrefix = sequencePrefix + "account" + separator

// createSequenceAccountKey creates a key for the Account Sequence.
func (user *User) createSequenceAccountKey() []byte {
	return []byte(sequenceAccountPrefix + strconv.FormatUint(user.ID, 10))
}

// sequenceTransactionPrefix is the key prefix for the Trasaction Sequence.
const sequenceTransactionPrefix = sequencePrefix + "transaction" + separator

// createSequenceTransactionKey creates a key for the Transaction Sequence.
func (user *User) createSequenceTransactionKey() []byte {
	return []byte(sequenceTransactionPrefix + strconv.FormatUint(user.ID, 10))
}

package data

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

const separator = "/"
const indexPrefix = "index"

func encodePart(part string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(part))
}

func decodePart(part string) (string, error) {
	res, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

// UserKeyPrefix is the key prefix for User entries.
const UserKeyPrefix = "user" + separator

// CreateUserKey creates a key for user.
func CreateUserKey(username string) []byte {
	return []byte(UserKeyPrefix + encodePart(username))
}

// CreateKey creates a key for user.
func (user *User) CreateKey() []byte {
	return CreateUserKey(user.username)
}

// DecodeUserKey decodes the username from a user key.
func DecodeUserKey(key []byte) (*string, error) {
	keyString := string(key)
	if !strings.HasPrefix(keyString, UserKeyPrefix) {
		return nil, fmt.Errorf("Not a user key: %v", keyString)
	}
	parts := strings.Split(keyString, separator)
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid format of user key: %v", keyString)
	}
	username, err := decodePart(parts[1])
	if err != nil {
		return nil, fmt.Errorf("Failed to decode username: %v because of %w", keyString, err)
	}
	return &username, nil
}

// AccountKeyPrefix is the key prefix for Account.
const AccountKeyPrefix = "account" + separator

// CreateAccountKeyPrefix creates an Account key prefix for user.
func (user *User) CreateAccountKeyPrefix() string {
	return AccountKeyPrefix + strconv.FormatUint(user.ID, 10) + separator
}

// CreateAccountKeyFromID creates a key for an Account based on its ID.
func (user *User) CreateAccountKeyFromID(accountID uint64) []byte {
	id := make([]byte, 8)
	binary.BigEndian.PutUint64(id, accountID)
	return append([]byte(user.CreateAccountKeyPrefix()), id...)
}

// CreateAccountKey creates a key for an Account entry.
func (user *User) CreateAccountKey(account *Account) []byte {
	return user.CreateAccountKeyFromID(account.ID)
}

// TransactionKeyPrefix is the key prefix for Transaction.
const TransactionKeyPrefix = "transaction" + separator

// CreateTransactionKeyPrefix creates a Transaction key prefix for user.
func (user *User) CreateTransactionKeyPrefix() string {
	return TransactionKeyPrefix + strconv.FormatUint(user.ID, 10) + separator
}

// CreateTransactionKeyFromID creates a key for a Transaction based on its ID.
func (user *User) CreateTransactionKeyFromID(transactionID uint64) []byte {
	id := make([]byte, 8)
	binary.BigEndian.PutUint64(id, transactionID)
	return append([]byte(user.CreateTransactionKeyPrefix()), id...)
}

// CreateTransactionKey creates a key for a Transaction entry.
func (user *User) CreateTransactionKey(transaction *Transaction) []byte {
	return user.CreateTransactionKeyFromID(transaction.ID)
}

// TransactionIndexPrefix is the key prefix for a Transaction sort index.
const TransactionIndexPrefix = indexPrefix + separator + "transaction" + separator

// CreateTransactionIndexKeyPrefix creates a Transaction sort index key prefix for user.
func (user *User) CreateTransactionIndexKeyPrefix() string {
	return TransactionIndexPrefix + strconv.FormatUint(user.ID, 10) + separator
}

// CreateTransactionIndexKey creates a Transaction sort index key a Transaciton.
func (user *User) CreateTransactionIndexKey(transaction *Transaction) []byte {
	id := make([]byte, 8)
	binary.BigEndian.PutUint64(id, transaction.ID)
	return append([]byte(user.CreateTransactionIndexKeyPrefix()+transaction.Date+separator), id...)
}

// DecodeTransactionIndexKey decodes the Transaction ID from a transaction sort index key.
func (user *User) DecodeTransactionIndexKey(key []byte) (uint64, error) {
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
		return 0, fmt.Errorf("Invalid format of transaction index key: %v", string(key))
	}

	return binary.BigEndian.Uint64(key[len(key)-8:]), nil
}

// ServerConfigKeyPrefix is the key prefix for a ServerConfig item.
const ServerConfigKeyPrefix = "serverconfig" + separator

// CreateServerConfigKey creates a key for a ServerConfig item.
func CreateServerConfigKey(varName string) []byte {
	return []byte(ServerConfigKeyPrefix + encodePart(varName))
}

// DecodeServerConfigKey decodes the name of a ServerConfig key.
func DecodeServerConfigKey(key []byte) (string, error) {
	keyString := string(key)
	if !strings.HasPrefix(keyString, ServerConfigKeyPrefix) {
		return "", fmt.Errorf("Not a config item key: %v", keyString)
	}
	parts := strings.Split(keyString, separator)
	if len(parts) != 2 {
		return "", fmt.Errorf("Invalid format of config item key: %v", keyString)
	}
	value, err := decodePart(parts[1])
	if err != nil {
		return "", fmt.Errorf("Failed to config item value: %v because of %w", keyString, err)
	}
	return value, nil
}

// SequencePrefix is the key prefix for Sequence items.
const SequencePrefix = "sequence" + separator

// SequenceUserKey is the key prefix for the User Sequence.
const SequenceUserKey = SequencePrefix + "user"

// SequenceAccountPrefix is the key prefix for the Account Sequence.
const SequenceAccountPrefix = SequencePrefix + "account" + separator

// CreateSequenceAccountKey creates a key for the Account Sequence.
func (user *User) CreateSequenceAccountKey() []byte {
	return []byte(SequenceAccountPrefix + strconv.FormatUint(user.ID, 10))
}

// SequenceTransactionPrefix is the key prefix for the Trasaction Sequence.
const SequenceTransactionPrefix = SequencePrefix + "transaction" + separator

// CreateSequenceTransactionKey creates a key for the Transaction Sequence.
func (user *User) CreateSequenceTransactionKey() []byte {
	return []byte(SequenceTransactionPrefix + strconv.FormatUint(user.ID, 10))
}

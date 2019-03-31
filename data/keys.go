package data

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
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

const UserKeyPrefix = "user" + separator

func (user *User) CreateKey() []byte {
	return []byte(UserKeyPrefix + user.username)
}

func DecodeUserKey(key []byte) (*string, error) {
	keyString := string(key)
	if !strings.HasPrefix(keyString, UserKeyPrefix) {
		return nil, errors.Errorf("Not a user key: %v", keyString)
	}
	parts := strings.Split(keyString, separator)
	if len(parts) != 2 {
		return nil, errors.Errorf("Invalid format of user key: %v", keyString)
	}
	return &parts[1], nil
}

const AccountKeyPrefix = "account" + separator

func (user *User) CreateAccountKeyPrefix() string {
	return AccountKeyPrefix + strconv.FormatUint(user.ID, 10) + separator
}

func (user *User) CreateAccountKeyFromID(accountID uint64) []byte {
	id := make([]byte, 8)
	binary.BigEndian.PutUint64(id, accountID)
	return append([]byte(user.CreateAccountKeyPrefix()), id...)
}

func (user *User) CreateAccountKey(account *Account) []byte {
	return user.CreateAccountKeyFromID(account.ID)
}

const TransactionKeyPrefix = "transaction" + separator

func (user *User) CreateTransactionKeyPrefix() string {
	return TransactionKeyPrefix + strconv.FormatUint(user.ID, 10) + separator
}

func (user *User) CreateTransactionKeyFromID(transactionID uint64) []byte {
	id := make([]byte, 8)
	binary.BigEndian.PutUint64(id, transactionID)
	return append([]byte(user.CreateTransactionKeyPrefix()), id...)
}

func (user *User) CreateTransactionKey(transaction *Transaction) []byte {
	return user.CreateTransactionKeyFromID(transaction.ID)
}

const TransactionIndexPrefix = indexPrefix + separator + "transaction" + separator

func (user *User) CreateTransactionIndexKeyPrefix() string {
	return TransactionIndexPrefix + strconv.FormatUint(user.ID, 10) + separator
}

func (user *User) CreateTransactionIndexKey(transaction *Transaction) []byte {
	id := make([]byte, 8)
	binary.BigEndian.PutUint64(id, transaction.ID)
	return append([]byte(user.CreateTransactionIndexKeyPrefix()+transaction.Date+separator), id...)
}

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

const ServerConfigKeyPrefix = "serverconfig" + separator

func CreateServerConfigKey(varName string) []byte {
	return []byte(ServerConfigKeyPrefix + encodePart(varName))
}

func DecodeServerConfigKey(key []byte) (string, error) {
	keyString := string(key)
	if !strings.HasPrefix(keyString, ServerConfigKeyPrefix) {
		return "", errors.Errorf("Not a config item key: %v", keyString)
	}
	parts := strings.Split(keyString, separator)
	if len(parts) != 2 {
		return "", errors.Errorf("Invalid format of config item key: %v", keyString)
	}
	value, err := decodePart(parts[1])
	if err != nil {
		return "", errors.Errorf("Failed to config item value: %v because of %v", keyString, err)
	}
	return value, nil
}

const SequencePrefix = "sequence" + separator
const SequenceUserKey = SequencePrefix + "user"

const SequenceAccountPrefix = SequencePrefix + "account" + separator

func (user *User) CreateSequenceAccountKey() []byte {
	return []byte(SequenceAccountPrefix + strconv.FormatUint(user.ID, 10))
}

const SequenceTransactionPrefix = SequencePrefix + "transaction" + separator

func (user *User) CreateSequenceTransactionKey() []byte {
	return []byte(SequenceTransactionPrefix + strconv.FormatUint(user.ID, 10))
}

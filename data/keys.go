package data

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const separator = "/"

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
	return []byte(user.CreateAccountKeyPrefix() + strconv.FormatUint(accountID, 10))
}

func (user *User) CreateAccountKey(account *Account) []byte {
	return user.CreateAccountKeyFromID(account.ID)
}

const TransactionKeyPrefix = "transaction" + separator

func (user *User) CreateTransactionKeyPrefix() string {
	return TransactionKeyPrefix + strconv.FormatUint(user.ID, 10) + separator
}

func (user *User) CreateTransactionKey(transaction *Transaction) []byte {
	return []byte(user.CreateTransactionKeyPrefix() + strconv.FormatUint(transaction.ID, 10))
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
		return "", errors.Errorf("Failed to config item valye: %v because of %v", keyString, err)
	}
	return value, nil
}

const SequencePrefix = "sequence" + separator
const SequenceUserKey = SequencePrefix + "user"

const SequenceAccountPrefix = SequencePrefix + "account" + separator

func CreateSequenceAccountKey(user *User) []byte {
	return []byte(SequenceAccountPrefix + strconv.FormatUint(user.ID, 10))
}

const SequenceTransactionPrefix = SequencePrefix + "transaction" + separator

func CreateSequenceTransactionKey(user *User) []byte {
	return []byte(SequenceTransactionPrefix + strconv.FormatUint(user.ID, 10))
}

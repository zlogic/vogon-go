package server

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path"

	"github.com/dgraph-io/badger/v2"
	"github.com/gorilla/securecookie"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/zlogic/vogon-go/data"
)

var testUser = data.User{ID: 11}

var testExistingUsers = make(map[string]data.User)

type DBMock struct {
	mock.Mock
}

func (m *DBMock) GetOrCreateConfigVariable(varName string, generator func() (string, error)) (string, error) {
	args := m.Called(varName, generator)
	return args.Get(0).(string), args.Error(1)
}

func (m *DBMock) GetUser(username string) (*data.User, error) {
	args := m.Called(username)
	user := args.Get(0)
	var returnUser *data.User
	if user != nil {
		returnUser = user.(*data.User)
	}
	return returnUser, args.Error(1)
}

func (m *DBMock) SaveUser(user *data.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *DBMock) GetAccounts(user *data.User) ([]*data.Account, error) {
	args := m.Called(user)
	accounts := args.Get(0)
	var returnAccounts []*data.Account
	if accounts != nil {
		returnAccounts = accounts.([]*data.Account)
	}
	return returnAccounts, args.Error(1)
}

func (m *DBMock) CountTransactions(user *data.User, options data.TransactionFilterOptions) (uint64, error) {
	args := m.Called(user, options)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *DBMock) GetTransactions(user *data.User, options data.GetTransactionOptions) ([]*data.Transaction, error) {
	args := m.Called(user, options)
	transactions := args.Get(0)
	var returnTransactions []*data.Transaction
	if transactions != nil {
		returnTransactions = transactions.([]*data.Transaction)
	}
	return returnTransactions, args.Error(1)
}

func (m *DBMock) CreateAccount(user *data.User, account *data.Account) error {
	args := m.Called(user, account)
	return args.Error(0)
}

func (m *DBMock) UpdateAccount(user *data.User, account *data.Account) error {
	args := m.Called(user, account)
	return args.Error(0)
}

func (m *DBMock) GetAccount(user *data.User, accountID uint64) (*data.Account, error) {
	args := m.Called(user, accountID)
	account := args.Get(0)
	var returnAccount *data.Account
	if account != nil {
		returnAccount = account.(*data.Account)
	}
	return returnAccount, args.Error(1)
}

func (m *DBMock) DeleteAccount(user *data.User, accountID uint64) error {
	args := m.Called(user, accountID)
	return args.Error(0)
}

func (m *DBMock) CreateTransaction(user *data.User, transaction *data.Transaction) error {
	args := m.Called(user, transaction)
	return args.Error(0)
}

func (m *DBMock) UpdateTransaction(user *data.User, transaction *data.Transaction) error {
	args := m.Called(user, transaction)
	return args.Error(0)
}

func (m *DBMock) GetTransaction(user *data.User, transactionID uint64) (*data.Transaction, error) {
	args := m.Called(user, transactionID)
	transaction := args.Get(0)
	var returnTransaction *data.Transaction
	if transaction != nil {
		returnTransaction = transaction.(*data.Transaction)
	}
	return returnTransaction, args.Error(1)
}

func (m *DBMock) DeleteTransaction(user *data.User, transactionID uint64) error {
	args := m.Called(user, transactionID)
	return args.Error(0)
}

func (m *DBMock) GetTags(user *data.User) ([]string, error) {
	args := m.Called(user)
	return args.Get(0).([]string), args.Error(1)
}

func (m *DBMock) Backup(user *data.User) (string, error) {
	args := m.Called(user)
	return args.Get(0).(string), args.Error(1)
}

func (m *DBMock) Restore(user *data.User, value string) error {
	args := m.Called(user, value)
	return args.Error(0)
}

func createTestCookieHandler() (*CookieHandler, error) {
	hashKey := base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(64))
	blockKey := base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
	dbMock := new(DBMock)

	dbMock.On("GetOrCreateConfigVariable", "cookie-hash-key", mock.AnythingOfType("func() (string, error)")).Return(hashKey, nil).Once()
	dbMock.On("GetOrCreateConfigVariable", "cookie-block-key", mock.AnythingOfType("func() (string, error)")).Return(blockKey, nil).Once()
	return NewCookieHandler(dbMock)
}

func prepareTempDir() (string, func(), error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}
	recover := func() {
		os.Chdir(currentDir)
	}
	tempDir, err := ioutil.TempDir("", "vogon")
	if err != nil {
		return currentDir, recover, err
	}
	recover = func() {
		os.Chdir(currentDir)
		os.RemoveAll(tempDir)
	}
	err = os.Chdir(tempDir)
	return tempDir, recover, err
}

func prepareTestFile(dir, fileName string, data []byte) error {
	err := os.Mkdir(dir, 0755)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(dir, fileName), data, 0644)
}

func prepareExistingUser(username string) *data.User {
	existingUser, ok := testExistingUsers[username]
	if ok {
		user := existingUser
		return &user
	}

	tempDir, recover, err := prepareTempDir()
	defer func() {
		if recover != nil {
			recover()
		}
	}()

	var opts = badger.DefaultOptions(tempDir)
	opts.Logger = log.New()
	opts.ValueLogFileSize = 1 << 20
	opts.SyncWrites = false
	opts.CompactL0OnClose = false

	dbService, err := data.Open(opts)
	if err != nil {
		return nil
	}

	user := data.NewUser(username)
	err = dbService.SaveUser(user)
	dbService.Close()
	if err != nil {
		return nil
	}
	testExistingUsers[username] = *user
	return user
}

package server

import (
	"context"
	"net/http"

	"github.com/akrylysov/pogreb"
	"github.com/akrylysov/pogreb/fs"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"

	"github.com/zlogic/vogon-go/data"
	"github.com/zlogic/vogon-go/server/auth"
)

var testUser = data.User{UUID: "uuid11"}

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

func (m *DBMock) GetAccount(user *data.User, accountUUID string) (*data.Account, error) {
	args := m.Called(user, accountUUID)
	account := args.Get(0)
	var returnAccount *data.Account
	if account != nil {
		returnAccount = account.(*data.Account)
	}
	return returnAccount, args.Error(1)
}

func (m *DBMock) DeleteAccount(user *data.User, accountUUID string) error {
	args := m.Called(user, accountUUID)
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

func (m *DBMock) GetTransaction(user *data.User, transactionUUID string) (*data.Transaction, error) {
	args := m.Called(user, transactionUUID)
	transaction := args.Get(0)
	var returnTransaction *data.Transaction
	if transaction != nil {
		returnTransaction = transaction.(*data.Transaction)
	}
	return returnTransaction, args.Error(1)
}

func (m *DBMock) DeleteTransaction(user *data.User, transactionUUID string) error {
	args := m.Called(user, transactionUUID)
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

var testAuthCookie = "testusername"

type AuthHandlerMock struct {
	mock.Mock
	authUser *data.User
}

func (m *AuthHandlerMock) SetCookieUsername(w http.ResponseWriter, username string, rememberMe bool) error {
	args := m.Called(w, username, rememberMe)
	return args.Error(0)
}

func (m *AuthHandlerMock) AuthHandlerFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if m.authUser != nil {
			ctx = context.WithValue(ctx, auth.UserContextKey, m.authUser)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthHandlerMock) HasAuthenticationCookie(r *http.Request) bool {
	args := m.Called(r)
	return args.Get(0).(bool)
}

func (m *AuthHandlerMock) AllowUser(user *data.User) *http.Cookie {
	m.authUser = user
	return nil
}

func prepareExistingUser(username string) *data.User {
	existingUser, ok := testExistingUsers[username]
	if ok {
		user := existingUser
		return &user
	}

	logger := log.New()
	logger.SetLevel(log.FatalLevel)

	opts := pogreb.Options{FileSystem: fs.Mem}

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

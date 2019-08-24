package server

import (
	"github.com/zlogic/vogon-go/data"
)

// DB provides functions to read and write items in the database.
type DB interface {
	GetOrCreateConfigVariable(varName string, generator func() (string, error)) (string, error)

	GetUser(username string) (data.User, error)
	SaveUser(*data.User) error

	GetAccounts(data.User) ([]data.Account, error)
	GetTransactions(data.User, data.GetTransactionOptions) ([]data.Transaction, error)
	CountTransactions(data.User, data.TransactionFilterOptions) (uint64, error)
	CreateAccount(data.User, *data.Account) error
	UpdateAccount(data.User, data.Account) error
	GetAccount(user data.User, accountID uint64) (data.Account, error)
	DeleteAccount(user data.User, transactionID uint64) error
	CreateTransaction(data.User, *data.Transaction) error
	UpdateTransaction(data.User, data.Transaction) error
	GetTransaction(user data.User, transactionID uint64) (data.Transaction, error)
	DeleteTransaction(user data.User, transactionID uint64) error

	GetTags(user data.User) ([]string, error)

	Backup(user data.User) (string, error)
	Restore(user data.User, value string) error
}

// Services keeps references to all services needed by handlers.
type Services struct {
	db            DB
	cookieHandler *CookieHandler
}

// CreateServices creates a Services instance with db and default implementations of other services.
func CreateServices(db data.DBService) (Services, error) {
	cookieHandler, err := NewCookieHandler(db)
	if err != nil {
		return Services{}, err
	}
	return Services{
		db:            db,
		cookieHandler: cookieHandler,
	}, nil
}

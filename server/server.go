package server

import (
	"github.com/zlogic/vogon-go/data"
)

type DB interface {
	GetOrCreateConfigVariable(varName string, generator func() (string, error)) (string, error)
	CreateUser(username string) (*data.User, error)

	GetUser(username string) (*data.User, error)
	SaveUser(*data.User) error
	SaveNewUser(*data.User) error
	SetUsername(user *data.User, newUsername string) error

	GetAccounts(*data.User) ([]*data.Account, error)
	GetTransactions(*data.User) ([]*data.Transaction, error)
	CreateAccount(*data.User, *data.Account) error
	UpdateAccount(*data.User, *data.Account) error
	GetAccount(user *data.User, accountID uint64) (*data.Account, error)
	DeleteAccount(user *data.User, transactionID uint64) error
	CreateTransaction(*data.User, *data.Transaction) error
	UpdateTransaction(*data.User, *data.Transaction) error
	GetTransaction(user *data.User, transactionID uint64) (*data.Transaction, error)
	DeleteTransaction(user *data.User, transactionID uint64) error

	Backup(user *data.User) (string, error)
	Restore(user *data.User, value string) error
}

type Services struct {
	db            DB
	cookieHandler *CookieHandler
}

func CreateServices(db *data.DBService) (*Services, error) {
	cookieHandler, err := NewCookieHandler(db)
	if err != nil {
		return nil, err
	}
	return &Services{
		db:            db,
		cookieHandler: cookieHandler,
	}, nil
}

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

	GetAccounts(user *data.User) ([]*data.Account, error)
	GetTransactions(user *data.User) ([]*data.Transaction, error)
	CreateTransaction(user *data.User, transaction *data.Transaction) error
	UpdateTransaction(user *data.User, transaction *data.Transaction) error
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

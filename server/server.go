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

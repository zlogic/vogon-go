package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var user = User{ID: 11}

func TestCreateAccount(t *testing.T) {
	dbService, cleanup, err := createDb()
	assert.NoError(t, err)
	defer cleanup()

	account1 := Account{
		Name:           "a1",
		Currency:       "USD",
		IncludeInTotal: false,
		ShowInList:     true,
	}

	saveAccount := account1
	err = dbService.CreateAccount(&user, &saveAccount)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), saveAccount.ID)

	accounts, err := dbService.GetAccounts(&user)
	assert.NoError(t, err)
	assert.Equal(t, []*Account{&account1}, accounts)

	account2 := Account{
		Name:           "a2",
		Currency:       "EUR",
		IncludeInTotal: true,
		ShowInList:     false,
	}

	saveAccount = account2
	err = dbService.CreateAccount(&user, &saveAccount)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), saveAccount.ID)

	account2.ID = 1
	accounts, err = dbService.GetAccounts(&user)
	assert.NoError(t, err)
	assert.Equal(t, []*Account{&account1, &account2}, accounts)
}

func TestUpdateAccount(t *testing.T) {
	dbService, cleanup, err := createDb()
	assert.NoError(t, err)
	defer cleanup()

	account1 := Account{
		Name:           "a1",
		Currency:       "USD",
		IncludeInTotal: false,
		ShowInList:     true,
	}
	account2 := Account{
		Name:           "a2",
		Currency:       "EUR",
		IncludeInTotal: true,
		ShowInList:     false,
	}

	saveAccount := account1
	err = dbService.CreateAccount(&user, &saveAccount)
	assert.NoError(t, err)

	saveAccount = account2
	err = dbService.CreateAccount(&user, &saveAccount)
	assert.NoError(t, err)

	account2.Name = "a2-"
	account2.Currency = "a2-"
	account2.IncludeInTotal = false
	account2.ShowInList = true
	account2.ID = 1
	saveAccount = account2
	err = dbService.UpdateAccount(&user, &saveAccount)
	assert.NoError(t, err)

	accounts, err := dbService.GetAccounts(&user)
	assert.NoError(t, err)
	assert.Equal(t, []*Account{&account1, &account2}, accounts)
}

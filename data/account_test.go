package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAccount(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	account1 := Account{
		Name:           "a1",
		Currency:       "USD",
		IncludeInTotal: false,
		ShowInList:     true,
	}

	saveAccount := account1
	err = dbService.CreateAccount(&testUser, &saveAccount)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), saveAccount.ID)

	accounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Account{&account1}, accounts)

	account2 := Account{
		Name:           "a2",
		Currency:       "EUR",
		IncludeInTotal: true,
		ShowInList:     false,
	}

	saveAccount = account2
	err = dbService.CreateAccount(&testUser, &saveAccount)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), saveAccount.ID)

	account2.ID = 1
	accounts, err = dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Account{&account1, &account2}, accounts)
}

func TestGetAccount(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

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

	err = dbService.CreateAccount(&testUser, &account1)
	assert.NoError(t, err)
	err = dbService.CreateAccount(&testUser, &account2)
	assert.NoError(t, err)

	account, err := dbService.GetAccount(&testUser, 0)
	assert.NoError(t, err)
	assert.Equal(t, &account1, account)

	account, err = dbService.GetAccount(&testUser, 1)
	assert.NoError(t, err)
	assert.Equal(t, &account2, account)
}

func TestGetAccountDoesNotExist(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	account, err := dbService.GetAccount(&testUser, 0)
	assert.Error(t, err)
	assert.Nil(t, account)
}

func TestUpdateAccount(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

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
	err = dbService.CreateAccount(&testUser, &saveAccount)
	assert.NoError(t, err)

	saveAccount = account2
	err = dbService.CreateAccount(&testUser, &saveAccount)
	assert.NoError(t, err)

	account2.Name = "a2-"
	account2.Currency = "a2-"
	account2.IncludeInTotal = false
	account2.ShowInList = true
	account2.ID = 1
	saveAccount = account2
	err = dbService.UpdateAccount(&testUser, &saveAccount)
	assert.NoError(t, err)

	accounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Account{&account1, &account2}, accounts)
}

func TestDeleteAccount(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

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
	err = dbService.CreateAccount(&testUser, &saveAccount)
	assert.NoError(t, err)

	saveAccount = account2
	err = dbService.CreateAccount(&testUser, &saveAccount)
	assert.NoError(t, err)

	err = dbService.DeleteAccount(&testUser, uint64(1))
	assert.NoError(t, err)

	accounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Account{&account1}, accounts)

	err = dbService.DeleteAccount(&testUser, uint64(0))
	assert.NoError(t, err)

	accounts, err = dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Empty(t, accounts)
}

func TestDeleteNonExistingAccount(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	account := Account{
		Name:           "a1",
		Currency:       "USD",
		IncludeInTotal: false,
		ShowInList:     true,
	}

	saveAccount := account
	err = dbService.CreateAccount(&testUser, &saveAccount)
	assert.NoError(t, err)

	err = dbService.DeleteAccount(&testUser, uint64(1))
	assert.Error(t, err)

	accounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Account{&account}, accounts)
}

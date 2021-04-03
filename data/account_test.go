package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAccount(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	assertIndexEquals(t, testUser.createAccountKeyPrefix())

	account1 := Account{
		Name:           "a1",
		Currency:       "USD",
		IncludeInTotal: false,
		ShowInList:     true,
	}

	saveAccount := account1
	err = dbService.CreateAccount(&testUser, &saveAccount)
	account1.UUID = saveAccount.UUID
	assert.NoError(t, err)
	assert.NotEmpty(t, saveAccount.UUID)

	accounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Account{&account1}, accounts)

	account2 := Account{
		Name:           "a2",
		Currency:       "EUR",
		IncludeInTotal: true,
		ShowInList:     false,
	}

	assertIndexEquals(t, testUser.createAccountKeyPrefix(), account1.UUID)

	saveAccount = account2
	err = dbService.CreateAccount(&testUser, &saveAccount)
	account2.UUID = saveAccount.UUID
	assert.NoError(t, err)
	assert.NotEmpty(t, saveAccount.UUID)
	assert.NotEqual(t, account1.UUID, saveAccount.UUID)

	accounts, err = dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Account{&account1, &account2}, accounts)

	assertIndexEquals(t, testUser.createAccountKeyPrefix(), account1.UUID, account2.UUID)
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

	account, err := dbService.GetAccount(&testUser, account1.UUID)
	assert.NoError(t, err)
	assert.Equal(t, &account1, account)

	account, err = dbService.GetAccount(&testUser, account2.UUID)
	assert.NoError(t, err)
	assert.Equal(t, &account2, account)
}

func TestGetAccountDoesNotExist(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	account, err := dbService.GetAccount(&testUser, "non-existing")
	assert.NoError(t, err)
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
	account1.UUID = saveAccount.UUID
	assert.NoError(t, err)

	saveAccount = account2
	err = dbService.CreateAccount(&testUser, &saveAccount)
	account2.UUID = saveAccount.UUID
	assert.NoError(t, err)

	account2.Name = "a2-"
	account2.Currency = "a2-"
	account2.IncludeInTotal = false
	account2.ShowInList = true

	saveAccount = account2
	err = dbService.UpdateAccount(&testUser, &saveAccount)
	assert.NoError(t, err)

	accounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Account{&account1, &account2}, accounts)

	assertIndexEquals(t, testUser.createAccountKeyPrefix(), account1.UUID, account2.UUID)
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
	account1.UUID = saveAccount.UUID

	saveAccount = account2
	err = dbService.CreateAccount(&testUser, &saveAccount)
	assert.NoError(t, err)
	account2.UUID = saveAccount.UUID

	assertIndexEquals(t, testUser.createAccountKeyPrefix(), account1.UUID, account2.UUID)

	err = dbService.DeleteAccount(&testUser, account2.UUID)
	assert.NoError(t, err)

	assertIndexEquals(t, testUser.createAccountKeyPrefix(), account1.UUID)

	accounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Account{&account1}, accounts)

	err = dbService.DeleteAccount(&testUser, account1.UUID)
	assert.NoError(t, err)

	accounts, err = dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Empty(t, accounts)

	assertIndexEquals(t, testUser.createAccountKeyPrefix())
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
	account.UUID = saveAccount.UUID
	assert.NoError(t, err)

	err = dbService.DeleteAccount(&testUser, "non-existing")
	assert.Error(t, err)

	accounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Account{&account}, accounts)

	assertIndexEquals(t, testUser.createAccountKeyPrefix(), account.UUID)
}

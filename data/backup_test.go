package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const restoreData = `{
  "Accounts": [
    {
      "ID": 0,
      "Name": "Orange Bank",
      "Balance": 99000,
      "Currency": "PLN",
      "IncludeInTotal": true,
      "ShowInList": true
    },
    {
      "ID": 1,
      "Name": "Green Bank",
      "Balance": 90000,
      "Currency": "ALL",
      "IncludeInTotal": true,
      "ShowInList": false
    },
    {
      "ID": 2,
      "Name": "Purple Bank",
      "Balance": 80000,
      "Currency": "ZWL",
      "IncludeInTotal": true,
      "ShowInList": false
    },
    {
      "ID": 3,
      "Name": "Magical Credit Card",
      "Balance": -8000,
      "Currency": "PLN",
      "IncludeInTotal": false,
      "ShowInList": false
    }
  ],
  "Transactions": [
    {
      "ID": 0,
      "Description": "Widgets",
      "Type": 0,
      "Tags": [
        "Widgets"
      ],
      "Date": "2015-11-02",
      "Components": [
        {
          "Amount": -10000,
          "AccountID": 1
        }
      ]
    },
    {
      "ID": 1,
      "Description": "Salary",
      "Type": 0,
      "Tags": [
        "Salary"
      ],
      "Date": "2015-11-01",
      "Components": [
        {
          "Amount": 100000,
          "AccountID": 0
        },
        {
          "Amount": 100000,
          "AccountID": 1
        },
        {
          "Amount": 100000,
          "AccountID": 2
        }
      ]
    },
    {
      "ID": 2,
      "Description": "Gadgets",
      "Type": 0,
      "Tags": [
        "Gadgets"
      ],
      "Date": "2015-11-03",
      "Components": [
        {
          "Amount": -10000,
          "AccountID": 3
        }
      ]
    },
    {
      "ID": 3,
      "Description": "Credit card payment",
      "Type": 1,
      "Tags": [
        "Credit"
      ],
      "Date": "2015-11-09",
      "Components": [
        {
          "Amount": -10000,
          "AccountID": 2
        },
        {
          "Amount": 2000,
          "AccountID": 3
        }
      ]
    },
    {
      "ID": 4,
      "Description": "Stuff",
      "Type": 1,
      "Tags": [
        "Gadgets",
        "Widgets"
      ],
      "Date": "2015-11-07",
      "Components": [
        {
          "Amount": -1000,
          "AccountID": 0
        },
        {
          "Amount": -10000,
          "AccountID": 2
        }
      ]
    }
  ]
}`

const backupData = `{
  "Accounts": [
    {
      "ID": 0,
      "Name": "Orange Bank",
      "Balance": 99000,
      "Currency": "PLN",
      "IncludeInTotal": true,
      "ShowInList": true
    },
    {
      "ID": 1,
      "Name": "Green Bank",
      "Balance": 90000,
      "Currency": "ALL",
      "IncludeInTotal": true,
      "ShowInList": false
    },
    {
      "ID": 2,
      "Name": "Purple Bank",
      "Balance": 80000,
      "Currency": "ZWL",
      "IncludeInTotal": true,
      "ShowInList": false
    },
    {
      "ID": 3,
      "Name": "Magical Credit Card",
      "Balance": -8000,
      "Currency": "PLN",
      "IncludeInTotal": false,
      "ShowInList": false
    }
  ],
  "Transactions": [
    {
      "ID": 1,
      "Description": "Salary",
      "Type": 0,
      "Tags": [
        "Salary"
      ],
      "Date": "2015-11-01",
      "Components": [
        {
          "Amount": 100000,
          "AccountID": 0
        },
        {
          "Amount": 100000,
          "AccountID": 1
        },
        {
          "Amount": 100000,
          "AccountID": 2
        }
      ]
    },
    {
      "ID": 0,
      "Description": "Widgets",
      "Type": 0,
      "Tags": [
        "Widgets"
      ],
      "Date": "2015-11-02",
      "Components": [
        {
          "Amount": -10000,
          "AccountID": 1
        }
      ]
    },
    {
      "ID": 2,
      "Description": "Gadgets",
      "Type": 0,
      "Tags": [
        "Gadgets"
      ],
      "Date": "2015-11-03",
      "Components": [
        {
          "Amount": -10000,
          "AccountID": 3
        }
      ]
    },
    {
      "ID": 4,
      "Description": "Stuff",
      "Type": 1,
      "Tags": [
        "Gadgets",
        "Widgets"
      ],
      "Date": "2015-11-07",
      "Components": [
        {
          "Amount": -1000,
          "AccountID": 0
        },
        {
          "Amount": -10000,
          "AccountID": 2
        }
      ]
    },
    {
      "ID": 3,
      "Description": "Credit card payment",
      "Type": 1,
      "Tags": [
        "Credit"
      ],
      "Date": "2015-11-09",
      "Components": [
        {
          "Amount": -10000,
          "AccountID": 2
        },
        {
          "Amount": 2000,
          "AccountID": 3
        }
      ]
    }
  ]
}`

func createBackupAccounts() []*Account {
	return []*Account{&Account{
		Name:           "Orange Bank",
		Currency:       "PLN",
		IncludeInTotal: true,
		ShowInList:     true,
	}, &Account{
		Name:           "Green Bank",
		Currency:       "ALL",
		IncludeInTotal: true,
		ShowInList:     false,
	}, &Account{
		Name:           "Purple Bank",
		Currency:       "ZWL",
		IncludeInTotal: true,
		ShowInList:     false,
	}, &Account{
		Name:           "Magical Credit Card",
		Currency:       "PLN",
		IncludeInTotal: false,
		ShowInList:     false,
	}}
}

func createBackupTransactions(accounts []*Account) []*Transaction {
	return []*Transaction{&Transaction{
		Description: "Widgets",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"Widgets"},
		Date:        "2015-11-02",
		Components: []TransactionComponent{
			TransactionComponent{AccountID: accounts[1].ID, Amount: -10000},
		},
	}, &Transaction{
		Description: "Salary",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"Salary"},
		Date:        "2015-11-01",
		Components: []TransactionComponent{
			TransactionComponent{AccountID: accounts[0].ID, Amount: 100000},
			TransactionComponent{AccountID: accounts[1].ID, Amount: 100000},
			TransactionComponent{AccountID: accounts[2].ID, Amount: 100000},
		},
	}, &Transaction{
		Description: "Gadgets",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"Gadgets"},
		Date:        "2015-11-03",
		Components: []TransactionComponent{
			TransactionComponent{AccountID: accounts[3].ID, Amount: -10000},
		},
	}, &Transaction{
		Description: "Credit card payment",
		Type:        TransactionTypeTransfer,
		Tags:        []string{"Credit"},
		Date:        "2015-11-09",
		Components: []TransactionComponent{
			TransactionComponent{AccountID: accounts[2].ID, Amount: -10000},
			TransactionComponent{AccountID: accounts[3].ID, Amount: 2000},
		},
	}, &Transaction{
		Description: "Stuff",
		Type:        TransactionTypeTransfer,
		Tags:        []string{"Gadgets", "Widgets"},
		Date:        "2015-11-07",
		Components: []TransactionComponent{
			TransactionComponent{AccountID: accounts[0].ID, Amount: -1000},
			TransactionComponent{AccountID: accounts[2].ID, Amount: -10000},
		},
	}}
}

func TestBackup(t *testing.T) {
	dbService, cleanup, err := createDb()
	assert.NoError(t, err)
	defer cleanup()

	accounts := createBackupAccounts()
	for _, account := range accounts {
		dbService.CreateAccount(&testUser, account)
	}

	transactions := createBackupTransactions(accounts)
	transactions[4].Tags = []string{"Widgets", "Gadgets"}
	for _, transaction := range transactions {
		assert.NoError(t, transaction.Normalize())
		dbService.CreateTransaction(&testUser, transaction)
	}

	json, err := dbService.Backup(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, backupData, json)
}

func TestRestore(t *testing.T) {
	dbService, cleanup, err := createDb()
	assert.NoError(t, err)
	defer cleanup()

	err = dbService.Restore(&testUser, restoreData)
	assert.NoError(t, err)

	expectedAccounts := createBackupAccounts()
	for i, account := range expectedAccounts {
		account.ID = uint64(i)
	}
	expectedAccounts[0].Balance = 99000
	expectedAccounts[1].Balance = 90000
	expectedAccounts[2].Balance = 80000
	expectedAccounts[3].Balance = -8000

	expectedTransactions := createBackupTransactions(expectedAccounts)
	for i, transaction := range expectedTransactions {
		transaction.ID = uint64(i)
	}
	sortTransactionsAsc(expectedTransactions)

	dbAccounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, expectedAccounts, dbAccounts)

	dbTransactions, err := dbService.GetTransactions(&testUser, GetAllTransactionsOptions)
	assert.NoError(t, err)
	sortTransactionsAsc(dbTransactions)
	assert.Equal(t, expectedTransactions, dbTransactions)
}

func TestRestoreOverwriteExistingData(t *testing.T) {
	dbService, cleanup, err := createDb()
	assert.NoError(t, err)
	defer cleanup()

	accounts := createBackupAccounts()
	accounts = append(accounts, &Account{Name: "Extra account", Currency: "PLN", IncludeInTotal: true, ShowInList: true})
	for _, account := range accounts {
		dbService.CreateAccount(&testUser, account)
	}

	transactions := createBackupTransactions(accounts)
	transactions[4].Tags = []string{"Widgets", "Gadgets"}
	transactions = append(transactions, &Transaction{
		Description: "Extra transaction",
		Type:        TransactionTypeTransfer, Tags: []string{"Extra"}, Date: "2019-03-23",
		Components: []TransactionComponent{
			TransactionComponent{AccountID: accounts[0].ID, Amount: -8800},
			TransactionComponent{AccountID: accounts[4].ID, Amount: -42000},
		},
	})
	for _, transaction := range transactions {
		assert.NoError(t, transaction.Normalize())
		dbService.CreateTransaction(&testUser, transaction)
	}

	err = dbService.Restore(&testUser, restoreData)
	assert.NoError(t, err)

	expectedAccounts := createBackupAccounts()
	for i, account := range expectedAccounts {
		account.ID = uint64(len(accounts)) + uint64(i)
	}
	expectedAccounts[0].Balance = 99000
	expectedAccounts[1].Balance = 90000
	expectedAccounts[2].Balance = 80000
	expectedAccounts[3].Balance = -8000

	expectedTransactions := createBackupTransactions(expectedAccounts)
	for i, transaction := range expectedTransactions {
		transaction.ID = uint64(len(transactions)) + uint64(i)
	}
	sortTransactionsAsc(expectedTransactions)

	dbAccounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, expectedAccounts, dbAccounts)

	dbTransactions, err := dbService.GetTransactions(&testUser, GetAllTransactionsOptions)
	assert.NoError(t, err)
	sortTransactionsAsc(dbTransactions)
	assert.Equal(t, expectedTransactions, dbTransactions)
}

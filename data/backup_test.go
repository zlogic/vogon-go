package data

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testRestoreData = `{
  "Accounts": [
    {
      "UUID": "uuid1",
      "Name": "Orange Bank",
      "Balance": 99000,
      "Currency": "PLN",
      "IncludeInTotal": true,
      "ShowInList": true
    },
    {
      "UUID": "uuid2",
      "Name": "Green Bank",
      "Balance": 90000,
      "Currency": "ALL",
      "IncludeInTotal": true,
      "ShowInList": false
    },
    {
      "UUID": "uuid3",
      "Name": "Purple Bank",
      "Balance": 80000,
      "Currency": "ZWL",
      "IncludeInTotal": true,
      "ShowInList": false
    },
    {
      "UUID": "uuid4",
      "Name": "Magical Credit Card",
      "Balance": -8000,
      "Currency": "PLN",
      "IncludeInTotal": false,
      "ShowInList": false
    }
  ],
  "Transactions": [
    {
      "UUID": "uuid2",
      "Description": "Salary",
      "Type": 0,
      "Tags": [
        "Salary"
      ],
      "Date": "2015-11-01",
      "Components": [
        {
          "Amount": 100000,
          "AccountUUID": "uuid1"
        },
        {
          "Amount": 100000,
          "AccountUUID": "uuid2"
        },
        {
          "Amount": 100000,
          "AccountUUID": "uuid3"
        }
      ]
    },
    {
      "UUID": "uuid1",
      "Description": "Widgets",
      "Type": 0,
      "Tags": [
        "Widgets"
      ],
      "Date": "2015-11-02",
      "Components": [
        {
          "Amount": -10000,
          "AccountUUID": "uuid2"
        }
      ]
    },
    {
      "UUID": "uuid3",
      "Description": "Gadgets",
      "Type": 0,
      "Tags": [
        "Gadgets"
      ],
      "Date": "2015-11-03",
      "Components": [
        {
          "Amount": -10000,
          "AccountUUID": "uuid4"
        }
      ]
    },
    {
      "UUID": "uuid5",
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
          "AccountUUID": "uuid1"
        },
        {
          "Amount": -10000,
          "AccountUUID": "uuid3"
        }
      ]
    },
    {
      "UUID": "uuid4",
      "Description": "Credit card payment",
      "Type": 1,
      "Tags": [
        "Credit"
      ],
      "Date": "2015-11-09",
      "Components": [
        {
          "Amount": -10000,
          "AccountUUID": "uuid3"
        },
        {
          "Amount": 2000,
          "AccountUUID": "uuid4"
        }
      ]
    }
  ]
}`

const testBackupData = `{
  "Accounts": [
    {
      "UUID": "uuid1",
      "Name": "Orange Bank",
      "Balance": 99000,
      "Currency": "PLN",
      "IncludeInTotal": true,
      "ShowInList": true
    },
    {
      "UUID": "uuid2",
      "Name": "Green Bank",
      "Balance": 90000,
      "Currency": "ALL",
      "IncludeInTotal": true,
      "ShowInList": false
    },
    {
      "UUID": "uuid3",
      "Name": "Purple Bank",
      "Balance": 80000,
      "Currency": "ZWL",
      "IncludeInTotal": true,
      "ShowInList": false
    },
    {
      "UUID": "uuid4",
      "Name": "Magical Credit Card",
      "Balance": -8000,
      "Currency": "PLN",
      "IncludeInTotal": false,
      "ShowInList": false
    }
  ],
  "Transactions": [
    {
      "UUID": "uuid2",
      "Description": "Salary",
      "Type": 0,
      "Tags": [
        "Salary"
      ],
      "Date": "2015-11-01",
      "Components": [
        {
          "Amount": 100000,
          "AccountUUID": "uuid1"
        },
        {
          "Amount": 100000,
          "AccountUUID": "uuid2"
        },
        {
          "Amount": 100000,
          "AccountUUID": "uuid3"
        }
      ]
    },
    {
      "UUID": "uuid1",
      "Description": "Widgets",
      "Type": 0,
      "Tags": [
        "Widgets"
      ],
      "Date": "2015-11-02",
      "Components": [
        {
          "Amount": -10000,
          "AccountUUID": "uuid2"
        }
      ]
    },
    {
      "UUID": "uuid3",
      "Description": "Gadgets",
      "Type": 0,
      "Tags": [
        "Gadgets"
      ],
      "Date": "2015-11-03",
      "Components": [
        {
          "Amount": -10000,
          "AccountUUID": "uuid4"
        }
      ]
    },
    {
      "UUID": "uuid5",
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
          "AccountUUID": "uuid1"
        },
        {
          "Amount": -10000,
          "AccountUUID": "uuid3"
        }
      ]
    },
    {
      "UUID": "uuid4",
      "Description": "Credit card payment",
      "Type": 1,
      "Tags": [
        "Credit"
      ],
      "Date": "2015-11-09",
      "Components": [
        {
          "Amount": -10000,
          "AccountUUID": "uuid3"
        },
        {
          "Amount": 2000,
          "AccountUUID": "uuid4"
        }
      ]
    }
  ]
}`

func createBackupAccounts() []*Account {
	return []*Account{{
		UUID:           "uuid1",
		Name:           "Orange Bank",
		Currency:       "PLN",
		IncludeInTotal: true,
		ShowInList:     true,
	}, {
		UUID:           "uuid2",
		Name:           "Green Bank",
		Currency:       "ALL",
		IncludeInTotal: true,
		ShowInList:     false,
	}, {
		UUID:           "uuid3",
		Name:           "Purple Bank",
		Currency:       "ZWL",
		IncludeInTotal: true,
		ShowInList:     false,
	}, {
		UUID:           "uuid4",
		Name:           "Magical Credit Card",
		Currency:       "PLN",
		IncludeInTotal: false,
		ShowInList:     false,
	}}
}

func createBackupTransactions(accounts []*Account) []*Transaction {
	return []*Transaction{{
		UUID:        "uuid1",
		Description: "Widgets",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"Widgets"},
		Date:        "2015-11-02",
		Components: []TransactionComponent{
			{AccountUUID: accounts[1].UUID, Amount: -10000},
		},
	}, {
		UUID:        "uuid2",
		Description: "Salary",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"Salary"},
		Date:        "2015-11-01",
		Components: []TransactionComponent{
			{AccountUUID: accounts[0].UUID, Amount: 100000},
			{AccountUUID: accounts[1].UUID, Amount: 100000},
			{AccountUUID: accounts[2].UUID, Amount: 100000},
		},
	}, {
		UUID:        "uuid3",
		Description: "Gadgets",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"Gadgets"},
		Date:        "2015-11-03",
		Components: []TransactionComponent{
			{AccountUUID: accounts[3].UUID, Amount: -10000},
		},
	}, {
		UUID:        "uuid4",
		Description: "Credit card payment",
		Type:        TransactionTypeTransfer,
		Tags:        []string{"Credit"},
		Date:        "2015-11-09",
		Components: []TransactionComponent{
			{AccountUUID: accounts[2].UUID, Amount: -10000},
			{AccountUUID: accounts[3].UUID, Amount: 2000},
		},
	}, {
		UUID:        "uuid5",
		Description: "Stuff",
		Type:        TransactionTypeTransfer,
		Tags:        []string{"Gadgets", "Widgets"},
		Date:        "2015-11-07",
		Components: []TransactionComponent{
			{AccountUUID: accounts[0].UUID, Amount: -1000},
			{AccountUUID: accounts[2].UUID, Amount: -10000},
		},
	}}
}

func TestBackup(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	accounts := createBackupAccounts()
	for _, account := range accounts {
		dbService.createAccount(&testUser, account)
	}

	transactions := createBackupTransactions(accounts)
	transactions[4].Tags = []string{"Widgets", "Gadgets"}
	for _, transaction := range transactions {
		assert.NoError(t, transaction.normalize())
		dbService.createTransaction(&testUser, transaction)
	}

	json, err := dbService.Backup(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, testBackupData, json)
}

func TestRestore(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	err = dbService.Restore(&testUser, testRestoreData)
	assert.NoError(t, err)

	expectedAccounts := createBackupAccounts()
	expectedAccounts[0].Balance = 99000
	expectedAccounts[1].Balance = 90000
	expectedAccounts[2].Balance = 80000
	expectedAccounts[3].Balance = -8000

	expectedTransactions := createBackupTransactions(expectedAccounts)
	sort.Slice(expectedTransactions, func(i, j int) bool {
		return strings.Compare(expectedTransactions[i].Date, expectedTransactions[j].Date) > 0
	})

	dbAccounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, expectedAccounts, dbAccounts)

	dbTransactions, err := dbService.GetTransactions(&testUser, GetAllTransactionsOptions)
	assert.NoError(t, err)
	assert.Equal(t, expectedTransactions, dbTransactions)
}

func TestRestoreOverwriteExistingData(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

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
			{AccountUUID: accounts[0].UUID, Amount: -8800},
			{AccountUUID: accounts[4].UUID, Amount: -42000},
		},
	})
	for _, transaction := range transactions {
		assert.NoError(t, transaction.normalize())
		dbService.CreateTransaction(&testUser, transaction)
	}

	err = dbService.Restore(&testUser, testRestoreData)
	assert.NoError(t, err)

	expectedAccounts := createBackupAccounts()
	expectedAccounts[0].Balance = 99000
	expectedAccounts[1].Balance = 90000
	expectedAccounts[2].Balance = 80000
	expectedAccounts[3].Balance = -8000

	expectedTransactions := createBackupTransactions(expectedAccounts)
	sort.Slice(expectedTransactions, func(i, j int) bool {
		return strings.Compare(expectedTransactions[i].Date, expectedTransactions[j].Date) > 0
	})

	dbAccounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, expectedAccounts, dbAccounts)

	dbTransactions, err := dbService.GetTransactions(&testUser, GetAllTransactionsOptions)
	assert.NoError(t, err)
	assert.Equal(t, expectedTransactions, dbTransactions)
}

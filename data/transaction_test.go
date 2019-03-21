package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeTransaction(t *testing.T) {
	typicalTransaction := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"t1", "t2"},
	}
	err := typicalTransaction.Normalize()
	assert.NoError(t, err)
	assert.Equal(t, Transaction{
		Description: "t1",
		Date:        "2019-03-20",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"t1", "t2"},
	}, typicalTransaction)

	unsortedTagsUnformattedDateTransaction := Transaction{Date: "2019-3-2", Tags: []string{"t2", "t1", "ta"}}
	err = unsortedTagsUnformattedDateTransaction.Normalize()
	assert.NoError(t, err)
	assert.Equal(t, Transaction{Date: "2019-03-02", Tags: []string{"t1", "t2", "ta"}}, unsortedTagsUnformattedDateTransaction)

	noTagsTransaction := Transaction{Date: "2019-03-02"}
	err = noTagsTransaction.Normalize()
	assert.NoError(t, err)
	assert.Equal(t, Transaction{Date: "2019-03-02"}, noTagsTransaction)

	duplicateTagsTransaction := Transaction{Date: "2019-03-02", Tags: []string{"t2", "t1", "t1"}}
	err = duplicateTagsTransaction.Normalize()
	assert.NoError(t, err)
	assert.Equal(t, Transaction{Date: "2019-03-02", Tags: []string{"t1", "t2"}}, duplicateTagsTransaction)

	emptyTagTransaction := Transaction{Date: "2019-03-02", Tags: []string{"t2", "t1", ""}}
	err = emptyTagTransaction.Normalize()
	assert.NoError(t, err)
	assert.Equal(t, Transaction{Date: "2019-03-02", Tags: []string{"t1", "t2"}}, emptyTagTransaction)

	noDateTransaction := Transaction{Tags: []string{"t2", "t1", "ta"}}
	err = noDateTransaction.Normalize()
	assert.Error(t, err)
	assert.Equal(t, Transaction{Tags: []string{"t2", "t1", "ta"}}, noDateTransaction)

	badDateTransaction := Transaction{Date: "helloworld", Tags: []string{"t2", "t1", "ta"}}
	err = badDateTransaction.Normalize()
	assert.Error(t, err)
	assert.Equal(t, Transaction{Date: "helloworld", Tags: []string{"t2", "t1", "ta"}}, badDateTransaction)
}

func TestCreateTransactionNoComponents(t *testing.T) {
	dbService, cleanup, err := createDb()
	assert.NoError(t, err)
	defer cleanup()

	transaction1 := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"t1", "t2"},
	}

	saveTransaction := transaction1
	err = dbService.CreateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), saveTransaction.ID)

	transactions, err := dbService.GetTransactions(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Transaction{&transaction1}, transactions)

	transaction2 := Transaction{
		Description: "t2",
		Date:        "2019-03-21",
		Type:        TransactionTypeTransfer,
		Tags:        []string{"t1", "t3"},
	}

	saveTransaction = transaction2
	err = dbService.CreateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), saveTransaction.ID)

	transaction2.ID = 1
	transactions, err = dbService.GetTransactions(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Transaction{&transaction1, &transaction2}, transactions)
}

func TestUpdateTransactionNoComponents(t *testing.T) {
	dbService, cleanup, err := createDb()
	assert.NoError(t, err)
	defer cleanup()

	transaction1 := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"t1", "t2"},
	}
	transaction2 := Transaction{
		Description: "t2",
		Date:        "2019-03-21",
		Type:        TransactionTypeTransfer,
		Tags:        []string{"t1", "t3"},
	}

	saveTransaction := transaction1
	err = dbService.CreateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)

	saveTransaction = transaction2
	err = dbService.CreateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)

	transaction2.Date = "2019-03-22"
	transaction2.Description = "t2-"
	transaction2.Tags = []string{"t1", "t3", "t4"}
	transaction2.Type = TransactionTypeTransfer
	transaction2.ID = 1
	saveTransaction = transaction2
	err = dbService.UpdateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)

	transactions, err := dbService.GetTransactions(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Transaction{&transaction1, &transaction2}, transactions)
}

func TestDeleteTransactionNoComponents(t *testing.T) {
	dbService, cleanup, err := createDb()
	assert.NoError(t, err)
	defer cleanup()

	transaction1 := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"t1", "t2"},
	}
	transaction2 := Transaction{
		Description: "t2",
		Date:        "2019-03-21",
		Type:        TransactionTypeTransfer,
		Tags:        []string{"t1", "t3"},
	}

	saveTransaction := transaction1
	err = dbService.CreateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)

	saveTransaction = transaction2
	err = dbService.CreateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)

	transaction2.ID = 1

	err = dbService.DeleteTransaction(&testUser, &transaction2)
	assert.NoError(t, err)

	transactions, err := dbService.GetTransactions(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Transaction{&transaction1}, transactions)

	err = dbService.DeleteTransaction(&testUser, &transaction1)
	assert.NoError(t, err)

	transactions, err = dbService.GetTransactions(&testUser)
	assert.NoError(t, err)
	assert.Empty(t, transactions)
}

func TestDeleteNonExistingTransaction(t *testing.T) {
	dbService, cleanup, err := createDb()
	assert.NoError(t, err)
	defer cleanup()

	transaction := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"t1", "t2"},
	}

	saveTransaction := transaction
	err = dbService.CreateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)

	deleteTransaction := transaction
	deleteTransaction.ID = 1

	err = dbService.DeleteTransaction(&testUser, &deleteTransaction)
	assert.Error(t, err)

	transactions, err := dbService.GetTransactions(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Transaction{&transaction}, transactions)
}

func TestCreateTransactionWithComponents(t *testing.T) {
	dbService, cleanup, err := createDb()
	assert.NoError(t, err)
	defer cleanup()

	err = createTestAccounts(dbService)
	assert.NoError(t, err)

	transaction1 := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"t1", "t2"},
		Components: []TransactionComponent{
			TransactionComponent{AccountID: 0, Amount: -1},
			TransactionComponent{AccountID: 1, Amount: 2},
		},
	}

	saveTransaction := transaction1
	err = dbService.CreateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), saveTransaction.ID)

	transactions, err := dbService.GetTransactions(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Transaction{&transaction1}, transactions)

	accounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	expectedAccount1 := testAccount1
	expectedAccount1.Balance = -1
	expectedAccount2 := testAccount2
	expectedAccount2.Balance = 2
	assert.Equal(t, []*Account{&expectedAccount1, &expectedAccount2}, accounts)

	transaction2 := Transaction{
		Description: "t2",
		Date:        "2019-03-21",
		Type:        TransactionTypeTransfer,
		Tags:        []string{"t1", "t3"},
		Components: []TransactionComponent{
			TransactionComponent{AccountID: 0, Amount: 100},
			TransactionComponent{AccountID: 0, Amount: 100},
			TransactionComponent{AccountID: 1, Amount: 100},
		},
	}

	saveTransaction = transaction2
	err = dbService.CreateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), saveTransaction.ID)

	transaction2.ID = 1
	transactions, err = dbService.GetTransactions(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Transaction{&transaction1, &transaction2}, transactions)

	accounts, err = dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	expectedAccount1.Balance = -1 + 100 + 100
	expectedAccount2.Balance = 2 + 100
	assert.Equal(t, []*Account{&expectedAccount1, &expectedAccount2}, accounts)
}

func TestUpdateTransactionWithComponents(t *testing.T) {
	dbService, cleanup, err := createDb()
	assert.NoError(t, err)
	defer cleanup()

	err = createTestAccounts(dbService)
	assert.NoError(t, err)

	transaction1 := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"t1", "t2"},
		Components: []TransactionComponent{
			TransactionComponent{AccountID: 0, Amount: -1},
			TransactionComponent{AccountID: 1, Amount: 2},
		},
	}
	transaction2 := Transaction{
		Description: "t2",
		Date:        "2019-03-21",
		Type:        TransactionTypeTransfer,
		Tags:        []string{"t1", "t3"},
		Components: []TransactionComponent{
			TransactionComponent{AccountID: 0, Amount: 100},
			TransactionComponent{AccountID: 0, Amount: 100},
			TransactionComponent{AccountID: 1, Amount: 100},
		},
	}

	saveTransaction := transaction1
	err = dbService.CreateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)

	saveTransaction = transaction2
	err = dbService.CreateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)

	transaction2.Components = []TransactionComponent{
		TransactionComponent{AccountID: 0, Amount: 5},
		TransactionComponent{AccountID: 0, Amount: 10},
		TransactionComponent{AccountID: 0, Amount: 17},
		TransactionComponent{AccountID: 1, Amount: 4},
	}
	transaction2.ID = 1

	saveTransaction = transaction1
	err = dbService.UpdateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)

	saveTransaction = transaction2
	err = dbService.UpdateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)

	transactions, err := dbService.GetTransactions(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Transaction{&transaction1, &transaction2}, transactions)

	accounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	expectedAccount1 := testAccount1
	expectedAccount2 := testAccount2
	expectedAccount1.Balance = -1 + 5 + 10 + 17
	expectedAccount2.Balance = 2 + 4
	assert.Equal(t, []*Account{&expectedAccount1, &expectedAccount2}, accounts)
}

func TestDeleteTransactionWithComponents(t *testing.T) {
	dbService, cleanup, err := createDb()
	assert.NoError(t, err)
	defer cleanup()

	err = createTestAccounts(dbService)
	assert.NoError(t, err)

	transaction1 := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{"t1", "t2"},
		Components: []TransactionComponent{
			TransactionComponent{AccountID: 0, Amount: -1},
			TransactionComponent{AccountID: 1, Amount: 2},
		},
	}
	transaction2 := Transaction{
		Description: "t2",
		Date:        "2019-03-21",
		Type:        TransactionTypeTransfer,
		Tags:        []string{"t1", "t3"},
		Components: []TransactionComponent{
			TransactionComponent{AccountID: 0, Amount: 100},
			TransactionComponent{AccountID: 0, Amount: 100},
			TransactionComponent{AccountID: 1, Amount: 100},
		},
	}

	saveTransaction := transaction1
	err = dbService.CreateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)

	saveTransaction = transaction2
	err = dbService.CreateTransaction(&testUser, &saveTransaction)
	assert.NoError(t, err)

	transaction2.ID = 1

	err = dbService.DeleteTransaction(&testUser, &transaction2)
	assert.NoError(t, err)

	transactions, err := dbService.GetTransactions(&testUser)
	assert.NoError(t, err)
	assert.Equal(t, []*Transaction{&transaction1}, transactions)

	accounts, err := dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	expectedAccount1 := testAccount1
	expectedAccount2 := testAccount2
	expectedAccount1.Balance = -1
	expectedAccount2.Balance = 2
	assert.Equal(t, []*Account{&expectedAccount1, &expectedAccount2}, accounts)

	err = dbService.DeleteTransaction(&testUser, &transaction1)
	assert.NoError(t, err)

	transactions, err = dbService.GetTransactions(&testUser)
	assert.NoError(t, err)
	assert.Empty(t, transactions)

	accounts, err = dbService.GetAccounts(&testUser)
	assert.NoError(t, err)
	expectedAccount1.Balance = 0
	expectedAccount2.Balance = 0
	assert.Equal(t, []*Account{&expectedAccount1, &expectedAccount2}, accounts)
}

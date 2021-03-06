package data

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTransactionsFilterDescription(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	transaction := Transaction{
		Date: "2019-03-20",
	}

	transactions := make([]*Transaction, 100)
	for i := 0; i < 100; i++ {
		saveTransaction := transaction
		saveTransaction.Description = "ta" + strconv.Itoa(i)
		err = dbService.CreateTransaction(&testUser, &saveTransaction)
		assert.NoError(t, err)
		transactions[99-i] = &saveTransaction
	}

	getTransactionOptions := GetAllTransactionsOptions
	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterDescription: "ta1"}
	dbTransactions, err := dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, append(transactions[80:90], transactions[98]), dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterDescription: "A1"}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, append(transactions[80:90], transactions[98]), dbTransactions)
}

func TestGetTransactionsFilterDate(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	transaction := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
	}

	transactions := make([]*Transaction, 10)
	for i := 0; i < 10; i++ {
		saveTransaction := transaction
		saveTransaction.Date = "2019-03-2" + strconv.Itoa(i)
		err = dbService.CreateTransaction(&testUser, &saveTransaction)
		assert.NoError(t, err)
		transactions[9-i] = &saveTransaction
	}

	getTransactionOptions := GetAllTransactionsOptions
	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterFromDate: "2019-03-25"}
	dbTransactions, err := dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, transactions[:5], dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterToDate: "2019-03-24"}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, transactions[5:], dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterFromDate: "2019-03-01", FilterToDate: "2019-03-24"}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, transactions[5:], dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterFromDate: "2019-03-25", FilterToDate: "2020-03-24"}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, transactions[:5], dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterFromDate: "2019-03-21", FilterToDate: "2020-03-29"}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, transactions[:9], dbTransactions)
}

func TestGetTransactionsFilterTags(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	transaction := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
	}

	transactions := make([]*Transaction, 20)
	for i := 0; i < 20; i++ {
		saveTransaction := transaction
		saveTransaction.Tags = []string{"t1", "a" + strconv.Itoa(i)}
		err = dbService.CreateTransaction(&testUser, &saveTransaction)
		assert.NoError(t, err)
		transactions[19-i] = &saveTransaction
	}

	getTransactionOptions := GetAllTransactionsOptions
	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterTags: []string{"t1"}}
	dbTransactions, err := dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, transactions, dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterTags: []string{"t1", "b1"}}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, transactions, dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterTags: []string{"a1"}}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, transactions[18:19], dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterTags: []string{"a1", "a2"}}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, transactions[17:19], dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterTags: []string{"A1"}}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Empty(t, dbTransactions)
}

func TestGetTransactionsFilterAccounts(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	transaction := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
	}

	transactions := make([]*Transaction, 10)
	for i := 0; i < 10; i++ {
		saveTransaction := transaction
		saveTransaction.Components = []TransactionComponent{
			{AccountUUID: fmt.Sprintf("uuid%v", i)},
			{AccountUUID: "uuid42"},
		}
		err = dbService.CreateTransaction(&testUser, &saveTransaction)
		assert.NoError(t, err)
		transactions[9-i] = &saveTransaction
	}

	getTransactionOptions := GetAllTransactionsOptions
	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterAccounts: []string{"uuid42"}}
	dbTransactions, err := dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, transactions, dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterAccounts: []string{"uuid42", "uuid88"}}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, transactions, dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterAccounts: []string{"uuid1"}}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, transactions[8:9], dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterAccounts: []string{"uuid1", "uuid2"}}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, transactions[7:9], dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{FilterAccounts: []string{"uuid88"}}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Empty(t, dbTransactions)
}

func TestGetTransactionsFilterType(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	transaction := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
	}

	transactions := make([]*Transaction, 4)
	for i := 0; i < 4; i++ {
		saveTransaction := transaction
		if i%2 == 0 {
			saveTransaction.Type = TransactionTypeExpenseIncome
		} else if i%2 == 1 {
			saveTransaction.Type = TransactionTypeTransfer
		}
		err = dbService.CreateTransaction(&testUser, &saveTransaction)
		assert.NoError(t, err)
		transactions[i] = &saveTransaction
	}

	getTransactionOptions := GetAllTransactionsOptions
	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{}
	dbTransactions, err := dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, []*Transaction{transactions[3], transactions[2], transactions[1], transactions[0]}, dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{ExcludeExpenseIncome: true}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, []*Transaction{transactions[3], transactions[1]}, dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{ExcludeTransfer: true}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Equal(t, []*Transaction{transactions[2], transactions[0]}, dbTransactions)

	getTransactionOptions.TransactionFilterOptions = TransactionFilterOptions{ExcludeExpenseIncome: true, ExcludeTransfer: true}
	dbTransactions, err = dbService.GetTransactions(&testUser, getTransactionOptions)
	assert.NoError(t, err)
	assert.Empty(t, dbTransactions)
}

func TestCountTransactionsFilterDescription(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	transaction := Transaction{
		Date: "2019-03-20",
	}

	for i := 0; i < 100; i++ {
		saveTransaction := transaction
		saveTransaction.Description = "ta" + strconv.Itoa(i)
		err = dbService.CreateTransaction(&testUser, &saveTransaction)
		assert.NoError(t, err)
	}

	filterOptions := TransactionFilterOptions{FilterDescription: "ta1"}
	count, err := dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(11), count)

	filterOptions = TransactionFilterOptions{FilterDescription: "A1"}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(11), count)
}

func TestCountTransactionsFilterDate(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	transaction := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
	}

	for i := 0; i < 10; i++ {
		saveTransaction := transaction
		saveTransaction.Date = "2019-03-2" + strconv.Itoa(i)
		err = dbService.CreateTransaction(&testUser, &saveTransaction)
		assert.NoError(t, err)
	}

	filterOptions := TransactionFilterOptions{FilterFromDate: "2019-03-25"}
	count, err := dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), count)

	filterOptions = TransactionFilterOptions{FilterToDate: "2019-03-24"}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), count)

	filterOptions = TransactionFilterOptions{FilterFromDate: "2019-03-01", FilterToDate: "2019-03-24"}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), count)

	filterOptions = TransactionFilterOptions{FilterFromDate: "2019-03-25", FilterToDate: "2020-03-24"}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), count)

	filterOptions = TransactionFilterOptions{FilterFromDate: "2019-03-21", FilterToDate: "2020-03-29"}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(9), count)
}

func TestCountTransactionsFilterTags(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	transaction := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
	}

	for i := 0; i < 20; i++ {
		saveTransaction := transaction
		saveTransaction.Tags = []string{"t1", "a" + strconv.Itoa(i)}
		err = dbService.CreateTransaction(&testUser, &saveTransaction)
		assert.NoError(t, err)
	}

	filterOptions := TransactionFilterOptions{FilterTags: []string{"t1"}}
	count, err := dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(20), count)

	filterOptions = TransactionFilterOptions{FilterTags: []string{"t1", "b1"}}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(20), count)

	filterOptions = TransactionFilterOptions{FilterTags: []string{"a1"}}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), count)

	filterOptions = TransactionFilterOptions{FilterTags: []string{"a1", "a2"}}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), count)

	filterOptions = TransactionFilterOptions{FilterTags: []string{"A1"}}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), count)
}

func TestCountTransactionsFilterAccounts(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	transaction := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
	}

	for i := 0; i < 10; i++ {
		saveTransaction := transaction
		saveTransaction.Components = []TransactionComponent{
			{AccountUUID: fmt.Sprintf("uuid%v", i)},
			{AccountUUID: "uuid42"},
		}
		err = dbService.CreateTransaction(&testUser, &saveTransaction)
		assert.NoError(t, err)
	}

	filterOptions := TransactionFilterOptions{FilterAccounts: []string{"uuid42"}}
	count, err := dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), count)

	filterOptions = TransactionFilterOptions{FilterAccounts: []string{"uuid42", "uuid88"}}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), count)

	filterOptions = TransactionFilterOptions{FilterAccounts: []string{"uuid1"}}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), count)

	filterOptions = TransactionFilterOptions{FilterAccounts: []string{"uuid1", "uuid2"}}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), count)

	filterOptions = TransactionFilterOptions{FilterAccounts: []string{"uuid88"}}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), count)
}

func TestCountTransactionsFilterType(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	transaction := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
	}

	for i := 0; i < 4; i++ {
		saveTransaction := transaction
		if i%2 == 0 {
			saveTransaction.Type = TransactionTypeExpenseIncome
		} else if i%2 == 1 {
			saveTransaction.Type = TransactionTypeTransfer
		}
		err = dbService.CreateTransaction(&testUser, &saveTransaction)
		assert.NoError(t, err)
	}

	filterOptions := TransactionFilterOptions{}
	count, err := dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(4), count)

	filterOptions = TransactionFilterOptions{ExcludeExpenseIncome: true}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), count)

	filterOptions = TransactionFilterOptions{ExcludeTransfer: true}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), count)

	filterOptions = TransactionFilterOptions{ExcludeExpenseIncome: true, ExcludeTransfer: true}
	count, err = dbService.CountTransactions(&testUser, filterOptions)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), count)
}

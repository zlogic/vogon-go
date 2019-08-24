package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTags(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

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
		Tags:        []string{"a1", "t1"},
	}

	err = dbService.CreateTransaction(testUser, &transaction1)
	assert.NoError(t, err)
	err = dbService.CreateTransaction(testUser, &transaction2)
	assert.NoError(t, err)

	tags, err := dbService.GetTags(testUser)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"a1", "t1", "t2"}, tags)
}

func TestGetTransactionsWithoutTags(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	transaction1 := Transaction{
		Description: "t1",
		Date:        "2019-03-20",
		Type:        TransactionTypeExpenseIncome,
		Tags:        []string{},
	}
	transaction2 := Transaction{
		Description: "t2",
		Date:        "2019-03-21",
		Type:        TransactionTypeTransfer,
		Tags:        []string{},
	}

	err = dbService.CreateTransaction(testUser, &transaction1)
	assert.NoError(t, err)
	err = dbService.CreateTransaction(testUser, &transaction2)
	assert.NoError(t, err)

	tags, err := dbService.GetTags(testUser)
	assert.NoError(t, err)
	assert.Empty(t, tags)
}

func TestGetTagsNoTransactions(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	tags, err := dbService.GetTags(testUser)
	assert.NoError(t, err)
	assert.Empty(t, tags)
}

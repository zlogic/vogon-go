package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zlogic/vogon-go/data"
)

func createTestTransactions() []*data.Transaction {
	return []*data.Transaction{&data.Transaction{
		ID:          0,
		Description: "Widgets",
		Type:        data.TransactionTypeExpenseIncome,
		Tags:        []string{"Widgets"},
		Date:        "2015-11-02",
		Components: []data.TransactionComponent{
			data.TransactionComponent{AccountID: 1, Amount: -10000},
		},
	}, &data.Transaction{
		ID:          1,
		Description: "Salary",
		Type:        data.TransactionTypeExpenseIncome,
		Tags:        []string{"Salary"},
		Date:        "2015-11-01",
		Components: []data.TransactionComponent{
			data.TransactionComponent{AccountID: 0, Amount: 100000},
			data.TransactionComponent{AccountID: 1, Amount: 100000},
		},
	}, &data.Transaction{
		ID:          2,
		Description: "Gadgets",
		Type:        data.TransactionTypeExpenseIncome,
		Tags:        []string{"Gadgets", "Widgets"},
		Date:        "2015-11-03",
		Components:  []data.TransactionComponent{},
	}, &data.Transaction{
		ID:          3,
		Description: "Gadgets 2",
		Type:        data.TransactionTypeTransfer,
		Tags:        []string{"Gadgets"},
		Date:        "2015-11-03",
		Components:  []data.TransactionComponent{},
	}}
}

func TestGetTransactionsAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	req, _ := http.NewRequest("GET", "/api/transactions", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	transactions := createTestTransactions()
	dbMock.On("GetTransactions", &user).Return(transactions, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "["+
		`{"ID":3,"Description":"Gadgets 2","Type":1,"Tags":["Gadgets"],"Date":"2015-11-03","Components":[]}`+","+
		`{"ID":2,"Description":"Gadgets","Type":0,"Tags":["Gadgets","Widgets"],"Date":"2015-11-03","Components":[]}`+","+
		`{"ID":0,"Description":"Widgets","Type":0,"Tags":["Widgets"],"Date":"2015-11-02","Components":[{"Amount":-10000,"AccountID":1}]}`+","+
		`{"ID":1,"Description":"Salary","Type":0,"Tags":["Salary"],"Date":"2015-11-01","Components":[{"Amount":100000,"AccountID":0},{"Amount":100000,"AccountID":1}]}`+
		"]\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestGetTransactionsUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/transactions", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestGetTransactionsUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	dbMock.On("GetUser", "user01").Return(nil, nil).Once()

	req, _ := http.NewRequest("GET", "/api/transactions", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

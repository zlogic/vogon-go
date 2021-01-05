package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zlogic/vogon-go/data"
)

func createTestTransactions() []*data.Transaction {
	return []*data.Transaction{{
		ID:          3,
		Description: "Gadgets 2",
		Type:        data.TransactionTypeTransfer,
		Tags:        []string{"Gadgets"},
		Date:        "2015-11-03",
		Components:  []data.TransactionComponent{},
	}, {
		ID:          2,
		Description: "Gadgets",
		Type:        data.TransactionTypeExpenseIncome,
		Tags:        []string{"Gadgets", "Widgets"},
		Date:        "2015-11-03",
		Components:  []data.TransactionComponent{},
	}, {
		ID:          0,
		Description: "Widgets",
		Type:        data.TransactionTypeExpenseIncome,
		Tags:        []string{"Widgets"},
		Date:        "2015-11-02",
		Components: []data.TransactionComponent{
			{AccountID: 1, Amount: -10000},
		},
	}, {
		ID:          1,
		Description: "Salary",
		Type:        data.TransactionTypeExpenseIncome,
		Tags:        []string{"Salary"},
		Date:        "2015-11-01",
		Components: []data.TransactionComponent{
			{AccountID: 0, Amount: 100000},
			{AccountID: 1, Amount: 100000},
		},
	}}
}

func createTestTransaction() *data.Transaction {
	return &data.Transaction{
		ID:          42,
		Description: "Widgets",
		Type:        data.TransactionTypeExpenseIncome,
		Tags:        []string{"Widgets"},
		Date:        "2015-11-02",
		Components: []data.TransactionComponent{
			{AccountID: 1, Amount: -10000},
		},
	}
}

func TestGetTransactionsAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/transactions/getpage", strings.NewReader("offset=0&limit=10"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	transactions := createTestTransactions()
	options := data.GetTransactionOptions{Offset: 0, Limit: 10}
	dbMock.On("GetTransactions", &user, options).Return(transactions, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "["+
		`{"ID":3,"Description":"Gadgets 2","Type":1,"Tags":["Gadgets"],"Date":"2015-11-03","Components":[]}`+","+
		`{"ID":2,"Description":"Gadgets","Type":0,"Tags":["Gadgets","Widgets"],"Date":"2015-11-03","Components":[]}`+","+
		`{"ID":0,"Description":"Widgets","Type":0,"Tags":["Widgets"],"Date":"2015-11-02","Components":[{"Amount":-10000,"AccountID":1}]}`+","+
		`{"ID":1,"Description":"Salary","Type":0,"Tags":["Salary"],"Date":"2015-11-01","Components":[{"Amount":100000,"AccountID":0},{"Amount":100000,"AccountID":1}]}`+
		"]\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetTransactionsFilterAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/transactions/getpage", strings.NewReader("offset=0&limit=10&filterDescription=d1&filterFrom=f1&filterTo=t1&filterTags=s1,s2&filterAccounts=1,2&filterIncludeExpenseIncome=false&filterIncludeTransfer=false"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	transactions := createTestTransactions()
	options := data.GetTransactionOptions{
		Offset: 0,
		Limit:  10,
		TransactionFilterOptions: data.TransactionFilterOptions{
			FilterDescription:    "d1",
			FilterFromDate:       "f1",
			FilterToDate:         "t1",
			FilterTags:           []string{"s1", "s2"},
			FilterAccounts:       []uint64{1, 2},
			ExcludeExpenseIncome: true,
			ExcludeTransfer:      true,
		},
	}
	dbMock.On("GetTransactions", &user, options).Return(transactions, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "["+
		`{"ID":3,"Description":"Gadgets 2","Type":1,"Tags":["Gadgets"],"Date":"2015-11-03","Components":[]}`+","+
		`{"ID":2,"Description":"Gadgets","Type":0,"Tags":["Gadgets","Widgets"],"Date":"2015-11-03","Components":[]}`+","+
		`{"ID":0,"Description":"Widgets","Type":0,"Tags":["Widgets"],"Date":"2015-11-02","Components":[{"Amount":-10000,"AccountID":1}]}`+","+
		`{"ID":1,"Description":"Salary","Type":0,"Tags":["Salary"],"Date":"2015-11-01","Components":[{"Amount":100000,"AccountID":0},{"Amount":100000,"AccountID":1}]}`+
		"]\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}
func TestGetTransactionsUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/transactions/getpage", strings.NewReader("offset=0&limit=0"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetTransactionsUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/transactions/getpage", strings.NewReader("offset=0&limit=0"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetTransactionsCountAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/transactions/getcount", strings.NewReader(""))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	options := data.TransactionFilterOptions{}
	dbMock.On("CountTransactions", &user, options).Return(uint64(123), nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "123\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetTransactionsCountFilterAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/transactions/getcount", strings.NewReader("filterDescription=d1&filterFrom=f1&filterTo=t1&filterTags=s1,s2&filterAccounts=1,2&filterIncludeExpenseIncome=false&filterIncludeTransfer=false"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	options := data.TransactionFilterOptions{
		FilterDescription:    "d1",
		FilterFromDate:       "f1",
		FilterToDate:         "t1",
		FilterTags:           []string{"s1", "s2"},
		FilterAccounts:       []uint64{1, 2},
		ExcludeExpenseIncome: true,
		ExcludeTransfer:      true,
	}
	dbMock.On("CountTransactions", &user, options).Return(uint64(123), nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "123\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetTransactionsCountUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/transactions/getcount", strings.NewReader(""))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetTransactionsCountUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/transactions/getcount", strings.NewReader(""))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}
func TestGetTransactionAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/transaction/42", nil)
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	transaction := createTestTransaction()
	dbMock.On("GetTransaction", &user, uint64(42)).Return(transaction, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"ID":42,"Description":"Widgets","Type":0,"Tags":["Widgets"],"Date":"2015-11-02","Components":[{"Amount":-10000,"AccountID":1}]}`+"\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetTransactionUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/transaction/42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetTransactionUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/transaction/42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestDeleteTransactionAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("DELETE", "/api/transaction/42", nil)
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	dbMock.On("DeleteTransaction", &user, uint64(42)).Return(nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "OK", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestDeleteTransactionUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("DELETE", "/api/transaction/42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestDeleteTransactionUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("DELETE", "/api/transaction/42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestPostCreateTransactionAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/transaction/new", strings.NewReader(`{"ID":42,"Description":"Widgets","Type":0,"Tags":["Widgets"],"Date":"2015-11-02","Components":[{"Amount":-10000,"AccountID":1}]}`))
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	transaction := createTestTransaction()
	dbMock.On("CreateTransaction", &user, transaction).Return(nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "OK", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestPostUpdateTransactionAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/transaction/42", strings.NewReader(`{"ID":42,"Description":"Widgets","Type":0,"Tags":["Widgets"],"Date":"2015-11-02","Components":[{"Amount":-10000,"AccountID":1}]}`))
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	transaction := createTestTransaction()
	dbMock.On("UpdateTransaction", &user, transaction).Return(nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "OK", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestPostTransactionUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/transaction/42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestPostTransactionUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/transaction/42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

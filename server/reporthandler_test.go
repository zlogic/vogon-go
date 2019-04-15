package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zlogic/vogon-go/data"
)

func createReportTransactions() []*data.Transaction {
	return []*data.Transaction{
		&data.Transaction{
			Description: "Unrelated",
			Type:        data.TransactionTypeExpenseIncome,
			Tags:        []string{"Something"},
			Date:        "2015-11-07",
			Components: []data.TransactionComponent{
				data.TransactionComponent{AccountID: 3, Amount: -5000},
			},
		},
		&data.Transaction{
			Description: "Another transfer",
			Type:        data.TransactionTypeTransfer,
			Tags:        []string{"Transfer"},
			Date:        "2015-11-06",
			Components: []data.TransactionComponent{
				data.TransactionComponent{AccountID: 0, Amount: -7000},
				data.TransactionComponent{AccountID: 2, Amount: 100},
			},
		},
		&data.Transaction{
			Description: "More stuff",
			Type:        data.TransactionTypeExpenseIncome,
			Tags:        []string{"Gadgets", "Widgets"},
			Date:        "2015-11-05",
			Components: []data.TransactionComponent{
				data.TransactionComponent{AccountID: 1, Amount: -5000},
			},
		},
		&data.Transaction{
			Description: "Widgets",
			Type:        data.TransactionTypeExpenseIncome,
			Tags:        []string{"Widgets"},
			Date:        "2015-11-04",
			Components: []data.TransactionComponent{
				data.TransactionComponent{AccountID: 0, Amount: -3000},
			},
		},
		&data.Transaction{
			Description: "Gadgets",
			Type:        data.TransactionTypeExpenseIncome,
			Tags:        []string{"Gadgets"},
			Date:        "2015-11-04",
			Components: []data.TransactionComponent{
				data.TransactionComponent{AccountID: 0, Amount: -3000},
			},
		},
		&data.Transaction{
			Description: "Stuff",
			Type:        data.TransactionTypeExpenseIncome,
			Tags:        []string{"Gadgets", "Widgets"},
			Date:        "2015-11-03",
			Components: []data.TransactionComponent{
				data.TransactionComponent{AccountID: 0, Amount: -2000},
				data.TransactionComponent{AccountID: 1, Amount: -2000},
				data.TransactionComponent{AccountID: 2, Amount: -2000},
			},
		},
		&data.Transaction{
			Description: "Transfer",
			Type:        data.TransactionTypeTransfer,
			Tags:        []string{"Transfer"},
			Date:        "2015-11-02",
			Components: []data.TransactionComponent{
				data.TransactionComponent{AccountID: 0, Amount: -1000},
				data.TransactionComponent{AccountID: 1, Amount: 1000},
			},
		},
		&data.Transaction{
			Description: "Salary",
			Type:        data.TransactionTypeExpenseIncome,
			Tags:        []string{"Salary"},
			Date:        "2015-11-01",
			Components: []data.TransactionComponent{
				data.TransactionComponent{AccountID: 0, Amount: 100000},
				data.TransactionComponent{AccountID: 1, Amount: 100000},
				data.TransactionComponent{AccountID: 2, Amount: 100000},
			},
		},
	}
}

func TestReportEverything(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	req, _ := http.NewRequest("POST", "/api/report", strings.NewReader(""))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	transactions := createReportTransactions()
	options := data.GetAllTransactionsOptions
	dbMock.On("GetTransactions", &user, options).Return(transactions, nil).Once()

	accounts := []*data.Account{
		&data.Account{ID: 0, Name: "a1", Currency: "USD"},
		&data.Account{ID: 1, Name: "a2", Currency: "USD"},
		&data.Account{ID: 2, Name: "a3", Currency: "EUR"},
		&data.Account{ID: 3, Name: "a4", Currency: "EUR"},
	}
	dbMock.On("GetAccounts", &user).Return(accounts, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"BalanceChart":{`+
		`"EUR":{"2015-11-01":100000,"2015-11-03":98000,"2015-11-06":98100,"2015-11-07":93100},`+
		`"USD":{"2015-11-01":200000,"2015-11-02":200000,"2015-11-03":196000,"2015-11-04":190000,"2015-11-05":185000,"2015-11-06":178000}`+
		`},"TagsChart":{`+
		`"EUR":{"Positive":{"Gadgets,Widgets":0,"Salary":100000,"Something":0},"Negative":{"Gadgets,Widgets":2000,"Salary":0,"Something":5000},"Transfer":{"Transfer":100}},`+
		`"USD":{"Positive":{"Gadgets":0,"Gadgets,Widgets":0,"Salary":200000,"Widgets":0},"Negative":{"Gadgets":3000,"Gadgets,Widgets":9000,"Salary":0,"Widgets":3000},"Transfer":{"Transfer":8000}}`+
		"}}\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestReportFilterDescription(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	req, _ := http.NewRequest("POST", "/api/report", strings.NewReader("filterDescription=stuff&filterAccounts=0,1,2"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	transactions := createReportTransactions()
	options := data.GetAllTransactionsOptions
	dbMock.On("GetTransactions", &user, options).Return(transactions, nil).Once()

	accounts := []*data.Account{
		&data.Account{ID: 0, Name: "a1", Currency: "USD"},
		&data.Account{ID: 1, Name: "a2", Currency: "USD"},
		&data.Account{ID: 2, Name: "a3", Currency: "EUR"},
	}
	dbMock.On("GetAccounts", &user).Return(accounts, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"BalanceChart":{`+
		`"EUR":{"2015-11-01":100000,"2015-11-03":98000,"2015-11-06":98100},`+
		`"USD":{"2015-11-01":200000,"2015-11-02":200000,"2015-11-03":196000,"2015-11-04":190000,"2015-11-05":185000,"2015-11-06":178000}`+
		`},"TagsChart":{`+
		`"EUR":{"Positive":{"Gadgets,Widgets":0},"Negative":{"Gadgets,Widgets":2000},"Transfer":{}},`+
		`"USD":{"Positive":{"Gadgets,Widgets":0},"Negative":{"Gadgets,Widgets":9000},"Transfer":{}}`+
		"}}\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestReportDateRange(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	req, _ := http.NewRequest("POST", "/api/report", strings.NewReader("filterAccounts=0,1,2&filterFrom=2015-11-02&filterTo=2015-11-04"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	transactions := createReportTransactions()
	options := data.GetAllTransactionsOptions
	dbMock.On("GetTransactions", &user, options).Return(transactions, nil).Once()

	accounts := []*data.Account{
		&data.Account{ID: 0, Name: "a1", Currency: "USD"},
		&data.Account{ID: 1, Name: "a2", Currency: "USD"},
		&data.Account{ID: 2, Name: "a3", Currency: "EUR"},
		&data.Account{ID: 3, Name: "a4", Currency: "EUR"},
	}
	dbMock.On("GetAccounts", &user).Return(accounts, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"BalanceChart":{`+
		`"EUR":{"2015-11-03":98000},`+
		`"USD":{"2015-11-02":200000,"2015-11-03":196000,"2015-11-04":190000}`+
		`},"TagsChart":{`+
		`"EUR":{"Positive":{"Gadgets,Widgets":0},"Negative":{"Gadgets,Widgets":2000},"Transfer":{}},`+
		`"USD":{"Positive":{"Gadgets":0,"Gadgets,Widgets":0,"Widgets":0},"Negative":{"Gadgets":3000,"Gadgets,Widgets":4000,"Widgets":3000},"Transfer":{"Transfer":1000}}`+
		"}}\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestReportFilterTags(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	req, _ := http.NewRequest("POST", "/api/report", strings.NewReader("filterAccounts=0,1,2&filterTags=Gadgets,Widgets"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	transactions := createReportTransactions()
	options := data.GetAllTransactionsOptions
	dbMock.On("GetTransactions", &user, options).Return(transactions, nil).Once()

	accounts := []*data.Account{
		&data.Account{ID: 0, Name: "a1", Currency: "USD"},
		&data.Account{ID: 1, Name: "a2", Currency: "USD"},
		&data.Account{ID: 2, Name: "a3", Currency: "EUR"},
		&data.Account{ID: 3, Name: "a4", Currency: "EUR"},
	}
	dbMock.On("GetAccounts", &user).Return(accounts, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"BalanceChart":{`+
		`"EUR":{"2015-11-01":100000,"2015-11-03":98000,"2015-11-06":98100},`+
		`"USD":{"2015-11-01":200000,"2015-11-02":200000,"2015-11-03":196000,"2015-11-04":190000,"2015-11-05":185000,"2015-11-06":178000}`+
		`},"TagsChart":{`+
		`"EUR":{"Positive":{"Gadgets,Widgets":0},"Negative":{"Gadgets,Widgets":2000},"Transfer":{}},`+
		`"USD":{"Positive":{"Gadgets":0,"Gadgets,Widgets":0,"Widgets":0},"Negative":{"Gadgets":3000,"Gadgets,Widgets":9000,"Widgets":3000},"Transfer":{}}`+
		"}}\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestReportAccounts012(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	req, _ := http.NewRequest("POST", "/api/report", strings.NewReader("filterAccounts=0,1,2"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	transactions := createReportTransactions()
	options := data.GetAllTransactionsOptions
	dbMock.On("GetTransactions", &user, options).Return(transactions, nil).Once()

	accounts := []*data.Account{
		&data.Account{ID: 0, Name: "a1", Currency: "USD"},
		&data.Account{ID: 1, Name: "a2", Currency: "USD"},
		&data.Account{ID: 2, Name: "a3", Currency: "EUR"},
		&data.Account{ID: 3, Name: "a4", Currency: "EUR"},
	}
	dbMock.On("GetAccounts", &user).Return(accounts, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"BalanceChart":{`+
		`"EUR":{"2015-11-01":100000,"2015-11-03":98000,"2015-11-06":98100},`+
		`"USD":{"2015-11-01":200000,"2015-11-02":200000,"2015-11-03":196000,"2015-11-04":190000,"2015-11-05":185000,"2015-11-06":178000}`+
		`},"TagsChart":{`+
		`"EUR":{"Positive":{"Gadgets,Widgets":0,"Salary":100000},"Negative":{"Gadgets,Widgets":2000,"Salary":0},"Transfer":{"Transfer":100}},`+
		`"USD":{"Positive":{"Gadgets":0,"Gadgets,Widgets":0,"Salary":200000,"Widgets":0},"Negative":{"Gadgets":3000,"Gadgets,Widgets":9000,"Salary":0,"Widgets":3000},"Transfer":{"Transfer":8000}}`+
		"}}\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestReportOnlyAccount0(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	req, _ := http.NewRequest("POST", "/api/report", strings.NewReader("filterAccounts=0"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	transactions := createReportTransactions()
	options := data.GetAllTransactionsOptions
	dbMock.On("GetTransactions", &user, options).Return(transactions, nil).Once()

	accounts := []*data.Account{
		&data.Account{ID: 0, Name: "a1", Currency: "USD"},
		&data.Account{ID: 1, Name: "a2", Currency: "USD"},
		&data.Account{ID: 2, Name: "a3", Currency: "EUR"},
		&data.Account{ID: 3, Name: "a4", Currency: "EUR"},
	}
	dbMock.On("GetAccounts", &user).Return(accounts, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"BalanceChart":{`+
		`"USD":{"2015-11-01":100000,"2015-11-02":99000,"2015-11-03":97000,"2015-11-04":91000,"2015-11-06":84000}`+
		`},"TagsChart":{`+
		`"USD":{"Positive":{"Gadgets":0,"Gadgets,Widgets":0,"Salary":100000,"Widgets":0},"Negative":{"Gadgets":3000,"Gadgets,Widgets":2000,"Salary":0,"Widgets":3000},"Transfer":{"Transfer":8000}}`+
		"}}\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestReportOnlyAccount1(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	req, _ := http.NewRequest("POST", "/api/report", strings.NewReader("filterAccounts=1"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	transactions := createReportTransactions()
	options := data.GetAllTransactionsOptions
	dbMock.On("GetTransactions", &user, options).Return(transactions, nil).Once()

	accounts := []*data.Account{
		&data.Account{ID: 0, Name: "a1", Currency: "USD"},
		&data.Account{ID: 1, Name: "a2", Currency: "USD"},
		&data.Account{ID: 2, Name: "a3", Currency: "EUR"},
		&data.Account{ID: 3, Name: "a4", Currency: "EUR"},
	}
	dbMock.On("GetAccounts", &user).Return(accounts, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"BalanceChart":{`+
		`"USD":{"2015-11-01":100000,"2015-11-02":101000,"2015-11-03":99000,"2015-11-05":94000}`+
		`},"TagsChart":{`+
		`"USD":{"Positive":{"Gadgets,Widgets":0,"Salary":100000},"Negative":{"Gadgets,Widgets":7000,"Salary":0},"Transfer":{"Transfer":1000}}`+
		"}}\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestReportFilterExcludeTransfer(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	req, _ := http.NewRequest("POST", "/api/report", strings.NewReader("filterAccounts=0,1,2&filterIncludeTransfer=false"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	transactions := createReportTransactions()
	options := data.GetAllTransactionsOptions
	dbMock.On("GetTransactions", &user, options).Return(transactions, nil).Once()

	accounts := []*data.Account{
		&data.Account{ID: 0, Name: "a1", Currency: "USD"},
		&data.Account{ID: 1, Name: "a2", Currency: "USD"},
		&data.Account{ID: 2, Name: "a3", Currency: "EUR"},
		&data.Account{ID: 3, Name: "a4", Currency: "EUR"},
	}
	dbMock.On("GetAccounts", &user).Return(accounts, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"BalanceChart":{`+
		`"EUR":{"2015-11-01":100000,"2015-11-03":98000,"2015-11-06":98100},`+
		`"USD":{"2015-11-01":200000,"2015-11-02":200000,"2015-11-03":196000,"2015-11-04":190000,"2015-11-05":185000,"2015-11-06":178000}`+
		`},"TagsChart":{`+
		`"EUR":{"Positive":{"Gadgets,Widgets":0,"Salary":100000},"Negative":{"Gadgets,Widgets":2000,"Salary":0},"Transfer":{}},`+
		`"USD":{"Positive":{"Gadgets":0,"Gadgets,Widgets":0,"Salary":200000,"Widgets":0},"Negative":{"Gadgets":3000,"Gadgets,Widgets":9000,"Salary":0,"Widgets":3000},"Transfer":{}}`+
		"}}\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestReportFilterExcludeExpenseIncome(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	req, _ := http.NewRequest("POST", "/api/report", strings.NewReader("filterAccounts=0,1,2&filterIncludeExpenseIncome=false"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	transactions := createReportTransactions()
	options := data.GetAllTransactionsOptions
	dbMock.On("GetTransactions", &user, options).Return(transactions, nil).Once()

	accounts := []*data.Account{
		&data.Account{ID: 0, Name: "a1", Currency: "USD"},
		&data.Account{ID: 1, Name: "a2", Currency: "USD"},
		&data.Account{ID: 2, Name: "a3", Currency: "EUR"},
		&data.Account{ID: 3, Name: "a4", Currency: "EUR"},
	}
	dbMock.On("GetAccounts", &user).Return(accounts, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"BalanceChart":{`+
		`"EUR":{"2015-11-01":100000,"2015-11-03":98000,"2015-11-06":98100},`+
		`"USD":{"2015-11-01":200000,"2015-11-02":200000,"2015-11-03":196000,"2015-11-04":190000,"2015-11-05":185000,"2015-11-06":178000}`+
		`},"TagsChart":{`+
		`"EUR":{"Positive":{},"Negative":{},"Transfer":{"Transfer":100}},`+
		`"USD":{"Positive":{},"Negative":{},"Transfer":{"Transfer":8000}}`+
		"}}\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}
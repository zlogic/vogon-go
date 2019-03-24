package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zlogic/vogon-go/data"
)

func TestGetAccountsAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	req, _ := http.NewRequest("GET", "/api/accounts", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	accounts := []*data.Account{
		&data.Account{ID: 0, Name: "a1", Currency: "USD", Balance: 100, IncludeInTotal: false, ShowInList: true},
		&data.Account{ID: 4, Name: "a2", Currency: "EUR", Balance: -4200, IncludeInTotal: true, ShowInList: false},
	}
	dbMock.On("GetAccounts", &user).Return(accounts, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "["+
		`{"ID":0,"Name":"a1","Balance":100,"Currency":"USD","IncludeInTotal":false,"ShowInList":true}`+","+
		`{"ID":4,"Name":"a2","Balance":-4200,"Currency":"EUR","IncludeInTotal":true,"ShowInList":false}`+
		"]\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestGetAccountsUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/accounts", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestGetAccountsUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	dbMock.On("GetUser", "user01").Return(nil, nil).Once()

	req, _ := http.NewRequest("GET", "/api/accounts", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

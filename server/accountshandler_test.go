package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zlogic/vogon-go/data"
)

func TestGetAccountsAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/accounts", nil)
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	accounts := []*data.Account{
		{UUID: "uuid1", Name: "a1", Currency: "USD", Balance: 100, IncludeInTotal: false, ShowInList: true},
		{UUID: "uuid5", Name: "a2", Currency: "EUR", Balance: -4200, IncludeInTotal: true, ShowInList: false},
	}
	dbMock.On("GetAccounts", &user).Return(accounts, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "["+
		`{"UUID":"uuid1","Name":"a1","Balance":100,"Currency":"USD","IncludeInTotal":false,"ShowInList":true}`+","+
		`{"UUID":"uuid5","Name":"a2","Balance":-4200,"Currency":"EUR","IncludeInTotal":true,"ShowInList":false}`+
		"]\n", res.Body.String())

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetAccountsUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/accounts", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", res.Body.String())

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetAccountAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/account/uuid42", nil)
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	account := &data.Account{UUID: "uuid42", Name: "a1", Currency: "USD", Balance: 100, IncludeInTotal: false, ShowInList: true}
	dbMock.On("GetAccount", &user, "uuid42").Return(account, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"UUID":"uuid42","Name":"a1","Balance":100,"Currency":"USD","IncludeInTotal":false,"ShowInList":true}`+"\n", res.Body.String())

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetAccountUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/account/uuid42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", res.Body.String())

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestDeleteAccountAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("DELETE", "/api/account/uuid42", nil)
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	dbMock.On("DeleteAccount", &user, "uuid42").Return(nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "OK", res.Body.String())

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestDeleteAccountUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("DELETE", "/api/account/uuid42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", res.Body.String())

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestPostCreateAccountAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/account/new", strings.NewReader(`{"UUID":"uuid42","Name":"a1","Balance":100,"Currency":"USD","IncludeInTotal":false,"ShowInList":true}`))
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	account := &data.Account{UUID: "uuid42", Name: "a1", Currency: "USD", Balance: 100, IncludeInTotal: false, ShowInList: true}
	dbMock.On("CreateAccount", &user, account).Return(nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "OK", res.Body.String())

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestPostUpdateAccountAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/account/uuid42", strings.NewReader(`{"UUID":"uuid42","Name":"a1","Balance":100,"Currency":"USD","IncludeInTotal":false,"ShowInList":true}`))
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	account := &data.Account{UUID: "uuid42", Name: "a1", Currency: "USD", Balance: 100, IncludeInTotal: false, ShowInList: true}
	dbMock.On("UpdateAccount", &user, account).Return(nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "OK", res.Body.String())

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestPostAccountUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/account/uuid42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", res.Body.String())

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

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
		{ID: 0, Name: "a1", Currency: "USD", Balance: 100, IncludeInTotal: false, ShowInList: true},
		{ID: 4, Name: "a2", Currency: "EUR", Balance: -4200, IncludeInTotal: true, ShowInList: false},
	}
	dbMock.On("GetAccounts", &user).Return(accounts, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "["+
		`{"ID":0,"Name":"a1","Balance":100,"Currency":"USD","IncludeInTotal":false,"ShowInList":true}`+","+
		`{"ID":4,"Name":"a2","Balance":-4200,"Currency":"EUR","IncludeInTotal":true,"ShowInList":false}`+
		"]\n", string(res.Body.Bytes()))

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
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetAccountsUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/accounts", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetAccountAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/account/42", nil)
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	account := &data.Account{ID: 42, Name: "a1", Currency: "USD", Balance: 100, IncludeInTotal: false, ShowInList: true}
	dbMock.On("GetAccount", &user, uint64(42)).Return(account, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"ID":42,"Name":"a1","Balance":100,"Currency":"USD","IncludeInTotal":false,"ShowInList":true}`+"\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetAccountUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/account/42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetAccountUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/account/42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestDeleteAccountAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("DELETE", "/api/account/42", nil)
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	dbMock.On("DeleteAccount", &user, uint64(42)).Return(nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "OK", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestDeleteAccountUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("DELETE", "/api/account/42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestDeleteAccountUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("DELETE", "/api/account/42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestPostCreateAccountAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/account/new", strings.NewReader(`{"ID":42,"Name":"a1","Balance":100,"Currency":"USD","IncludeInTotal":false,"ShowInList":true}`))
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	account := &data.Account{ID: 42, Name: "a1", Currency: "USD", Balance: 100, IncludeInTotal: false, ShowInList: true}
	dbMock.On("CreateAccount", &user, account).Return(nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "OK", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestPostUpdateAccountAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/account/42", strings.NewReader(`{"ID":42,"Name":"a1","Balance":100,"Currency":"USD","IncludeInTotal":false,"ShowInList":true}`))
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	account := &data.Account{ID: 42, Name: "a1", Currency: "USD", Balance: 100, IncludeInTotal: false, ShowInList: true}
	dbMock.On("UpdateAccount", &user, account).Return(nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "OK", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestPostAccountUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/account/42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestPostAccountUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/account/42", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

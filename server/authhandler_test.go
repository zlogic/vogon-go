package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zlogic/vogon-go/data"
)

func TestLoginHandlerSuccessful(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := &data.User{ID: 1}
	user.SetPassword("pass")
	dbMock.On("GetUser", "user01").Return(user, nil).Once()

	req, _ := http.NewRequest("POST", "/api/login", strings.NewReader("username=user01&password=pass"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "OK", string(res.Body.Bytes()))
	cookies := res.Result().Cookies()
	assert.Equal(t, 1, len(cookies))
	if len(cookies) > 0 {
		decodedCookie := UserCookie{}
		err := cookieHandler.secureCookie.Decode(AuthenticationCookie, cookies[0].Value, &decodedCookie)
		assert.NoError(t, err)
		assert.Equal(t, "user01", decodedCookie.Username)
	}

	dbMock.AssertExpectations(t)
}

func TestLoginHandlerIncorrectPassword(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := &data.User{ID: 1}
	user.SetPassword("pass")
	dbMock.On("GetUser", "user01").Return(user, nil).Once()

	req, _ := http.NewRequest("POST", "/api/login", strings.NewReader("username=user01&password=accessdenied"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))
	assert.Empty(t, res.Result().Cookies())

	dbMock.AssertExpectations(t)
}

func TestLoginHandlerUnknownUsername(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	dbMock.On("GetUser", "user02").Return(nil, nil).Once()

	req, _ := http.NewRequest("POST", "/api/login", strings.NewReader("username=user02&password=pass"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))
	assert.Empty(t, res.Result().Cookies())

	dbMock.AssertExpectations(t)
}

func TestRegisterHandlerSuccessful(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := &data.User{ID: 1}
	user.SetPassword("pass")
	dbMock.On("CreateUser", "user01").Return(user, nil).Once()

	dbMock.On("SaveNewUser", user).Return(nil).Once()

	req, _ := http.NewRequest("POST", "/api/register", strings.NewReader("username=user01&password=pass"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "OK", string(res.Body.Bytes()))
	cookies := res.Result().Cookies()
	assert.Equal(t, 1, len(cookies))
	if len(cookies) > 0 {
		decodedCookie := UserCookie{}
		err := cookieHandler.secureCookie.Decode(AuthenticationCookie, cookies[0].Value, &decodedCookie)
		assert.NoError(t, err)
		assert.Equal(t, "user01", decodedCookie.Username)
	}

	dbMock.AssertExpectations(t)
}

func TestRegisterHandlerUsernameAlreadyInUse(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := &data.User{ID: 1}
	user.SetPassword("pass")
	dbMock.On("CreateUser", "user01").Return(user, nil).Once()

	dbMock.On("SaveNewUser", user).Return(data.ErrUserAlreadyExists).Once()

	req, _ := http.NewRequest("POST", "/api/register", strings.NewReader("username=user01&password=pass"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "Username is already in use\n", string(res.Body.Bytes()))
	assert.Empty(t, res.Result().Cookies())

	dbMock.AssertExpectations(t)
}

func TestRegisterRegistrationNotAllowed(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	err = os.Setenv("ALLOW_REGISTRATION", "false")
	defer func() { os.Unsetenv("ALLOW_REGISTRATION") }()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/register", strings.NewReader("username=user01&password=pass"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, "404 page not found\n", string(res.Body.Bytes()))
	assert.Empty(t, res.Result().Cookies())

	dbMock.AssertExpectations(t)
}

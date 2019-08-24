package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zlogic/vogon-go/data"
)

func TestGetTagsAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	dbMock.On("GetUser", "user01").Return(user, nil).Once()

	req, _ := http.NewRequest("GET", "/api/tags", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(&cookie, "user01")
	req.AddCookie(&cookie)

	dbMock.On("GetTags", user).Return([]string{"t2", "t1"}, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `["t1","t2"]`+"\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestGetTagsUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/tags", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestGetTagsUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	dbMock.On("GetUser", "user01").Return(data.User{}, nil).Once()

	req, _ := http.NewRequest("GET", "/api/tags", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(&cookie, "user01")
	req.AddCookie(&cookie)

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

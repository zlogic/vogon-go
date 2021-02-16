package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStaticResource(t *testing.T) {
	staticFileBytes, err := staticContent.ReadFile("static/style.css")
	assert.NoError(t, err)

	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}
	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/static/style.css", nil)
	res := newRecorder()
	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, staticFileBytes, res.Body.Bytes())

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestListingNotAllowed(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}
	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	for _, url := range []string{"/static", "/static/"} {
		req, _ := http.NewRequest("GET", url, nil)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)
		assert.Equal(t, "404 page not found\n", string(res.Body.Bytes()))
	}

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

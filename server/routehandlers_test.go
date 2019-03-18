package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zlogic/vogon-go/data"
)

const layoutTemplate = `{{ define "layout" }}User {{ .User }}
Name {{ .Name }}
Content {{ template "content" . }}{{ end }}`

func prepareLayoutTemplateTestFile(tempDir string) error {
	return prepareTestFile(path.Join(tempDir, "templates"), "layout.html", []byte(layoutTemplate))
}

func TestRootHandlerNotLoggedIn(t *testing.T) {
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/login", res.Header().Get("Location"))
}

func TestRootHandlerLoggedIn(t *testing.T) {
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/transactions", res.Header().Get("Location"))
}

func TestLogoutHandler(t *testing.T) {
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/logout", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/login", res.Header().Get("Location"))

	username := cookieHandler.GetUsername(res, req)
	assert.Equal(t, "", username)
}

func TestFaviconHandler(t *testing.T) {
	tempDir, recover, err := prepareTempDir()
	defer func() {
		if recover != nil {
			recover()
		}
	}()
	assert.NoError(t, err)
	faviconBytes := []byte("i am a favicon")
	err = prepareTestFile(path.Join(tempDir, "static"), "favicon.ico", faviconBytes)
	assert.NoError(t, err)

	router, err := CreateRouter(&Services{})
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/favicon.ico", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, faviconBytes, res.Body.Bytes())
}

func TestHtmlLoginHandlerNotLoggedIn(t *testing.T) {
	loginTemplate := []byte(`{{ define "content" }}loginpage{{ end }}`)
	tempDir, recover, err := prepareTempDir()
	defer func() {
		if recover != nil {
			recover()
		}
	}()
	assert.NoError(t, err)
	err = prepareLayoutTemplateTestFile(tempDir)
	assert.NoError(t, err)
	err = prepareTestFile(path.Join(tempDir, "templates", "pages"), "login.html", []byte(loginTemplate))
	assert.NoError(t, err)

	router, err := CreateRouter(&Services{})
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/login", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User <nil>\nName \nContent loginpage", string(res.Body.Bytes()))
}

func TestHtmlLoginHandlerAlreadyLoggedIn(t *testing.T) {
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/login", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/transactions", res.Header().Get("Location"))
}

func TestHtmlRegisterHandlerNotLoggedIn(t *testing.T) {
	registerTemplate := []byte(`{{ define "content" }}registerpage{{ end }}`)
	tempDir, recover, err := prepareTempDir()
	defer func() {
		if recover != nil {
			recover()
		}
	}()
	assert.NoError(t, err)
	err = prepareLayoutTemplateTestFile(tempDir)
	assert.NoError(t, err)
	err = prepareTestFile(path.Join(tempDir, "templates", "pages"), "register.html", []byte(registerTemplate))
	assert.NoError(t, err)

	router, err := CreateRouter(&Services{})
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/register", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User <nil>\nName \nContent registerpage", string(res.Body.Bytes()))
}

func TestHtmlRegisterHandlerAlreadyLoggedIn(t *testing.T) {
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/register", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/transactions", res.Header().Get("Location"))
}

func TestHtmlRegisterHandlerRegistrationNotAllowed(t *testing.T) {
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	err = os.Setenv("ALLOW_REGISTRATION", "false")
	defer func() { os.Unsetenv("ALLOW_REGISTRATION") }()
	assert.NoError(t, err)

	services := &Services{cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/register", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	router.ServeHTTP(res, req)
	assert.Equal(t, "404 page not found\n", string(res.Body.Bytes()))
	assert.Empty(t, res.Result().Cookies())
}

func TestTransactionsHandlerLoggedIn(t *testing.T) {
	transactionsTemplate := []byte(`{{ define "content" }}transactionspage{{ end }}`)
	tempDir, recover, err := prepareTempDir()
	defer func() {
		if recover != nil {
			recover()
		}
	}()
	assert.NoError(t, err)
	err = prepareLayoutTemplateTestFile(tempDir)
	assert.NoError(t, err)
	err = prepareTestFile(path.Join(tempDir, "templates", "pages"), "transactions.html", []byte(transactionsTemplate))
	assert.NoError(t, err)

	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	dbMock := new(DBMock)
	dbMock.On("GetUser", "user01").Return(&data.User{}, nil).Once()

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/transactions", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User { 0 }\nName transactions\nContent transactionspage", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestHtmlTransactionsHandlerNotLoggedIn(t *testing.T) {
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/transactions", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/login", res.Header().Get("Location"))
}

func TestAccountsHandlerLoggedIn(t *testing.T) {
	transactionsTemplate := []byte(`{{ define "content" }}accountspage{{ end }}`)
	tempDir, recover, err := prepareTempDir()
	defer func() {
		if recover != nil {
			recover()
		}
	}()
	assert.NoError(t, err)
	err = prepareLayoutTemplateTestFile(tempDir)
	assert.NoError(t, err)
	err = prepareTestFile(path.Join(tempDir, "templates", "pages"), "accounts.html", []byte(transactionsTemplate))
	assert.NoError(t, err)

	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	dbMock := new(DBMock)
	dbMock.On("GetUser", "user01").Return(&data.User{}, nil).Once()

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/accounts", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User { 0 }\nName accounts\nContent accountspage", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestHtmlAccountsHandlerNotLoggedIn(t *testing.T) {
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/accounts", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/login", res.Header().Get("Location"))
}

func TestHtmlSettingsHandlerLoggedIn(t *testing.T) {
	settingsTemplate := []byte(`{{ define "content" }}settingspage{{ end }}`)
	tempDir, recover, err := prepareTempDir()
	defer func() {
		if recover != nil {
			recover()
		}
	}()
	assert.NoError(t, err)
	err = prepareLayoutTemplateTestFile(tempDir)
	assert.NoError(t, err)
	err = prepareTestFile(path.Join(tempDir, "templates", "pages"), "settings.html", []byte(settingsTemplate))
	assert.NoError(t, err)

	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	dbMock := new(DBMock)
	dbMock.On("GetUser", "user01").Return(&data.User{}, nil).Once()

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/settings", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User { 0 }\nName settings\nContent settingspage", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestHtmlSettingsHandlerNotLoggedIn(t *testing.T) {
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/settings", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/login", res.Header().Get("Location"))
}

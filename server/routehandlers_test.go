package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/zlogic/vogon-go/data"
)

const layoutTemplate = `{{ define "layout" }}User {{ .User }}
Name {{ .Name }}
Content {{ template "content" . }}{{ end }}`

func prepareTemplate(pageName, tmpl string) fs.FS {
	files := fstest.MapFS{
		"layout.html":                 &fstest.MapFile{Data: []byte(layoutTemplate)},
		"pages/" + pageName + ".html": &fstest.MapFile{Data: []byte(tmpl)},
	}
	return files
}

func TestRootHandlerNotLoggedIn(t *testing.T) {
	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	authHandler.On("HasAuthenticationCookie", mock.Anything).
		Return(false).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/login", res.Header().Get("Location"))

	authHandler.AssertExpectations(t)
}

func TestRootHandlerLoggedIn(t *testing.T) {
	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	authHandler.On("HasAuthenticationCookie", mock.Anything).
		Return(true).Once()
	authHandler.AllowUser(&data.User{})

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/transactions", res.Header().Get("Location"))

	authHandler.AssertExpectations(t)
}

func TestLogoutHandler(t *testing.T) {
	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/logout", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/login", res.Header().Get("Location"))

	assert.Empty(t, res.Result().Cookies())

	authHandler.AssertExpectations(t)
}

func TestFaviconHandler(t *testing.T) {
	faviconBytes, err := staticContent.ReadFile(faviconFilename)
	assert.NoError(t, err)

	authHandler := AuthHandlerMock{}

	router, err := CreateRouter(&Services{cookieHandler: &authHandler})
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/favicon.ico", nil)
	res := newRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, faviconBytes, res.Body.Bytes())

	authHandler.AssertExpectations(t)
}

func TestHtmlLoginHandlerNotLoggedIn(t *testing.T) {
	templates := prepareTemplate("login", `{{ define "content" }}loginpage{{ end }}`)

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler, templates: templates}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/login", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User <nil>\nName \nContent loginpage", string(res.Body.Bytes()))

	authHandler.AssertExpectations(t)
}

func TestHtmlLoginHandlerAlreadyLoggedIn(t *testing.T) {
	templates := prepareTemplate("login", `{{ define "content" }}loginpage{{ end }}`)

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler, templates: templates}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/login", nil)
	res := httptest.NewRecorder()

	authHandler.AllowUser(&data.User{})

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User <nil>\nName \nContent loginpage", string(res.Body.Bytes()))

	authHandler.AssertExpectations(t)
}

func TestHtmlRegisterHandlerNotLoggedIn(t *testing.T) {
	templates := prepareTemplate("register", `{{ define "content" }}registerpage{{ end }}`)

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler, templates: templates}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/register", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User <nil>\nName \nContent registerpage", string(res.Body.Bytes()))

	authHandler.AssertExpectations(t)
}

func TestHtmlRegisterHandlerAlreadyLoggedIn(t *testing.T) {
	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/register", nil)
	res := httptest.NewRecorder()

	authHandler.AllowUser(&data.User{})

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/transactions", res.Header().Get("Location"))

	authHandler.AssertExpectations(t)
}

func TestHtmlRegisterHandlerRegistrationNotAllowed(t *testing.T) {
	authHandler := AuthHandlerMock{}

	err := os.Setenv("ALLOW_REGISTRATION", "false")
	defer func() { os.Unsetenv("ALLOW_REGISTRATION") }()
	assert.NoError(t, err)

	services := &Services{cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/register", nil)
	res := httptest.NewRecorder()

	authHandler.AllowUser(&data.User{})

	router.ServeHTTP(res, req)
	assert.Equal(t, "404 page not found\n", string(res.Body.Bytes()))
	assert.Empty(t, res.Result().Cookies())

	authHandler.AssertExpectations(t)
}

func TestHtmlTransactionsHandlerLoggedIn(t *testing.T) {
	templates := prepareTemplate("transactions", `{{ define "content" }}transactionspage{{ end }}`)

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler, templates: templates}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/transactions", nil)
	res := httptest.NewRecorder()

	authHandler.AllowUser(&data.User{})

	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User {  0 }\nName transactions\nContent transactionspage", string(res.Body.Bytes()))

	authHandler.AssertExpectations(t)
}

func TestHtmlTransactionsHandlerNotLoggedIn(t *testing.T) {
	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/transactions", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/login", res.Header().Get("Location"))

	authHandler.AssertExpectations(t)
}

func TestHtmlTransactionEditorHandlerLoggedInEmptyValues(t *testing.T) {
	templates := prepareTemplate("transactioneditor", `{{ define "content" }}transactioneditor{{ if .Form }}values{{ end }}{{ end }}`)

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler, templates: templates}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/transactioneditor", nil)
	res := httptest.NewRecorder()

	authHandler.AllowUser(&data.User{})

	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User {  0 }\nName transactioneditor\nContent transactioneditor", string(res.Body.Bytes()))

	authHandler.AssertExpectations(t)
}

func TestHtmlTransactionEditorHandlerLoggedInWithValues(t *testing.T) {
	templates := prepareTemplate("transactioneditor", `{{ define "content" }}transactioneditor{{ if .Form }} {{ index .Form "id" 0 }} {{ index .Form "action" 0 }}{{ end }}{{ end }}`)

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler, templates: templates}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/transactioneditor?id=1&action=duplicate", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	authHandler.AllowUser(&data.User{})

	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User {  0 }\nName transactioneditor\nContent transactioneditor 1 duplicate", string(res.Body.Bytes()))

	authHandler.AssertExpectations(t)
}
func TestHtmlTransactionEditorHandlerNotLoggedIn(t *testing.T) {
	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/transactioneditor", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/login", res.Header().Get("Location"))

	authHandler.AssertExpectations(t)
}

func TestHtmlReportHandlerLoggedIn(t *testing.T) {
	templates := prepareTemplate("report", `{{ define "content" }}report {{ index .Form "filterDescription" 0 }} {{ index .Form "filterAccounts" 0 }}{{ end }}`)

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler, templates: templates}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/report", strings.NewReader("filterDescription=test&filterAccounts=1,2"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	authHandler.AllowUser(&data.User{})

	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User {  0 }\nName report\nContent report test 1,2", string(res.Body.Bytes()))

	authHandler.AssertExpectations(t)
}

func TestHtmlReportHandlerNotLoggedIn(t *testing.T) {
	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/report", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/login", res.Header().Get("Location"))

	authHandler.AssertExpectations(t)
}

func TestHtmlAccountsHandlerLoggedIn(t *testing.T) {
	templates := prepareTemplate("accounts", `{{ define "content" }}accountspage{{ end }}`)

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler, templates: templates}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/accounts", nil)
	res := httptest.NewRecorder()

	authHandler.AllowUser(&data.User{})

	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User {  0 }\nName accounts\nContent accountspage", string(res.Body.Bytes()))

	authHandler.AssertExpectations(t)
}

func TestHtmlAccountsHandlerNotLoggedIn(t *testing.T) {
	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/accounts", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/login", res.Header().Get("Location"))

	authHandler.AssertExpectations(t)
}

func TestHtmlSettingsHandlerLoggedIn(t *testing.T) {
	templates := prepareTemplate("settings", `{{ define "content" }}settingspage{{ end }}`)

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler, templates: templates}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/settings", nil)
	res := httptest.NewRecorder()

	authHandler.AllowUser(&data.User{})

	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "User {  0 }\nName settings\nContent settingspage", string(res.Body.Bytes()))

	authHandler.AssertExpectations(t)
}

func TestHtmlSettingsHandlerNotLoggedIn(t *testing.T) {
	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/settings", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusSeeOther, res.Code)
	assert.Equal(t, "/login", res.Header().Get("Location"))

	authHandler.AssertExpectations(t)
}

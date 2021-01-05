package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/zlogic/vogon-go/data"
)

const layoutTemplate = `{{ define "layout" }}User {{ .User }}
Name {{ .Name }}
Content {{ template "content" . }}{{ end }}`

func prepareLayoutTemplateTestFile(tempDir string) error {
	return prepareTestFile(path.Join(tempDir, "templates"), "layout.html", []byte(layoutTemplate))
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

	authHandler := AuthHandlerMock{}

	router, err := CreateRouter(&Services{cookieHandler: &authHandler})
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/favicon.ico", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, faviconBytes, res.Body.Bytes())

	authHandler.AssertExpectations(t)
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

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
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

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
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

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
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

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
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
	transactionEditorTemplate := []byte(`{{ define "content" }}transactioneditor{{ if .Form }}values{{ end }}{{ end }}`)
	tempDir, recover, err := prepareTempDir()
	defer func() {
		if recover != nil {
			recover()
		}
	}()
	assert.NoError(t, err)
	err = prepareLayoutTemplateTestFile(tempDir)
	assert.NoError(t, err)
	err = prepareTestFile(path.Join(tempDir, "templates", "pages"), "transactioneditor.html", []byte(transactionEditorTemplate))
	assert.NoError(t, err)

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
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
	transactionEditorTemplate := []byte(`{{ define "content" }}transactioneditor{{ if .Form }} {{ index .Form "id" 0 }} {{ index .Form "action" 0 }}{{ end }}{{ end }}`)
	tempDir, recover, err := prepareTempDir()
	defer func() {
		if recover != nil {
			recover()
		}
	}()
	assert.NoError(t, err)
	err = prepareLayoutTemplateTestFile(tempDir)
	assert.NoError(t, err)
	err = prepareTestFile(path.Join(tempDir, "templates", "pages"), "transactioneditor.html", []byte(transactionEditorTemplate))
	assert.NoError(t, err)

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
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
	reportTemplate := []byte(`{{ define "content" }}report {{ index .Form "filterDescription" 0 }} {{ index .Form "filterAccounts" 0 }}{{ end }}`)
	tempDir, recover, err := prepareTempDir()
	defer func() {
		if recover != nil {
			recover()
		}
	}()
	assert.NoError(t, err)
	err = prepareLayoutTemplateTestFile(tempDir)
	assert.NoError(t, err)
	err = prepareTestFile(path.Join(tempDir, "templates", "pages"), "report.html", []byte(reportTemplate))
	assert.NoError(t, err)

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
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

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
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

	authHandler := AuthHandlerMock{}

	services := &Services{cookieHandler: &authHandler}
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

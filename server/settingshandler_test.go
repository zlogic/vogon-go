package server

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zlogic/vogon-go/data"
)

func TestGetSettingsAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	dbMock.On("GetUser", "user01").Return(&testUser, nil).Once()

	req, _ := http.NewRequest("GET", "/api/settings", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"Username":"user01"}`+"\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestGetSettingsNotAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/settings", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestGetSettingsUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	dbMock.On("GetUser", "user01").Return(nil, nil).Once()

	req, _ := http.NewRequest("GET", "/api/settings", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestSaveSettingsNoChangesAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := prepareExistingUser("user01")
	assert.NotNil(t, user)
	user.SetPassword("pass")
	dbMock.On("GetUser", "user01").Return(user, nil).Twice()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("form", "Username=user01")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/settings", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	dbMock.On("SaveUser", user).Return(nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"Username":"user01"}`+"\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestSaveSettingsChangePasswordAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := prepareExistingUser("user01")
	assert.NotNil(t, user)
	user.SetPassword("pass")
	dbMock.On("GetUser", "user01").Return(user, nil).Twice()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("form", "Username=user01&Password=newpass")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/settings", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	dbMock.On("SaveUser", mock.AnythingOfType("*data.User")).Return(nil).Once().
		Run(func(args mock.Arguments) {
			saveUser := args.Get(0).(*data.User)
			assert.NoError(t, saveUser.ValidatePassword("newpass"))
		})

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"Username":"user01"}`+"\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestSaveSettingsChangeUsernameAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	user.SetPassword("pass")
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("form", "Username=user02")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/settings", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	saveUser := user
	saveUser.SetUsername("user02")
	getUpdatedUser := prepareExistingUser("user02")
	assert.NotNil(t, getUpdatedUser)
	dbMock.On("SaveUser", &saveUser).Return(nil).Once().
		Run(func(args mock.Arguments) {
			userArg := args.Get(0).(*data.User)
			*userArg = *getUpdatedUser
		})

	dbMock.On("GetUser", "user02").Return(&saveUser, nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"Username":"user02"}`+"\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestSaveSettingsChangeUsernameFailedAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	user.SetPassword("pass")
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("form", "Username=user02")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/settings", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	saveUser := user
	saveUser.SetUsername("user02")
	dbMock.On("SaveUser", &saveUser).Return(fmt.Errorf("Username already in use")).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "Internal server error\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestSaveSettingsRestoreBackupAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := prepareExistingUser("user01")
	assert.NotNil(t, user)
	user.SetPassword("pass")
	dbMock.On("GetUser", "user01").Return(user, nil).Twice()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("form", "Username=user01")
	fileWriter, err := writer.CreateFormFile("restorefile", "backup.json")
	assert.NoError(t, err)
	_, err = fileWriter.Write([]byte("json backup"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/settings", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	saveUser := user
	dbMock.On("SaveUser", saveUser).Return(nil).Once()
	dbMock.On("Restore", user, "json backup").Return(nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"Username":"user01"}`+"\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestSaveSettingsUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("form", "Username=user01")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/settings", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestSaveSettingsUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	dbMock.On("GetUser", "user01").Return(nil, nil).Once()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("form", "Username=user01")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/settings", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestBackupAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := testUser
	dbMock.On("GetUser", "user01").Return(&user, nil).Once()

	req, _ := http.NewRequest("POST", "/api/backup", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	dbMock.On("Backup", &user).Return("json backup", nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "json backup", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestBackupUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/backup", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

func TestBackupUserDoesNotExist(t *testing.T) {
	dbMock := new(DBMock)
	cookieHandler, err := createTestCookieHandler()
	assert.NoError(t, err)

	services := &Services{db: dbMock, cookieHandler: cookieHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	dbMock.On("GetUser", "user01").Return(nil, nil).Once()

	req, _ := http.NewRequest("POST", "/api/backup", nil)
	res := httptest.NewRecorder()

	cookie := cookieHandler.NewCookie()
	cookieHandler.SetCookieUsername(cookie, "user01")
	req.AddCookie(cookie)

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
}

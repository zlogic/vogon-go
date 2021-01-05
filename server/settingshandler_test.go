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
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/settings", nil)
	res := httptest.NewRecorder()

	user := prepareExistingUser("user01")
	assert.NotNil(t, user)
	authHandler.AllowUser(user)

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"Username":"user01"}`+"\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestGetSettingsUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/settings", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestSaveSettingsNoChangesAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := prepareExistingUser("user01")
	assert.NotNil(t, user)
	user.SetPassword("pass")
	dbMock.On("GetUser", "user01").Return(user, nil).Once()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("form", "Username=user01")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/settings", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()

	authHandler.AllowUser(user)

	dbMock.On("SaveUser", user).Return(nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"Username":"user01"}`+"\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestSaveSettingsChangePasswordAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := prepareExistingUser("user01")
	assert.NotNil(t, user)
	user.SetPassword("pass")
	dbMock.On("GetUser", "user01").Return(user, nil).Once()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("form", "Username=user01&Password=newpass")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/settings", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()

	authHandler.AllowUser(user)

	dbMock.On("SaveUser", mock.AnythingOfType("*data.User")).Return(nil).Once().
		Run(func(args mock.Arguments) {
			saveUser := args.Get(0).(*data.User)
			assert.NoError(t, saveUser.ValidatePassword("newpass"))
		})

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"Username":"user01"}`+"\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestSaveSettingsChangeUsernameAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("form", "Username=user02")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/settings", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()

	user := prepareExistingUser("user01")
	user.SetPassword("pass")
	authHandler.AllowUser(user)

	getUpdatedUser := prepareExistingUser("user02")
	getUpdatedUser.Password = user.Password
	assert.NotNil(t, getUpdatedUser)
	dbMock.On("SaveUser", user).Run(func(args mock.Arguments) {
		userArg := args.Get(0).(*data.User)
		*userArg = *getUpdatedUser
	}).Once().Return(nil).Once()

	dbMock.On("GetUser", "user02").Return(getUpdatedUser, nil)

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"Username":"user02"}`+"\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestSaveSettingsChangeUsernameFailedAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("form", "Username=user02")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/settings", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()

	user := prepareExistingUser("user02")
	authHandler.AllowUser(user)

	saveUser := *user
	saveUser.SetUsername("user02")
	dbMock.On("SaveUser", &saveUser).Return(fmt.Errorf("Username already in use")).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "Internal server error\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestSaveSettingsRestoreBackupAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	user := prepareExistingUser("user01")
	assert.NotNil(t, user)
	user.SetPassword("pass")
	dbMock.On("GetUser", "user01").Return(user, nil).Once()

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

	authHandler.AllowUser(user)

	saveUser := user
	dbMock.On("SaveUser", saveUser).Return(nil).Once()
	dbMock.On("Restore", user, "json backup").Return(nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"Username":"user01"}`+"\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestSaveSettingsUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
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
	authHandler.AssertExpectations(t)
}

func TestBackupAuthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/backup", nil)
	res := httptest.NewRecorder()

	user := testUser
	authHandler.AllowUser(&user)

	dbMock.On("Backup", &user).Return("json backup", nil).Once()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "json backup", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

func TestBackupUnauthorized(t *testing.T) {
	dbMock := new(DBMock)
	authHandler := AuthHandlerMock{}

	services := &Services{db: dbMock, cookieHandler: &authHandler}
	router, err := CreateRouter(services)
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/backup", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, "Bad credentials\n", string(res.Body.Bytes()))

	dbMock.AssertExpectations(t)
	authHandler.AssertExpectations(t)
}

package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	user1 := NewUser("user01")
	assert.Equal(t, &User{newUsername: "user01"}, user1)

	user2 := NewUser("user02")
	assert.Equal(t, &User{newUsername: "user02"}, user2)

	err = dbService.SaveUser(user1)
	assert.NoError(t, err)
	assert.Equal(t, &User{username: "user01", ID: 0}, user1)

	err = dbService.SaveUser(user2)
	assert.NoError(t, err)
	assert.Equal(t, &User{username: "user02", ID: 1}, user2)
}

func TestGetUserEmpty(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	user, err := dbService.GetUser("")
	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestCreateGetUser(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	user := &User{Password: "password"}
	err = user.SetUsername("user01")
	assert.NoError(t, err)
	err = dbService.SaveUser(user)
	assert.NoError(t, err)

	user, err = dbService.GetUser("user01")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "password", user.Password)
}

func TestSaveExistingUser(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	user := &User{Password: "password"}
	err = user.SetUsername("user01")
	assert.NoError(t, err)
	err = dbService.SaveUser(user)
	assert.NoError(t, err)

	user.Password = "newPassword"
	err = dbService.SaveUser(user)
	assert.NoError(t, err)

	user, err = dbService.GetUser("user01")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "newPassword", user.Password)
}

func TestSaveUsernameAlreadyInUse(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	user1 := &User{Password: "password"}
	err = user1.SetUsername("user01")
	assert.NoError(t, err)
	err = dbService.SaveUser(user1)
	assert.NoError(t, err)

	user2 := &User{Password: "newPassword"}
	err = user2.SetUsername("user01")
	assert.NoError(t, err)
	err = dbService.SaveUser(user2)
	assert.Equal(t, ErrUserAlreadyExists, err)

	dbUser, err := dbService.GetUser("user01")
	assert.NoError(t, err)
	assert.Equal(t, user1, dbUser)
}

func TestSetUserPassword(t *testing.T) {
	user := &User{}
	err := user.SetPassword("hello")
	assert.NoError(t, err)
	assert.NotNil(t, user.Password)
	assert.NotEqual(t, "password", user.Password)

	err = user.ValidatePassword("hello")
	assert.NoError(t, err)

	err = user.ValidatePassword("hellow")
	assert.Error(t, err)
}

func TestSetUsername(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	user := User{Password: "pass1"}
	err = user.SetUsername("user01")
	assert.NoError(t, err)
	users := []*User{&user}
	err = dbService.SaveUser(&user)
	assert.NoError(t, err)

	err = user.SetUsername("user02")
	assert.NoError(t, err)
	err = dbService.SaveUser(&user)
	assert.NoError(t, err)
	assert.Equal(t, "user02", user.username)

	dbUser, err := dbService.GetUser(user.username)
	assert.Equal(t, user, *dbUser)

	dbUsers, err := getAllUsers(dbService)
	assert.NoError(t, err)
	assert.EqualValues(t, users, dbUsers)
}

func TestSaveUsernameAndOtherFields(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	user := &User{Password: "password"}
	err = user.SetUsername("user01")
	assert.NoError(t, err)
	err = dbService.SaveUser(user)
	assert.NoError(t, err)

	user.Password = "newPassword"
	err = user.SetUsername("user02")
	assert.NoError(t, err)
	err = dbService.SaveUser(user)

	dbUser, err := dbService.GetUser("user1")
	assert.NoError(t, err)
	assert.Nil(t, dbUser)

	dbUser, err = dbService.GetUser("user02")
	assert.NoError(t, err)
	assert.Equal(t, user, dbUser)
}
func TestSetUsernameAlreadyExists(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	user1 := User{Password: "pass1"}
	err = user1.SetUsername("user01")
	assert.NoError(t, err)
	user2 := User{Password: "pass2"}
	err = user2.SetUsername("user02")
	assert.NoError(t, err)
	users := []*User{&user1, &user2}
	err = dbService.SaveUser(&user1)
	assert.NoError(t, err)
	err = dbService.SaveUser(&user2)
	assert.NoError(t, err)

	err = user1.SetUsername("user02")
	assert.NoError(t, err)
	err = dbService.SaveUser(&user1)
	assert.Error(t, err)
	assert.Equal(t, "user01", user1.username)
	assert.Equal(t, "user02", user1.newUsername)
	user1.newUsername = ""

	dbUser1, err := dbService.GetUser("user01")
	assert.Equal(t, user1, *dbUser1)

	dbUser2, err := dbService.GetUser("user02")
	assert.Equal(t, user2, *dbUser2)

	dbUsers, err := getAllUsers(dbService)
	assert.NoError(t, err)
	assert.EqualValues(t, users, dbUsers)
}

func TestSetUsernameEmptyString(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	user := User{Password: "pass1"}
	err = user.SetUsername("user01")
	assert.NoError(t, err)
	users := []*User{&user}
	err = dbService.SaveUser(&user)
	assert.NoError(t, err)

	err = user.SetUsername("  ")
	assert.Error(t, err)
	err = dbService.SaveUser(&user)
	assert.NoError(t, err)

	dbUser, err := dbService.GetUser(user.username)
	assert.Equal(t, user, *dbUser)

	dbUsers, err := getAllUsers(dbService)
	assert.NoError(t, err)
	assert.EqualValues(t, users, dbUsers)
}

func TestSaveUserIdConflict(t *testing.T) {
	err := resetDb()
	assert.NoError(t, err)

	user := User{
		Password: "pass1",
		username: "user01",
		ID:       1,
	}
	users := []*User{&user}
	saveUser := user
	err = dbService.SaveUser(&saveUser)
	assert.NoError(t, err)
	saveUser = user
	saveUser.ID = 2
	err = dbService.SaveUser(&saveUser)
	assert.Error(t, err)

	dbUsers, err := getAllUsers(dbService)
	assert.NoError(t, err)
	assert.EqualValues(t, users, dbUsers)
}

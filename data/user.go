package data

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User keeps configuration for a user and information used to link a user with their data.
type User struct {
	username    string
	newUsername string
	UUID        string
	Password    string
}

// ErrUserAlreadyExists is an error when a user cannot be renamed because their username is already in use.
var ErrUserAlreadyExists = fmt.Errorf("uuid conflicts with existing user")

// NewUser creates a User with the provided username and a generated UUID.
func NewUser(username string) *User {
	return &User{newUsername: username}
}

// decode deserializes a User.
func (user *User) decode(val []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(val)).Decode(user)
}

// GetUser returns the User by username.
// If user doesn't exist, returns nil.
func (s *DBService) GetUser(username string) (*User, error) {
	user := &User{username: username}
	err := s.view(func() error {
		value, err := s.db.Get(user.createKey())
		if err != nil {
			return err
		}
		if value == nil {
			user = nil
			return nil
		}

		return user.decode(value)
	})
	if err != nil {
		return nil, fmt.Errorf("cannot read user %v: %w", username, err)
	}
	return user, nil
}

// SaveUser saves updates an existing user in the database.
func (s *DBService) SaveUser(user *User) error {
	if user.newUsername == "" {
		user.newUsername = user.username
	}
	key := createUserKey(user.newUsername)

	newUser := user.username == ""
	if newUser {
		user.UUID = uuid.NewString()
	}

	err := s.update(func() error {
		// Check for username/id conflicts.
		existingUserValue, err := s.db.Get(key)
		if err != nil {
			return fmt.Errorf("cannot get existing value for user %v: %w", string(key), err)
		}

		existingUser := &User{}
		if existingUserValue != nil {
			if err := existingUser.decode(existingUserValue); err != nil {
				return fmt.Errorf("cannot decode existing value for user %v: %w", string(key), err)
			}
			if user.newUsername != user.username {
				return ErrUserAlreadyExists
			}
			if existingUser.UUID != user.UUID {
				return fmt.Errorf("id %v for user %v conflicts with existing user id %v: %w", user.UUID, string(key), existingUser.UUID, ErrUserAlreadyExists)
			}
		}

		// In case of rename, delete old username key
		if !newUser && user.newUsername != user.username {
			oldUserKey := createUserKey(user.username)
			if err := s.db.Delete(oldUserKey); err != nil {
				return err
			}
		}

		var value bytes.Buffer
		if err := gob.NewEncoder(&value).Encode(user); err != nil {
			return fmt.Errorf("cannot encode user: %w", err)
		}
		return s.db.Put(key, value.Bytes())
	})

	if err == nil {
		user.username = user.newUsername
		user.newUsername = ""
	}
	return err
}

// GetUsername returns the user's current username.
func (user *User) GetUsername() string {
	return user.username
}

// SetUsername sets a new username for User which will be updated when SaveUser is called.
func (user *User) SetUsername(newUsername string) error {
	newUsername = strings.TrimSpace(newUsername)
	if newUsername == "" {
		return fmt.Errorf("cannot set username to an empty string")
	}
	user.newUsername = newUsername
	return nil
}

// SetPassword sets a new password for user. The password is hashed and salted with bcrypt.
func (user *User) SetPassword(newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return nil
}

// ValidatePassword checks if password matches the user's password.
func (user *User) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
}

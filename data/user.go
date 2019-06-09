package data

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// User keeps configuration for a user and information used to link a user with their data.
type User struct {
	username    string
	newUsername string
	ID          uint64
	Password    string
}

// ErrUserAlreadyExists is an error when a user cannot be renamed because their username is already in use.
var ErrUserAlreadyExists = errors.New("id conflicts with existing user")

// NewUser creates a User with the provided username and a generated ID.
func NewUser(username string) *User {
	return &User{newUsername: username}
}

// GetUser returns the User by username.
// If user doesn't exist, returns nil.
func (s *DBService) GetUser(username string) (*User, error) {
	user := &User{username: username}
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(user.CreateKey())
		if err == badger.ErrKeyNotFound {
			user = nil
			return nil
		}

		value, err := item.Value()
		if err != nil {
			return err
		}
		err = gob.NewDecoder(bytes.NewBuffer(value)).Decode(&user)
		if err != nil {
			user = nil
		}
		return err
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot read User %v", username)
	}
	return user, nil
}

// SaveUser saves updates an existing user in the database.
func (s *DBService) SaveUser(user *User) error {
	if user.newUsername == "" {
		user.newUsername = user.username
	}
	key := CreateUserKey(user.newUsername)

	newUser := user.username == ""
	if newUser {
		seq, err := s.db.GetSequence([]byte(SequenceUserKey), 1)
		defer seq.Release()
		if err != nil {
			return errors.Wrap(err, "Cannot create user sequence object")
		}
		id, err := seq.Next()
		if err != nil {
			return errors.Wrap(err, "Cannot generate id for user")
		}
		user.ID = id
	}

	err := s.db.Update(func(txn *badger.Txn) error {
		// Check for username/id conflicts
		_, err := txn.Get(key)
		existingItem, err := txn.Get(key)

		if newUser && err != badger.ErrKeyNotFound {
			return ErrUserAlreadyExists
		}

		existingUser := &User{}
		if existingItem != nil || (err != nil && err != badger.ErrKeyNotFound) {
			value, err := existingItem.Value()
			if err != nil {
				return errors.Wrap(err, "Cannot get existing value")
			}
			if err := gob.NewDecoder(bytes.NewBuffer(value)).Decode(existingUser); err != nil {
				return errors.Wrap(err, "Cannot unmarshal user")
			}
			if user.newUsername != user.username {
				log.WithField("key", key).Error("New username already in use")
				return ErrUserAlreadyExists
			}
			if existingUser.ID != user.ID {
				log.WithField("key", key).WithField("existingID", existingUser.ID).WithField("ID", user.ID).Error("ID conflict with existing user")
				return ErrUserAlreadyExists
			}
		}

		// In case of rename, delete old username key
		if !newUser && user.newUsername != user.username {
			oldUserKey := CreateUserKey(user.username)
			if err := txn.Delete(oldUserKey); err != nil {
				return err
			}
		}

		var value bytes.Buffer
		if err := gob.NewEncoder(&value).Encode(user); err != nil {
			return errors.Wrap(err, "Cannot marshal user")
		}
		return txn.Set(key, value.Bytes())
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
		return fmt.Errorf("Cannot set username to an empty string")
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

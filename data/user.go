package data

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// User keeps configuration for a user and information used to link a user with their data.
type User struct {
	username string
	ID       uint64
	Password string
}

// ErrUserAlreadyExists is an error when a user cannot be renamed because their username is already in use.
var ErrUserAlreadyExists = errors.New("id conflicts with existing user")

// CreateUser creates a User with the provided username and a generated ID.
func (s *DBService) CreateUser(username string) (*User, error) {
	seq, err := s.db.GetSequence([]byte(SequenceUserKey), 1)
	defer seq.Release()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot create user sequence object")
	}
	id, err := seq.Next()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot generate id for user")
	}
	return &User{username: username, ID: id}, nil
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
func (s *DBService) SaveUser(user *User) (err error) {
	return s.saveUser(user, false)
}

// SaveNewUser saves saves a new user into the database.
func (s *DBService) SaveNewUser(user *User) (err error) {
	return s.saveUser(user, true)
}

func (s *DBService) saveUser(user *User, newUser bool) (err error) {
	key := user.CreateKey()

	return s.db.Update(func(txn *badger.Txn) error {
		// Check for username/id conflicts
		_, err := txn.Get(key)
		existingItem, err := txn.Get(key)

		if newUser && err != badger.ErrKeyNotFound {
			return ErrUserAlreadyExists
		}

		if existingItem != nil || (err != nil && err != badger.ErrKeyNotFound) {
			value, err := existingItem.Value()
			if err != nil {
				return errors.Wrap(err, "Cannot get existing value")
			}
			existingUser := &User{}
			if err := gob.NewDecoder(bytes.NewBuffer(value)).Decode(existingUser); err != nil {
				return errors.Wrap(err, "Cannot unmarshal user")
			}
			if existingUser.ID != user.ID {
				return fmt.Errorf("ID conflict with existing user %v", existingUser)
			}
		}
		var value bytes.Buffer
		if err := gob.NewEncoder(&value).Encode(user); err != nil {
			return errors.Wrap(err, "Cannot marshal user")
		}
		return txn.Set(key, value.Bytes())
	})
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

// SetUsername sets a new username for user.
// If newUsername is already in use, returns an error.
func (s *DBService) SetUsername(user *User, newUsername string) error {
	newUsername = strings.TrimSpace(newUsername)
	if newUsername == "" {
		return fmt.Errorf("Cannot set username to an empty string")
	}
	newUser := *user
	newUser.username = newUsername
	err := s.db.Update(func(txn *badger.Txn) error {
		oldUserKey := user.CreateKey()
		item, err := txn.Get(oldUserKey)
		if err != nil {
			return err
		}
		value, err := item.Value()
		if err != nil {
			return err
		}
		newUserKey := newUser.CreateKey()
		existingUser, err := txn.Get(newUserKey)
		if existingUser != nil || (err != nil && err != badger.ErrKeyNotFound) {
			return fmt.Errorf("New username %v is already in use", newUsername)
		}
		err = txn.Set(newUserKey, value)
		if err != nil {
			return err
		}
		return txn.Delete(oldUserKey)
	})
	if err == nil {
		user.username = newUser.username
	}
	return err
}

// ValidatePassword checks if password matches the user's password.
func (user *User) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
}

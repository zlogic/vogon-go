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

type User struct {
	username string
	ID       uint64
	Password string
}

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

func (s *DBService) SaveUser(user *User) (err error) {
	key := user.CreateKey()

	return s.db.Update(func(txn *badger.Txn) error {
		// Check for username/id conflicts
		_, err := txn.Get(key)
		existingItem, err := txn.Get(key)
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

func (user *User) SetPassword(newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return nil
}

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

func (user *User) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
}

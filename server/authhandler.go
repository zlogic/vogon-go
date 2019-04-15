package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/zlogic/vogon-go/data"
)

func handleBadCredentials(w http.ResponseWriter, r *http.Request, err error) {
	log.WithError(err).Error("Bad credentials for user")
	http.Error(w, "Bad credentials", http.StatusUnauthorized)
}

func validateUserForAPI(w http.ResponseWriter, r *http.Request, s *Services) *data.User {
	username := s.cookieHandler.GetUsername(w, r)
	if username == "" {
		http.Error(w, "Bad credentials", http.StatusUnauthorized)
		return nil
	}

	user, err := s.db.GetUser(username)
	if err != nil {
		handleError(w, r, err)
		return nil
	}
	if user == nil {
		handleBadCredentials(w, r, fmt.Errorf("Unknown username %v", username))
	}
	return user
}

// LoginHandler authenticates the user and sets the encrypted session cookie if the user provided valid credentials.
func LoginHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			handleError(w, r, err)
			return
		}

		username := r.Form.Get("username")
		password := r.Form.Get("password")
		rememberMe, err := strconv.ParseBool(r.Form.Get("rememberMe"))
		if err != nil {
			log.WithError(err).Error("Failed to parse rememberMe parameter")
			rememberMe = false
		}

		user, err := s.db.GetUser(username)
		if err != nil {
			handleError(w, r, err)
			return
		}
		if user == nil {
			handleBadCredentials(w, r, fmt.Errorf("User %v does not exist", username))
			return
		}
		err = user.ValidatePassword(password)
		if err != nil {
			handleBadCredentials(w, r, errors.Wrapf(err, "Invalid password for user %v", username))
			return
		}
		cookie := s.cookieHandler.NewCookie()
		s.cookieHandler.SetCookieUsername(cookie, username)
		if !rememberMe {
			cookie.Expires = time.Time{}
			cookie.MaxAge = 0
		}
		http.SetCookie(w, cookie)
		_, err = io.WriteString(w, "OK")
		if err != nil {
			log.WithError(err).Error("Failed to write response")
		}
	}
}

// RegisterHandler creates and logs in a new user.
func RegisterHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			handleError(w, r, err)
			return
		}

		username := r.Form.Get("username")
		password := r.Form.Get("password")
		rememberMe, err := strconv.ParseBool(r.Form.Get("rememberMe"))
		if err != nil {
			log.WithError(err).Error("Failed to parse rememberMe parameter")
			rememberMe = false
		}

		user, err := s.db.CreateUser(username)
		if err != nil {
			handleError(w, r, err)
			return
		}
		if err := user.SetPassword(password); err != nil {
			handleError(w, r, err)
			return
		}
		if err := s.db.SaveNewUser(user); err != nil {
			if err == data.ErrUserAlreadyExists {
				http.Error(w, "Username is already in use", http.StatusInternalServerError)
			} else {
				handleError(w, r, err)
			}
			return
		}
		cookie := s.cookieHandler.NewCookie()
		s.cookieHandler.SetCookieUsername(cookie, username)
		if !rememberMe {
			cookie.Expires = time.Time{}
			cookie.MaxAge = 0
		}
		http.SetCookie(w, cookie)
		_, err = io.WriteString(w, "OK")
		if err != nil {
			log.WithError(err).Error("Failed to write response")
		}
	}
}

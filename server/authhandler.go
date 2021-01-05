package server

import (
	"io"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/zlogic/vogon-go/data"
	"github.com/zlogic/vogon-go/server/auth"
)

// APIAuthHandler checks to see if the API is accessed by an authorized user,
// and returns an error if the request is done by an unauthorized user.
func APIAuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil {
			http.Error(w, "Bad credentials", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
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
			log.Errorf("User %v doesn't exist", username)
			http.Error(w, "Bad credentials", http.StatusUnauthorized)
			return
		}
		err = user.ValidatePassword(password)
		if err != nil {
			log.WithError(err).Errorf("Invalid password for user %v", username)
			http.Error(w, "Bad credentials", http.StatusUnauthorized)
			return
		}
		err = s.cookieHandler.SetCookieUsername(w, username, rememberMe)
		if err != nil {
			log.WithError(err).Error("Failed to set username cookie")
			http.Error(w, "Failed to set username cookie", http.StatusInternalServerError)
			return
		}
		if _, err := io.WriteString(w, "OK"); err != nil {
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

		user := data.NewUser(username)
		if err := user.SetPassword(password); err != nil {
			handleError(w, r, err)
			return
		}
		if err := s.db.SaveUser(user); err != nil {
			if err == data.ErrUserAlreadyExists {
				http.Error(w, "Username is already in use", http.StatusInternalServerError)
			} else {
				handleError(w, r, err)
			}
			return
		}
		err = s.cookieHandler.SetCookieUsername(w, username, rememberMe)
		if err != nil {
			log.WithError(err).Error("Failed to set username cookie")
			http.Error(w, "Failed to set username cookie", http.StatusInternalServerError)
			return
		}
		if _, err := io.WriteString(w, "OK"); err != nil {
			log.WithError(err).Error("Failed to write response")
		}
	}
}

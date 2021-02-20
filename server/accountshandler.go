package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"

	"github.com/zlogic/vogon-go/data"
	"github.com/zlogic/vogon-go/server/auth"
)

// AccountsHandler returns all Accounts for an authenticated user.
func AccountsHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil {
			// This should never happen.
			return
		}

		accounts, err := s.db.GetAccounts(user)
		if err != nil {
			handleError(w, r, err)
			return
		}

		if err := json.NewEncoder(w).Encode(accounts); err != nil {
			handleError(w, r, err)
		}
	}
}

// AccountHandler gets, updates or deletes an Account.
func AccountHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil {
			// This should never happen.
			return
		}

		requestID := chi.URLParam(r, "id")

		if r.Method == http.MethodPost {
			account := &data.Account{}

			err := json.NewDecoder(r.Body).Decode(&account)
			if err != nil {
				handleError(w, r, err)
				return
			}

			if requestID == "new" {
				err = s.db.CreateAccount(user, account)
				requestID = strconv.FormatUint(account.ID, 10)
			} else {
				err = s.db.UpdateAccount(user, account)
			}
			if err != nil {
				handleError(w, r, err)
				return
			}

			if _, err := io.WriteString(w, "OK"); err != nil {
				log.WithError(err).Error("Failed to write response")
			}
			return
		}

		id, err := strconv.ParseUint(requestID, 10, 64)
		if err != nil {
			handleError(w, r, err)
			return
		}

		if r.Method == http.MethodDelete {
			if err := s.db.DeleteAccount(user, id); err != nil {
				handleError(w, r, err)
				return
			}

			if _, err := io.WriteString(w, "OK"); err != nil {
				log.WithError(err).Error("Failed to write response")
			}
			return
		}

		account, err := s.db.GetAccount(user, id)
		if err != nil {
			handleError(w, r, err)
			return
		}

		if err := json.NewEncoder(w).Encode(account); err != nil {
			handleError(w, r, err)
		}
	}
}

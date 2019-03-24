package server

import (
	"encoding/json"
	"net/http"
)

func AccountsHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := validateUserForAPI(w, r, s)
		if user == nil {
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

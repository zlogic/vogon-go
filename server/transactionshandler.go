package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/zlogic/vogon-go/data"
)

func TransactionsCountHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := validateUserForAPI(w, r, s)
		if user == nil {
			return
		}

		count, err := s.db.CountTransactions(user)
		if err != nil {
			handleError(w, r, err)
			return
		}

		if err := json.NewEncoder(w).Encode(count); err != nil {
			handleError(w, r, err)
		}
	}
}

func TransactionsHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := validateUserForAPI(w, r, s)
		if user == nil {
			return
		}

		queryValues := r.URL.Query()
		parseQueryValueInt := func(name string) (uint64, error) {
			value := queryValues.Get(name)
			if value == "" {
				return 0, fmt.Errorf("Query parameter %v is empty", name)
			}
			return strconv.ParseUint(value, 10, 64)
		}

		offset, err := parseQueryValueInt("offset")
		if err != nil {
			handleError(w, r, err)
			return
		}

		limit, err := parseQueryValueInt("limit")
		if err != nil {
			handleError(w, r, err)
			return
		}

		options := data.GetTransactionOptions{Offset: offset, Limit: limit}
		transactions, err := s.db.GetTransactions(user, options)
		if err != nil {
			handleError(w, r, err)
			return
		}

		if err := json.NewEncoder(w).Encode(transactions); err != nil {
			handleError(w, r, err)
		}
	}
}

func TransactionHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := validateUserForAPI(w, r, s)
		if user == nil {
			return
		}

		vars := mux.Vars(r)
		requestID := vars["id"]

		if r.Method == http.MethodPost {
			transaction := &data.Transaction{}

			err := json.NewDecoder(r.Body).Decode(&transaction)
			if err != nil {
				handleError(w, r, err)
				return
			}

			if requestID == "new" {
				err = s.db.CreateTransaction(user, transaction)
				requestID = strconv.FormatUint(transaction.ID, 10)
			} else {
				err = s.db.UpdateTransaction(user, transaction)
			}
			if err != nil {
				handleError(w, r, err)
				return
			}

			_, err = io.WriteString(w, "OK")
			if err != nil {
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
			if err := s.db.DeleteTransaction(user, id); err != nil {
				handleError(w, r, err)
				return
			}

			_, err = io.WriteString(w, "OK")
			if err != nil {
				log.WithError(err).Error("Failed to write response")
			}
			return
		}

		transaction, err := s.db.GetTransaction(user, id)
		if err != nil {
			handleError(w, r, err)
			return
		}

		if err := json.NewEncoder(w).Encode(transaction); err != nil {
			handleError(w, r, err)
		}
	}
}

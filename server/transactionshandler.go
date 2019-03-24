package server

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/zlogic/vogon-go/data"
)

func sortTransactions(transactions []*data.Transaction) {
	sort.Slice(transactions, func(i, j int) bool {
		if transactions[i].Date != transactions[j].Date {
			return transactions[i].Date > transactions[j].Date
		} else {
			return transactions[i].ID > transactions[j].ID
		}
	})
}

func TransactionsHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := validateUserForAPI(w, r, s)
		if user == nil {
			return
		}

		transactions, err := s.db.GetTransactions(user)
		if err != nil {
			handleError(w, r, err)
			return
		}

		sortTransactions(transactions)

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

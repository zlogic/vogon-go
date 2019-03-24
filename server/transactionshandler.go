package server

import (
	"encoding/json"
	"net/http"
	"sort"

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

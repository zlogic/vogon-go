package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"

	"github.com/zlogic/vogon-go/data"
	"github.com/zlogic/vogon-go/server/auth"
)

func parseFilterForm(r *http.Request) (data.TransactionFilterOptions, error) {
	parseFormValueSet := func(name string) []string {
		value := r.Form.Get(name)
		if value == "" {
			return nil
		}
		return strings.Split(value, ",")
	}

	parseFormValueBool := func(name string, defaultValue bool) (bool, error) {
		value := r.Form.Get(name)
		if value == "" {
			return defaultValue, nil
		}
		return strconv.ParseBool(value)
	}

	filterAccountUUIDs := parseFormValueSet("filterAccounts")

	includeExpenseIncome, err := parseFormValueBool("filterIncludeExpenseIncome", true)
	if err != nil {
		return data.TransactionFilterOptions{}, err
	}
	includeTransfer, err := parseFormValueBool("filterIncludeTransfer", true)
	if err != nil {
		return data.TransactionFilterOptions{}, err
	}

	return data.TransactionFilterOptions{
		FilterDescription:    r.Form.Get("filterDescription"),
		FilterFromDate:       r.Form.Get("filterFrom"),
		FilterToDate:         r.Form.Get("filterTo"),
		FilterTags:           parseFormValueSet("filterTags"),
		FilterAccounts:       filterAccountUUIDs,
		ExcludeExpenseIncome: !includeExpenseIncome,
		ExcludeTransfer:      !includeTransfer,
	}, nil
}

// TransactionsCountHandler returns the number of transactions for an authenticated user.
func TransactionsCountHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil {
			// This should never happen.
			return
		}

		if err := r.ParseForm(); err != nil {
			handleError(w, r, err)
			return
		}

		options, err := parseFilterForm(r)
		if err != nil {
			handleError(w, r, err)
			return
		}

		count, err := s.db.CountTransactions(user, options)
		if err != nil {
			handleError(w, r, err)
			return
		}

		if err := json.NewEncoder(w).Encode(count); err != nil {
			handleError(w, r, err)
		}
	}
}

// TransactionsHandler returns a filtered, pages list of transactions for an authenticated user.
func TransactionsHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil {
			// This should never happen.
			return
		}

		if err := r.ParseForm(); err != nil {
			handleError(w, r, err)
			return
		}

		parseFormValueInt := func(name string) (uint64, error) {
			value := r.Form.Get(name)
			if value == "" {
				return 0, fmt.Errorf("form parameter %v is empty", name)
			}
			return strconv.ParseUint(value, 10, 64)
		}

		offset, err := parseFormValueInt("offset")
		if err != nil {
			handleError(w, r, err)
			return
		}

		limit, err := parseFormValueInt("limit")
		if err != nil {
			handleError(w, r, err)
			return
		}

		filterOptions, err := parseFilterForm(r)
		if err != nil {
			handleError(w, r, err)
			return
		}
		options := data.GetTransactionOptions{
			Offset:                   offset,
			Limit:                    limit,
			TransactionFilterOptions: filterOptions,
		}
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

// TransactionHandler gets, updates or deletes a Transaction.
func TransactionHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil {
			// This should never happen.
			return
		}

		requestUUID := chi.URLParam(r, "uuid")

		if r.Method == http.MethodPost {
			transaction := &data.Transaction{}

			err := json.NewDecoder(r.Body).Decode(&transaction)
			if err != nil {
				handleError(w, r, err)
				return
			}

			if requestUUID == "new" {
				err = s.db.CreateTransaction(user, transaction)
			} else {
				err = s.db.UpdateTransaction(user, transaction)
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

		if r.Method == http.MethodDelete {
			if err := s.db.DeleteTransaction(user, requestUUID); err != nil {
				handleError(w, r, err)
				return
			}

			if _, err := io.WriteString(w, "OK"); err != nil {
				log.WithError(err).Error("Failed to write response")
			}
			return
		}

		transaction, err := s.db.GetTransaction(user, requestUUID)
		if err != nil {
			handleError(w, r, err)
			return
		}
		if transaction == nil {
			handleNotFound(w, r, requestUUID)
			return
		}

		if err := json.NewEncoder(w).Encode(transaction); err != nil {
			handleError(w, r, err)
		}
	}
}

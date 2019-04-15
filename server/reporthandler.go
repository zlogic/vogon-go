package server

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/zlogic/vogon-go/data"
)

func getCurrency(component data.TransactionComponent, accounts []*data.Account) string {
	for _, account := range accounts {
		if account.ID == component.AccountID {
			return account.Currency
		}
	}
	return ""
}

func filterComponents(components *[]data.TransactionComponent, accountIDs []uint64) {
	resultComponents := (*components)[:0]
	for _, component := range *components {
		for _, accountID := range accountIDs {
			if component.AccountID == accountID {
				resultComponents = append(resultComponents, component)
				break
			}
		}
	}
	*components = resultComponents
}

func filterTransactions(transactions *[]*data.Transaction, filterOptions data.TransactionFilterOptions) {
	if filterOptions.IsEmpty() {
		return
	}
	result := (*transactions)[:0]
	for _, transaction := range *transactions {
		if filterOptions.Matches(transaction) {
			filterComponents(&transaction.Components, filterOptions.FilterAccounts)
			result = append(result, transaction)
		}
	}
	*transactions = result
}

func filterAccounts(accounts []*data.Account, filterOptions data.TransactionFilterOptions) []*data.Account {
	if filterOptions.IsEmpty() {
		return accounts
	}
	filteredAccounts := make([]*data.Account, 0, len(filterOptions.FilterAccounts))
	for _, accountID := range filterOptions.FilterAccounts {
		for _, account := range accounts {
			if account.ID == accountID {
				filteredAccounts = append(filteredAccounts, account)
				break
			}
		}
	}
	return filteredAccounts
}

type dateBalance map[string]int64
type currencyDateBalance map[string]dateBalance

type tagAmount map[string]int64
type tagAmounts struct {
	Positive tagAmount
	Negative tagAmount
	Transfer tagAmount
}
type currencyTagAmount map[string]tagAmounts

func createBalanceChart(transactions []*data.Transaction, accounts []*data.Account, filterOptions data.TransactionFilterOptions) currencyDateBalance {
	var chart = make(currencyDateBalance)
	var totals = make(map[string]int64)
	var filterMatches = func(transaction *data.Transaction) bool {
		return (filterOptions.FilterFromDate == "" || filterOptions.FilterFromDate <= transaction.Date) &&
			(filterOptions.FilterToDate == "" || transaction.Date <= filterOptions.FilterToDate)
	}
	emptyFilter := filterOptions.IsEmpty()
	for i := range transactions {
		transaction := transactions[len(transactions)-1-i]
		for _, component := range transaction.Components {
			currency := getCurrency(component, accounts)
			if currency == "" {
				continue
			}
			totals[currency] = totals[currency] + component.Amount

			currencyBalance, ok := chart[currency]
			if !ok {
				currencyBalance = make(dateBalance)
				chart[currency] = currencyBalance
			}
			if emptyFilter || filterMatches(transaction) {
				currencyBalance[transaction.Date] = totals[currency]
			}
		}
	}
	return chart
}

func createTagsChart(transactions []*data.Transaction, accounts []*data.Account) currencyTagAmount {
	var chart = make(currencyTagAmount)
	type transactionTotal struct {
		Positive int64
		Negative int64
	}
	for _, transaction := range transactions {
		sort.Strings(transaction.Tags)
		tags := strings.Join(transaction.Tags, ",")
		transactionTotals := make(map[string]transactionTotal)
		for _, component := range transaction.Components {
			currency := getCurrency(component, accounts)

			currencyTotal := transactionTotals[currency]
			if component.Amount > 0 {
				currencyTotal.Positive = currencyTotal.Positive + component.Amount
			} else if component.Amount < 0 {
				currencyTotal.Negative = currencyTotal.Negative - component.Amount
			}
			transactionTotals[currency] = currencyTotal
		}

		for currency, totals := range transactionTotals {
			currencyAmounts, ok := chart[currency]
			if !ok {
				currencyAmounts = tagAmounts{Positive: make(tagAmount), Negative: make(tagAmount), Transfer: make(tagAmount)}
			}
			if transaction.Type == data.TransactionTypeExpenseIncome {
				currencyAmounts.Positive[tags] = currencyAmounts.Positive[tags] + totals.Positive
				currencyAmounts.Negative[tags] += totals.Negative
			} else if transaction.Type == data.TransactionTypeTransfer {
				var amount int64
				if totals.Positive >= totals.Negative {
					amount = totals.Positive
				} else {
					amount = totals.Negative
				}
				currencyAmounts.Transfer[tags] = currencyAmounts.Transfer[tags] + amount
			}
			chart[currency] = currencyAmounts
		}
	}
	return chart
}

// ReportHandler generates data for a report.
func ReportHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := validateUserForAPI(w, r, s)
		if user == nil {
			return
		}

		if err := r.ParseForm(); err != nil {
			handleError(w, r, err)
			return
		}

		filterOptions, err := parseFilterForm(r)
		if err != nil {
			handleError(w, r, err)
			return
		}
		options := data.GetAllTransactionsOptions
		transactions, err := s.db.GetTransactions(user, options)
		if err != nil {
			handleError(w, r, err)
			return
		}

		accounts, err := s.db.GetAccounts(user)
		if err != nil {
			handleError(w, r, err)
			return
		}
		accounts = filterAccounts(accounts, filterOptions)

		chart := createBalanceChart(transactions, accounts, filterOptions)

		filterTransactions(&transactions, filterOptions)
		tags := createTagsChart(transactions, accounts)
		type report struct {
			BalanceChart currencyDateBalance
			TagsChart    currencyTagAmount
		}

		if err := json.NewEncoder(w).Encode(report{BalanceChart: chart, TagsChart: tags}); err != nil {
			handleError(w, r, err)
		}
	}
}
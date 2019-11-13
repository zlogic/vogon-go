package data

import (
	"fmt"

	"github.com/dgraph-io/badger/v2"
)

// GetTags returns an unsorted (but deduplicated) list of tags for user.
func (s *DBService) GetTags(user *User) ([]string, error) {
	var transactions []*Transaction

	err := s.db.View(func(txn *badger.Txn) error {
		var err error
		transactions, err = s.getTransactions(user, GetAllTransactionsOptions)(txn)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to get transactions for tags because of %w", err)
	}

	tags := make(map[string]bool)
	for _, transaction := range transactions {
		for _, tag := range transaction.Tags {
			tags[tag] = true
		}
	}

	tagsList := make([]string, 0, len(tags))
	for tag := range tags {
		tagsList = append(tagsList, tag)
	}
	return tagsList, nil
}

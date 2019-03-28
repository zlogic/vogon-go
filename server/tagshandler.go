package server

import (
	"encoding/json"
	"net/http"
	"sort"
)

func TagsHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := validateUserForAPI(w, r, s)
		if user == nil {
			return
		}

		tags, err := s.db.GetTags(user)
		if err != nil {
			handleError(w, r, err)
			return
		}

		sort.Strings(tags)

		if err := json.NewEncoder(w).Encode(tags); err != nil {
			handleError(w, r, err)
		}
	}
}

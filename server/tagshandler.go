package server

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/zlogic/vogon-go/server/auth"
)

// TagsHandler returns a sorted, deduplicated list of tags for an authenticated user.
func TagsHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil {
			// This should never happen.
			return
		}

		tags, err := s.db.GetTags(user)
		if err != nil {
			handleError(w, r, err)
			return
		}

		sort.Strings(tags)

		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(tags); err != nil {
			handleError(w, r, err)
		}
	}
}

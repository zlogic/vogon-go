package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/zlogic/vogon-go/server/auth"
)

// SettingsHandler gets or updates settings for an authenticated user.
func SettingsHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil {
			// This should never happen.
			return
		}

		if r.Method == http.MethodPost {
			if err := r.ParseMultipartForm(1 << 10); err != nil {
				handleError(w, r, err)
				return
			}

			formPart, ok := r.MultipartForm.Value["form"]
			defer r.MultipartForm.RemoveAll()
			if !ok {
				err := fmt.Errorf("cannot extract form part")
				handleError(w, r, err)
				return
			}
			values, err := url.ParseQuery(formPart[0])
			if err != nil {
				handleError(w, r, err)
				return
			}

			newPassword := values.Get("Password")
			if newPassword != "" {
				user.SetPassword(newPassword)
			}

			newUsername := values.Get("Username")
			err = user.SetUsername(newUsername)
			if err != nil {
				handleError(w, r, err)
				return
			}

			if err := s.db.SaveUser(user); err != nil {
				handleError(w, r, err)
				return
			}

			if user.GetUsername() != newUsername {
				// Force logout.
				err := s.cookieHandler.SetCookieUsername(w, "", false)
				if err != nil {
					log.WithError(err).Error("Error while clearing the cookie during logout")
				}
			}

			restoreFile, ok := r.MultipartForm.File["restorefile"]
			if ok {
				file, err := restoreFile[0].Open()
				if err != nil {
					handleError(w, r, err)
					return
				}
				defer file.Close()

				data, err := ioutil.ReadAll(file)
				if err != nil {
					handleError(w, r, err)
					return
				}

				if err := s.db.Restore(user, string(data)); err != nil {
					handleError(w, r, err)
					return
				}
			}

			// Reload user to return updated values.
			user, err = s.db.GetUser(user.GetUsername())
			if err != nil {
				handleError(w, r, err)
				return
			}
		}

		type clientUser struct {
			Username string
		}

		returnUser := &clientUser{Username: user.GetUsername()}

		if err := json.NewEncoder(w).Encode(returnUser); err != nil {
			handleError(w, r, err)
		}
	}
}

// BackupHandler returns a serialized backup of all data for an authenticated user.
func BackupHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil {
			// This should never happen.
			return
		}

		data, err := s.db.Backup(user)
		if err != nil {
			handleError(w, r, err)
			return
		}

		currentTime := time.Now()
		filename := "vogon-" + currentTime.Format(time.RFC3339) + ".json"

		w.Header().Set("Content-Disposition", "attachment; filename="+filename)

		http.ServeContent(w, r, filename, currentTime, bytes.NewReader([]byte(data)))
	}
}

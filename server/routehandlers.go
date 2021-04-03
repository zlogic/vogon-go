package server

import (
	"bytes"
	"net/http"
	"net/url"
	"path"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/zlogic/vogon-go/data"
	"github.com/zlogic/vogon-go/server/auth"
)

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	log.WithError(err).Error("Error while handling request")
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func handleNotFound(w http.ResponseWriter, r *http.Request, key string) {
	log.Errorf("Item %v not found", key)
	http.Error(w, "Not found", http.StatusNotFound)
}

// PageAuthHandler checks to see if an HTML page is accessed by an authorized user,
// and redirects to the login page if the request is done by an unauthorized user.
func PageAuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil {
			http.Redirect(w, r, "login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loadTemplate(s *Services, pageName string) (*template.Template, error) {
	return template.ParseFS(s.templates, "layout.html", path.Join("pages", pageName+".html"))
}

type viewData struct {
	User     *data.User
	Username string
	Name     string
	Form     url.Values
}

// RootHandler handles the root url.
// It redirects authenticated users to the default page and unauthenticated users to the login page.
func RootHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Light check for authentication cookie - prevent errors from liveness probe.
		var url string
		if !s.cookieHandler.HasAuthenticationCookie(r) {
			url = "login"
		} else {
			url = "transactions"
		}
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

// LogoutHandler logs out the user.
func LogoutHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.cookieHandler.SetCookieUsername(w, "", false)
		if err != nil {
			log.WithError(err).Error("Error while clearing the cookie during logout")
		}
		http.Redirect(w, r, "login", http.StatusSeeOther)
	}
}

// FaviconHandler serves the favicon.
func FaviconHandler(w http.ResponseWriter, r *http.Request) {
	data, err := staticContent.ReadFile(faviconFilename)
	if err != nil {
		handleError(w, r, err)
		return
	}

	f, err := staticContent.Open(faviconFilename)
	if err != nil {
		handleError(w, r, err)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		handleError(w, r, err)
		return
	}

	http.ServeContent(w, r, "favicon.ico", stat.ModTime(), bytes.NewReader(data))
}

// HTMLLoginHandler serves the login page.
func HTMLLoginHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := loadTemplate(s, "login")
		if err != nil {
			handleError(w, r, err)
			return
		}
		type loginData struct {
			viewData
			RegistrationAllowed bool
		}
		t.ExecuteTemplate(w, "layout", &loginData{RegistrationAllowed: registrationAllowed()})
	}
}

// HTMLRegisterHandler serves the register page.
func HTMLRegisterHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user != nil {
			http.Redirect(w, r, "transactions", http.StatusSeeOther)
			return
		}
		t, err := loadTemplate(s, "register")
		if err != nil {
			handleError(w, r, err)
			return
		}
		t.ExecuteTemplate(w, "layout", &viewData{})
	}
}

// HTMLUserPageHandler serves a user-specific page.
func HTMLUserPageHandler(s *Services, templateName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil {
			// This should never happen.
			return
		}

		t, err := loadTemplate(s, templateName)
		if err != nil {
			handleError(w, r, err)
			return
		}

		if err := r.ParseForm(); err != nil {
			handleError(w, r, err)
			return
		}

		t.ExecuteTemplate(w, "layout", &viewData{User: user, Username: user.GetUsername(), Name: templateName, Form: r.Form})
	}
}

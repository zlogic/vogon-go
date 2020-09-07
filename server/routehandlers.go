package server

import (
	"net/http"
	"net/url"
	"path"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/zlogic/vogon-go/data"
)

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	log.WithError(err).Error("Error while handling request")
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func validateUser(w http.ResponseWriter, r *http.Request, s *Services) string {
	username := s.cookieHandler.GetUsername(w, r)
	if username == "" {
		http.Redirect(w, r, "login", http.StatusSeeOther)
	}
	return username
}

func loadTemplate(pageName string) (*template.Template, error) {
	return template.ParseFiles(path.Join("templates", "layout.html"), path.Join("templates", "pages", pageName+".html"))
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
		username := s.cookieHandler.GetUsername(w, r)
		var url string
		if username == "" {
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
		cookie := s.cookieHandler.NewCookie()
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "login", http.StatusSeeOther)
	}
}

// FaviconHandler serves the favicon.
func FaviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join("static", "favicon.ico"))
}

// HTMLLoginHandler serves the login page.
func HTMLLoginHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		username := s.cookieHandler.GetUsername(w, r)
		if username != "" {
			http.Redirect(w, r, "transactions", http.StatusSeeOther)
			return
		}
		t, err := loadTemplate("login")
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
		username := s.cookieHandler.GetUsername(w, r)
		if username != "" {
			http.Redirect(w, r, "transactions", http.StatusSeeOther)
			return
		}
		t, err := loadTemplate("register")
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
		username := validateUser(w, r, s)
		if username == "" {
			return
		}
		user, err := s.db.GetUser(username)
		if err != nil {
			handleError(w, r, err)
			return
		}

		t, err := loadTemplate(templateName)
		if err != nil {
			handleError(w, r, err)
			return
		}

		if err := r.ParseForm(); err != nil {
			handleError(w, r, err)
			return
		}

		t.ExecuteTemplate(w, "layout", &viewData{User: user, Username: username, Name: templateName, Form: r.Form})
	}
}

package server

import (
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
	http.ServeFile(w, r, path.Join("static", "favicon.ico"))
}

// HTMLLoginHandler serves the login page.
func HTMLLoginHandler(s *Services) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
		user := auth.GetUser(r.Context())
		if user != nil {
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
		user := auth.GetUser(r.Context())
		if user == nil {
			// This should never happen.
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

		t.ExecuteTemplate(w, "layout", &viewData{User: user, Username: user.GetUsername(), Name: templateName, Form: r.Form})
	}
}

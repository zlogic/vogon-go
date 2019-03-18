package server

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func registrationAllowed() bool {
	allowRegistrationStr, ok := os.LookupEnv("ALLOW_REGISTRATION")
	var allowRegistration bool
	if !ok {
		allowRegistrationStr = "true"
	}
	allowRegistration, err := strconv.ParseBool(allowRegistrationStr)
	if err != nil {
		log.WithField("allowregistration", allowRegistrationStr).WithError(err).Error("Cannot parse parameter specifying if registration is allowed")
		return false
	}
	return allowRegistration
}

func CreateRouter(s *Services) (*mux.Router, error) {
	registrationAllowed := registrationAllowed()

	r := mux.NewRouter()
	r.HandleFunc("/", RootHandler(s)).Methods("GET")
	r.HandleFunc("/login", HtmlLoginHandler(s)).Methods("GET")
	if registrationAllowed {
		r.HandleFunc("/register", HtmlRegisterHandler(s)).Methods("GET")
	}
	r.HandleFunc("/logout", LogoutHandler(s)).Methods("GET")
	r.HandleFunc("/transactions", HtmlUserPageHandler(s)).Methods("GET").Name("transactions")
	r.HandleFunc("/accounts", HtmlUserPageHandler(s)).Methods("GET").Name("accounts")
	r.HandleFunc("/settings", HtmlUserPageHandler(s)).Methods("GET").Name("settings")
	r.HandleFunc("/favicon.ico", FaviconHandler)
	fs := http.FileServer(staticResourceFileSystem{http.Dir("static")})
	r.PathPrefix("/static/").Handler(http.StripPrefix(strings.TrimRight("/static", "/"), fs))

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/login", LoginHandler(s)).Methods("POST")
	if registrationAllowed {
		api.HandleFunc("/register", RegisterHandler(s)).Methods("POST")
	}
	return r, nil
}

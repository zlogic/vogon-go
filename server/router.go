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

// CreateRouter returns a router and all handlers.
func CreateRouter(s *Services) (*mux.Router, error) {
	registrationAllowed := registrationAllowed()

	r := mux.NewRouter()
	r.HandleFunc("/", RootHandler(s)).Methods("GET")
	r.HandleFunc("/login", HTMLLoginHandler(s)).Methods("GET")
	if registrationAllowed {
		r.HandleFunc("/register", HTMLRegisterHandler(s)).Methods("GET")
	}
	r.HandleFunc("/logout", LogoutHandler(s)).Methods("GET")
	r.HandleFunc("/transactions", HTMLUserPageHandler(s)).Methods("GET").Name("transactions")
	r.HandleFunc("/transactioneditor", HTMLUserPageHandler(s)).Methods("GET").Name("transactioneditor")
	r.HandleFunc("/report", HTMLUserPageHandler(s)).Methods("POST").Name("report")
	r.HandleFunc("/accounts", HTMLUserPageHandler(s)).Methods("GET").Name("accounts")
	r.HandleFunc("/accounteditor", HTMLUserPageHandler(s)).Methods("GET").Name("accounteditor")
	r.HandleFunc("/settings", HTMLUserPageHandler(s)).Methods("GET").Name("settings")
	r.HandleFunc("/favicon.ico", FaviconHandler)
	fs := http.FileServer(staticResourceFileSystem{http.Dir("static")})
	r.PathPrefix("/static/").Handler(http.StripPrefix(strings.TrimRight("/static", "/"), fs))

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/login", LoginHandler(s)).Methods("POST")
	if registrationAllowed {
		api.HandleFunc("/register", RegisterHandler(s)).Methods("POST")
	}
	api.HandleFunc("/settings", SettingsHandler(s)).Methods("GET", "POST")
	api.HandleFunc("/backup", BackupHandler(s)).Methods("POST")
	api.HandleFunc("/transactions/getcount", TransactionsCountHandler(s)).Methods("POST")
	api.HandleFunc("/transactions/getpage", TransactionsHandler(s)).Methods("POST")
	api.HandleFunc("/transaction/{id}", TransactionHandler(s)).Methods("GET", "POST", "DELETE")
	api.HandleFunc("/report", ReportHandler(s)).Methods("POST")
	api.HandleFunc("/accounts", AccountsHandler(s)).Methods("GET")
	api.HandleFunc("/account/{id}", AccountHandler(s)).Methods("GET", "POST", "DELETE")
	api.HandleFunc("/tags", TagsHandler(s)).Methods("GET")
	return r, nil
}

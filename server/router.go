package server

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// NoCacheHeaderMiddlewareFunc creates a handler to disable caching.
func NoCacheHeaderMiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private")
		next.ServeHTTP(w, r)
	})
}

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
func CreateRouter(s Services) (*mux.Router, error) {
	registrationAllowed := registrationAllowed()

	r := mux.NewRouter()
	r.HandleFunc("/", RootHandler(s)).Methods(http.MethodGet)
	r.HandleFunc("/login", HTMLLoginHandler(s)).Methods(http.MethodGet)
	if registrationAllowed {
		r.HandleFunc("/register", HTMLRegisterHandler(s)).Methods(http.MethodGet)
	}
	r.HandleFunc("/logout", LogoutHandler(s)).Methods(http.MethodGet)
	r.HandleFunc("/transactions", HTMLUserPageHandler(s)).Methods(http.MethodGet).Name("transactions")
	r.HandleFunc("/transactioneditor", HTMLUserPageHandler(s)).Methods(http.MethodGet).Name("transactioneditor")
	r.HandleFunc("/report", HTMLUserPageHandler(s)).Methods(http.MethodPost).Name("report")
	r.HandleFunc("/accounts", HTMLUserPageHandler(s)).Methods(http.MethodGet).Name("accounts")
	r.HandleFunc("/accounteditor", HTMLUserPageHandler(s)).Methods(http.MethodGet).Name("accounteditor")
	r.HandleFunc("/settings", HTMLUserPageHandler(s)).Methods(http.MethodGet).Name("settings")
	r.HandleFunc("/favicon.ico", FaviconHandler)
	fs := http.FileServer(staticResourceFileSystem{http.Dir("static")})
	r.PathPrefix("/static/").Handler(http.StripPrefix(strings.TrimRight("/static", "/"), fs))

	api := r.PathPrefix("/api").Subrouter()
	api.Use(NoCacheHeaderMiddlewareFunc)
	api.HandleFunc("/login", LoginHandler(s)).Methods(http.MethodPost)
	if registrationAllowed {
		api.HandleFunc("/register", RegisterHandler(s)).Methods(http.MethodPost)
	}
	api.HandleFunc("/settings", SettingsHandler(s)).Methods(http.MethodGet, http.MethodPost)
	api.HandleFunc("/backup", BackupHandler(s)).Methods(http.MethodPost)
	api.HandleFunc("/transactions/getcount", TransactionsCountHandler(s)).Methods(http.MethodPost)
	api.HandleFunc("/transactions/getpage", TransactionsHandler(s)).Methods(http.MethodPost)
	api.HandleFunc("/transaction/{id}", TransactionHandler(s)).Methods(http.MethodGet, http.MethodPost, http.MethodDelete)
	api.HandleFunc("/report", ReportHandler(s)).Methods(http.MethodPost)
	api.HandleFunc("/accounts", AccountsHandler(s)).Methods(http.MethodGet)
	api.HandleFunc("/account/{id}", AccountHandler(s)).Methods(http.MethodGet, http.MethodPost, http.MethodDelete)
	api.HandleFunc("/tags", TagsHandler(s)).Methods(http.MethodGet)
	return r, nil
}

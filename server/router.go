package server

import (
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
)

// NoCacheHeaderMiddlewareFunc creates a handler to disable caching.
func NoCacheHeaderMiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private")
		next.ServeHTTP(w, r)
	})
}

func parseBoolEnv(varName string, defaultValue bool) bool {
	valueStr, _ := os.LookupEnv(varName)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.WithField("variable", varName).WithField("value", value).WithError(err).Error("Cannot parse environment value")
		return defaultValue
	}
	return value
}

func registrationAllowed() bool {
	return parseBoolEnv("ALLOW_REGISTRATION", true)
}

// CreateRouter returns a router and all handlers.
func CreateRouter(s *Services) (*chi.Mux, error) {
	registrationAllowed := registrationAllowed()
	logRequests := parseBoolEnv("LOG_REQUESTS", true)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	if logRequests {
		r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: log.New(), NoColor: true}))
	}
	r.Use(middleware.Recoverer)

	r.Get("/", RootHandler(s))
	r.Get("/login", HTMLLoginHandler(s))
	if registrationAllowed {
		r.Get("/register", HTMLRegisterHandler(s))
	}
	r.Get("/logout", LogoutHandler(s))
	r.Get("/transactions", HTMLUserPageHandler(s, "transactions"))
	r.Get("/transactioneditor", HTMLUserPageHandler(s, "transactioneditor"))
	r.Post("/report", HTMLUserPageHandler(s, "report"))
	r.Get("/accounts", HTMLUserPageHandler(s, "accounts"))
	r.Get("/accounteditor", HTMLUserPageHandler(s, "accounteditor"))
	r.Get("/settings", HTMLUserPageHandler(s, "settings"))
	r.HandleFunc("/favicon.ico", FaviconHandler)
	fs := http.FileServer(staticResourceFileSystem{http.Dir("static")})
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Route("/api", func(api chi.Router) {
		api.Use(NoCacheHeaderMiddlewareFunc)
		api.Post("/login", LoginHandler(s))
		if registrationAllowed {
			api.Post("/register", RegisterHandler(s))
		}
		api.Get("/settings", SettingsHandler(s))
		api.Post("/settings", SettingsHandler(s))
		api.Post("/backup", BackupHandler(s))
		api.Post("/transactions/getcount", TransactionsCountHandler(s))
		api.Post("/transactions/getpage", TransactionsHandler(s))
		api.Get("/transaction/{id}", TransactionHandler(s))
		api.Post("/transaction/{id}", TransactionHandler(s))
		api.Delete("/transaction/{id}", TransactionHandler(s))
		api.Post("/report", ReportHandler(s))
		api.Get("/accounts", AccountsHandler(s))
		api.Get("/account/{id}", AccountHandler(s))
		api.Post("/account/{id}", AccountHandler(s))
		api.Delete("/account/{id}", AccountHandler(s))
		api.Get("/tags", TagsHandler(s))
	})
	return r, nil
}

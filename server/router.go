package server

import (
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
		r.Group(func(authorized chi.Router) {
			authorized.Use(s.cookieHandler.AuthHandlerFunc)
			authorized.Get("/register", HTMLRegisterHandler(s))
		})
	}

	r.Group(func(authorized chi.Router) {
		authorized.Use(s.cookieHandler.AuthHandlerFunc)
		authorized.Use(PageAuthHandler)
		authorized.Get("/logout", LogoutHandler(s))
		authorized.Get("/transactions", HTMLUserPageHandler(s, "transactions"))
		authorized.Get("/transactioneditor", HTMLUserPageHandler(s, "transactioneditor"))
		authorized.Post("/report", HTMLUserPageHandler(s, "report"))
		authorized.Get("/accounts", HTMLUserPageHandler(s, "accounts"))
		authorized.Get("/accounteditor", HTMLUserPageHandler(s, "accounteditor"))
		authorized.Get("/settings", HTMLUserPageHandler(s, "settings"))
	})
	r.HandleFunc("/favicon.ico", FaviconHandler)
	fs := http.FileServer(staticResourceFileSystem{http.FS(staticContent)})
	r.Handle("/static/*", fs)

	r.Route("/api", func(api chi.Router) {
		api.Use(NoCacheHeaderMiddlewareFunc)
		api.Post("/login", LoginHandler(s))
		if registrationAllowed {
			api.Post("/register", RegisterHandler(s))
		}
		api.Group(func(authorized chi.Router) {
			authorized.Use(s.cookieHandler.AuthHandlerFunc)
			authorized.Use(APIAuthHandler)
			authorized.Get("/settings", SettingsHandler(s))
			authorized.Post("/settings", SettingsHandler(s))
			authorized.Post("/backup", BackupHandler(s))
			authorized.Post("/transactions/getcount", TransactionsCountHandler(s))
			authorized.Post("/transactions/getpage", TransactionsHandler(s))
			authorized.Get("/transaction/{id}", TransactionHandler(s))
			authorized.Post("/transaction/{id}", TransactionHandler(s))
			authorized.Delete("/transaction/{id}", TransactionHandler(s))
			authorized.Post("/report", ReportHandler(s))
			authorized.Get("/accounts", AccountsHandler(s))
			authorized.Get("/account/{id}", AccountHandler(s))
			authorized.Post("/account/{id}", AccountHandler(s))
			authorized.Delete("/account/{id}", AccountHandler(s))
			authorized.Get("/tags", TagsHandler(s))
		})
	})
	return r, nil
}

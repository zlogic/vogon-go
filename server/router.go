package server

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func CreateRouter(s *Services) (*mux.Router, error) {
	r := mux.NewRouter()
	r.HandleFunc("/", RootHandler(s)).Methods("GET")
	r.HandleFunc("/login", HtmlLoginHandler(s)).Methods("GET")
	r.HandleFunc("/register", HtmlRegisterHandler(s)).Methods("GET").Name("settings")
	r.HandleFunc("/logout", LogoutHandler(s)).Methods("GET")
	r.HandleFunc("/settings", HtmlSettingsHandler(s)).Methods("GET").Name("settings")
	r.HandleFunc("/favicon.ico", FaviconHandler)
	fs := http.FileServer(staticResourceFileSystem{http.Dir("static")})
	r.PathPrefix("/static/").Handler(http.StripPrefix(strings.TrimRight("/static", "/"), fs))

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/login", LoginHandler(s)).Methods("POST")
	api.HandleFunc("/register", RegisterHandler(s)).Methods("POST")
	return r, nil
}

package web

import "github.com/gorilla/mux"

// RegisterRoutes adds all registered routes for this app to the
// main router. This should make it easier to get a quick overview
// over all routes.
func RegisterRoutes(r *mux.Router) {
	// all for /oauth
	oauth := r.PathPrefix("/oauth").Subrouter()
	oauth.HandleFunc("/authorize", Authorize).Methods("GET")
	oauth.HandleFunc("/login_page", LoginPage).Methods("GET")
	oauth.HandleFunc("/login", Login).Methods("POST")
}

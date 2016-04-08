package web

import (
	"fmt"
	"net/http"
)

// Error404 handles not found errors
type Error404 struct {
}

// ServeHTTP implements HandleFunc for Error404
func (err *Error404) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprintf(w, "<h1>Does not exist</h4><p>%s</p>", r.URL)
}

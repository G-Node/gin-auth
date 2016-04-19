package web

import (
	"github.com/G-Node/gin-auth/data"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// InitTestHttpHandler initializes a handler with all registered routes and returns it
// along with a response recorder.
func InitTestHttpHandler(t *testing.T) (*httptest.ResponseRecorder, http.Handler) {
	data.InitTestDb(t)
	router := mux.NewRouter()
	router.NotFoundHandler = &NotFoundHandler{}
	RegisterRoutes(router)
	return httptest.NewRecorder(), router
}

func TestAuthorize(t *testing.T) {

	response, handler := InitTestHttpHandler(t)
	request, _ := http.NewRequest("GET", "/oauth/authorize", strings.NewReader(""))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Error("Response code 400 expected")
	}
}

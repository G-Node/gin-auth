package web

import (
	"encoding/json"
	"fmt"
	"github.com/G-Node/gin-auth/util"
	"html/template"
	"net/http"
)

var codes = map[int]string{
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Time-out",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Request Entity Too Large",
	414: "Request-URL Too Long",
	415: "Unsupported Media Type",
	416: "Requested range not satisfiable",
	417: "Expectation Failed",
	418: "I’m a teapot",
	420: "Policy Not Fulfilled",
	421: "Misdirected Request",
	422: "Unprocessable Entity",
	423: "Locked",
	424: "Failed Dependency",
	425: "Unordered Collection",
	426: "Upgrade Required",
	428: "Precondition Required",
	429: "Too Many Requests",
	431: "Request Header Fields Too Large",
	451: "Unavailable For Legal Reasons",
	500: "Internal Server Error",
}

// NotFoundHandler deals with not found errors
type NotFoundHandler struct{}

// ServeHTTP implements HandleFunc for NotFoundHandler
func (err *NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	PrintErrorHTML(w, r, "The requested site does not exist.", http.StatusNotFound)
}

type errorData struct {
	Code    int               `json:"code"`
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Reasons map[string]string `json:"reasons"`
}

func (dat *errorData) FillFrom(err interface{}, code int) {
	dat.Code = code
	head, ok := codes[code]
	if ok {
		dat.Error = head
	} else {
		if code > 500 {
			dat.Error = "Internal Server Error"
		} else {
			dat.Error = "Unknown Error"
		}
	}

	switch err := err.(type) {
	case *util.ValidationError:
		dat.Message = err.Message
		dat.Reasons = err.FieldErrors
	case error:
		dat.Message = err.Error()
	case fmt.Stringer:
		dat.Message = err.String()
	case string:
		dat.Message = err
	}
}

type htmlErrorData struct {
	errorData
	Referrer string
}

// PrintErrorHTML shows an html error page.
func PrintErrorHTML(w http.ResponseWriter, r *http.Request, err interface{}, code int) {
	errData := &htmlErrorData{Referrer: r.Referer()}
	errData.FillFrom(err, code)

	tmpl, err := template.ParseFiles("assets/html/layout.html", "assets/html/error.html")

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(code)
	tmpl.ExecuteTemplate(w, "layout", errData)
}

// PrintErrorJSON writes an JSON error response.
func PrintErrorJSON(w http.ResponseWriter, r *http.Request, err interface{}, code int) {
	errData := &errorData{}
	errData.FillFrom(err, code)
	for k, v := range errData.Reasons {
		delete(errData.Reasons, k)
		errData.Reasons[util.ToSnakeCase(k)] = v
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.Encode(errData)
}

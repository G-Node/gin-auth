package web

import (
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

// Error404 handles not found errors
type Error404 struct {
}

// ServeHTTP implements HandleFunc for Error404
func (err *Error404) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprintf(w, "<h1>Does not exist</h1><p>%s</p>", r.URL)
}

type errorData struct {
	Code     int
	HeadLine string
	Message  string
	Reasons  []string
	Referrer string
}

func PrintErrorHTML(w http.ResponseWriter, r *http.Request, err interface{}, code int) {
	head, ok := codes[code]
	if !ok {
		head = "Unknown Error"
	}

	data := &errorData{
		Code:     code,
		HeadLine: head,
		Referrer: r.Referer(),
		Reasons:  make([]string, 0),
	}

	switch err := err.(type) {
	case *util.ValidationError:
		data.Message = err.Message
		for field, msg := range err.FieldErrors {
			data.Reasons = append(data.Reasons, fmt.Sprintf("%s: %s", field, msg))
		}
	case error:
		data.Message = err.Error()
	case fmt.Stringer:
		data.Message = err.String()
	case string:
		data.Message = err
	}

	tmpl, err := template.ParseFiles("assets/html/layout.html", "assets/html/error.html")

	w.WriteHeader(code)
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "layout", data)
}

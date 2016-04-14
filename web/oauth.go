package web

import (
	"fmt"
	"net/http"

	"database/sql"
	"github.com/G-Node/gin-auth/data"
	"github.com/G-Node/gin-auth/util"
	"html/template"
)

type authorize struct {
	ResponseType string
	ClientId     string
	RedirectURI  string
	State        string
	Scope        []string
}

// Authorize handles the beginning of an OAuth grant request following the schema
// of 'implicit' or 'code' grant types.
func Authorize(w http.ResponseWriter, r *http.Request) {
	param := &authorize{}
	err := util.ReadQueryIntoStruct(r, param, false)
	if err != nil {
		// TODO nice error handling
		panic(err)
	}

	if !(param.ResponseType == "code" || param.ResponseType == "token") {
		// TODO nice error handling
		panic("response_type must be code or token")
	}

	grantRequest := &data.GrantRequest{}
	grantRequest.GrantType = param.ResponseType

	client, ok := data.GetClientByName(param.ClientId)
	if !ok {
		// TODO nice error handling
		panic(fmt.Sprintf("Cliet '%s' does not exist", param.ClientId))
	}
	grantRequest.ClientUUID = client.UUID

	if !util.StringInSlice(client.RedirectURIs, param.RedirectURI) {
		// TODO nice error handling
		panic(fmt.Sprintf("Redirect URI '%s' not registered for client", param.RedirectURI))
	}
	grantRequest.RedirectURI = param.RedirectURI

	err = grantRequest.Create()
	if err != nil {
		// TODO nice error handling
		panic("Unable to save grant request")
	}

	w.Header().Add("Cache-Control", "no-store")
	http.Redirect(w, r, "/oauth/login?request_id="+grantRequest.Token, http.StatusFound)
}

type loginData struct {
	Login     string
	Password  string
	RequestID string
}

var loginTmpl, _ = template.ParseFiles("assets/html/layout.html", "assets/html/login.html")

func LoginPage(w http.ResponseWriter, r *http.Request) {
	// TODO check for session

	query := r.URL.Query()
	if query == nil {
		// TODO nice error handling
		panic("Query parameter 'request_id' was missing")
	}
	token := query.Get("request_id")

	_, ok := data.GetGrantRequest(token)
	if !ok {
		// TODO nice error handling
		panic("Grant request does not exist")
	}

	w.Header().Add("Cache-Control", "no-store")
	loginTmpl.ExecuteTemplate(w, "layout", &loginData{RequestID: token})
}

func Login(w http.ResponseWriter, r *http.Request) {
	form := &loginData{}
	err := util.ReadFormIntoStruct(r, form, false)
	if err != nil {
		// TODO nice error handling
		panic(err)
	}

	grantRequest, ok := data.GetGrantRequest(form.RequestID)
	if !ok {
		// TODO nice error handling
		panic("Grant request does not exist")
	}

	account, ok := data.GetAccountByLogin(form.Login)
	if !ok {
		// TODO nice error handling
		panic("Account does not exist")
	}

	ok = account.VerifyPassword(form.Password)
	if !ok {
		// TODO nice error handling
		panic("Wrong password")
	}

	grantRequest.AccountUUID = sql.NullString{String: account.UUID, Valid: true}
	grantRequest.Update()

	// TODO continue login
	fmt.Fprintln(w, form)
	fmt.Fprintln(w, grantRequest)
}

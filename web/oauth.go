package web

import (
	"fmt"
	"net/http"

	"github.com/G-Node/gin-auth/data"
	"github.com/G-Node/gin-auth/util"
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

	http.Redirect(w, r, "/oauth/login?request_id="+grantRequest.Token, 302)
}

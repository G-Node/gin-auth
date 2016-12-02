// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package web

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/G-Node/gin-auth/data"
	"github.com/G-Node/gin-auth/util"
)

// createGrantRequest creates a Grant Request for a client and redirects to a forwarding URI.
func createGrantRequest(w http.ResponseWriter, r *http.Request, forwardURI string) {
	param := &struct {
		ResponseType string
		ClientId     string
		RedirectURI  string
		State        string
		Scope        string
	}{}

	err := util.ReadQueryIntoStruct(r, param, false)
	if err != nil {
		PrintErrorHTML(w, r, err, http.StatusBadRequest)
		return
	}

	client, ok := data.GetClientByName(param.ClientId)
	if !ok {
		PrintErrorHTML(w, r, fmt.Sprintf("Client '%s' does not exist", param.ClientId), http.StatusBadRequest)
		return
	}

	scope := util.NewStringSet(strings.Split(param.Scope, " ")...)
	request, err := client.CreateGrantRequest(param.ResponseType, param.RedirectURI, param.State, scope)
	if err != nil {
		PrintErrorHTML(w, r, err, http.StatusBadRequest)
		return
	}

	queryVals := &url.Values{}
	queryVals.Add("request_id", request.Token)
	w.Header().Add("Cache-Control", "no-store")
	http.Redirect(w, r, forwardURI+"?"+queryVals.Encode(), http.StatusFound)
}

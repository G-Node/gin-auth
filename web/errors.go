// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/util"
)

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
	dat.Error = http.StatusText(code)
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

	tmpl := conf.MakeTemplate("error.html")
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(code)
	err = tmpl.ExecuteTemplate(w, "layout", errData)
	if err != nil {
		panic(err)
	}
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
	err = enc.Encode(errData)
	if err != nil {
		panic(err)
	}
}

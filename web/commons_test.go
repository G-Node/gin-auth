// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/G-Node/gin-auth/data"
	"github.com/G-Node/gin-auth/util"
)

func TestCreateGrantRequest(t *testing.T) {
	defer util.FailOnPanic(t)
	data.InitTestDb(t)

	const invalidClientId = "iDoNotExist"
	const invalidScope = "iDoNotExist"

	const validResponseType = "code"
	const validClientId = "gin"
	const validRedirectURI = "http://localhost:8080/notice"
	const validState = "clientState"
	const validScope = "account-create"

	const forwardURI = "/forward/uri"

	queryVals := &url.Values{}
	queryVals.Add("response_type", validResponseType)
	queryVals.Add("client_id", validClientId)
	queryVals.Add("redirect_uri", validRedirectURI)
	queryVals.Add("state", validState)
	queryVals.Add("scope", validScope)

	// Ensure fail if any of the fields is missing
	// Test missing response type
	queryVals.Del("response_type")
	request, _ := http.NewRequest("GET", "/root?"+queryVals.Encode(), strings.NewReader(""))
	response := httptest.NewRecorder()
	createGrantRequest(response, request, forwardURI)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected code %d but got %d\n", http.StatusBadRequest, response.Code)
	}

	// Test missing client id
	queryVals.Add("response_type", validResponseType)
	queryVals.Del("client_id")
	request, _ = http.NewRequest("GET", "/root?"+queryVals.Encode(), strings.NewReader(""))
	response = httptest.NewRecorder()
	createGrantRequest(response, request, forwardURI)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected code %d but got %d\n", http.StatusBadRequest, response.Code)
	}

	// Test missing redirect uri
	queryVals.Add("client_id", validClientId)
	queryVals.Del("redirect_uri")
	request, _ = http.NewRequest("GET", "/root?"+queryVals.Encode(), strings.NewReader(""))
	response = httptest.NewRecorder()
	createGrantRequest(response, request, forwardURI)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected code %d but got %d\n", http.StatusBadRequest, response.Code)
	}

	// Test missing state
	queryVals.Add("redirect_uri", validRedirectURI)
	queryVals.Del("state")
	request, _ = http.NewRequest("GET", "/root?"+queryVals.Encode(), strings.NewReader(""))
	response = httptest.NewRecorder()
	createGrantRequest(response, request, forwardURI)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected code %d but got %d\n", http.StatusBadRequest, response.Code)
	}

	// Test missing Scope
	queryVals.Add("state", validState)
	queryVals.Del("scope")
	request, _ = http.NewRequest("GET", "/root?"+queryVals.Encode(), strings.NewReader(""))
	response = httptest.NewRecorder()
	createGrantRequest(response, request, forwardURI)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected code %d but got %d\n", http.StatusBadRequest, response.Code)
	}

	// Test invalid client id
	queryVals.Add("scope", validScope)
	queryVals.Set("client_id", invalidClientId)
	request, _ = http.NewRequest("GET", "/root?"+queryVals.Encode(), strings.NewReader(""))
	response = httptest.NewRecorder()
	createGrantRequest(response, request, forwardURI)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected code %d but got %d\n", http.StatusBadRequest, response.Code)
	}

	// Test create fail using invalid scope
	queryVals.Set("client_id", validClientId)
	queryVals.Set("scope", invalidScope)
	request, _ = http.NewRequest("GET", "/root?"+queryVals.Encode(), strings.NewReader(""))
	response = httptest.NewRecorder()
	createGrantRequest(response, request, forwardURI)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected code %d but got %d\n", http.StatusBadRequest, response.Code)
	}

	// Test correct redirect and URI query after create
	queryVals.Set("scope", validScope)
	request, _ = http.NewRequest("GET", "/root?"+queryVals.Encode(), strings.NewReader(""))
	response = httptest.NewRecorder()
	createGrantRequest(response, request, forwardURI)
	if response.Code != http.StatusFound {
		t.Errorf("Expected code %d but got %d\n", http.StatusFound, response.Code)
	}
	location, err := response.Result().Location()
	if err != nil {
		t.Error(err)
	}
	if location.Path != forwardURI {
		t.Errorf("Expected forward uri to be %q but was %q\n", forwardURI, location.Path)
	}
	if !strings.Contains(location.RawQuery, "request_id=") {
		t.Errorf("Request token is missing from redirect uri query: %q\n", location.RawQuery)
	}
}

func TestRedirectionScript(t *testing.T) {
	const uri = "https://example.com"
	const delay = 500

	script := redirectionScript(uri, delay)

	if !strings.Contains(script, uri) {
		t.Errorf("Script block does not contain uri %q: \n%q\n", uri, script)
	}
	if !strings.Contains(script, strconv.Itoa(delay)) {
		t.Errorf("Script block does not contain delay %q: \n%q\n", delay, script)
	}
}

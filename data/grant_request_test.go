// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"database/sql"
	"testing"

	"github.com/G-Node/gin-auth/util"
)

const (
	grantReqTokenAlice        = "U7JIKKYI"
	grantReqTokenAliceExpired = "AGTBAI3D"
	grantReqCodeAlice         = "HGZQP6WE"
	grantReqCodeAliceExpired  = "KWANG2G4"
	grantReqTokenBob          = "B4LIMIMB"
	grantReqWBAlice           = "QH92T99D"
)

func TestListGrantRequests(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	requests := ListGrantRequests()
	if len(requests) != 3 {
		t.Errorf("Exactly 3 grant requests expected in list but was %d", len(requests))
	}
}

func TestGetGrantRequest(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	req, ok := GetGrantRequest(grantReqTokenAlice)
	if !ok {
		t.Error("Grant request does not exist")
	}
	if !req.ScopeRequested.Contains("repo-read") {
		t.Errorf("Requested scope should contain 'repo-read'")
	}

	_, ok = GetGrantRequest("doesNotExist")
	if ok {
		t.Error("Grant request should not exist")
	}

	_, ok = GetGrantRequest(grantReqTokenAliceExpired)
	if ok {
		t.Error("Expired grant request should not be retrieved.")
	}
}

func TestGetGrantRequestByCode(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	req, ok := GetGrantRequestByCode(grantReqCodeAlice)
	if !ok {
		t.Error("Grant request does not exist")
	}
	if !req.ScopeRequested.Contains("repo-read") {
		t.Errorf("Requested scope should contain  'repo-read'")
	}

	_, ok = GetGrantRequestByCode("doesNotExist")
	if ok {
		t.Error("Grant request should not exist")
	}

	_, ok = GetGrantRequestByCode(grantReqCodeAliceExpired)
	if ok {
		t.Error("Expired grant request should not be retrieved.")
	}
}

func TestGrantRequest_ExchangeCodeForTokens(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	req, ok := GetGrantRequest(grantReqTokenAlice)
	if !ok {
		t.Error("Grant request does not exist")
	}

	accessToken, refreshToken, err := req.ExchangeCodeForTokens()
	if err != nil {
		t.Error(err)
	}

	access, ok := GetAccessToken(accessToken)
	if !ok {
		t.Error("Unable to find created access token")
	}
	if !access.AccountUUID.Valid || access.AccountUUID.String != uuidAlice {
		t.Error("Access token has a wrong account UUID")
	}

	refresh, ok := GetRefreshToken(refreshToken)
	if !ok {
		t.Error("Unable to find created refresh token")
	}
	if refresh.AccountUUID != uuidAlice {
		t.Error("Refresh token has a wrong account UUID")
	}

	_, ok = GetGrantRequest(grantReqTokenAlice)
	if ok {
		t.Error("Grant request was expected to be deleted")
	}
}

func TestGrantRequest_Create(t *testing.T) {
	InitTestDb(t)

	token := util.RandomToken()
	state := util.RandomToken()
	code := util.RandomToken()
	fresh := GrantRequest{
		Token:          token,
		GrantType:      "code",
		State:          state,
		Code:           sql.NullString{String: code, Valid: true},
		ScopeRequested: util.NewStringSet("foo-read", "foo-write", "foo-admin"),
		ClientUUID:     uuidClientGin,
		AccountUUID:    sql.NullString{String: uuidAlice, Valid: true}}

	err := fresh.Create()
	if err != nil {
		t.Error(err)
	}

	check, ok := GetGrantRequest(token)
	if !ok {
		t.Error("Grant request does not exist")
	}
	if check.State != state {
		t.Error("State does not match")
	}
	if check.Code.Valid && check.Code.String != code {
		t.Error("Code does not match")
	}
	if !check.ScopeRequested.Contains("foo-read") {
		t.Error("Requested scope should contain 'foo-read'")
	}
}

func TestGrantRequest_Client(t *testing.T) {
	InitTestDb(t)
	defer util.FailOnPanic(t)

	req, ok := GetGrantRequest(grantReqTokenAlice)
	if !ok {
		t.Error("Grant request does not exist")
	}

	client := req.Client()
	if client.Name != "gin" {
		t.Error("Client name expected to be 'gin'")
	}
}

func TestGrantRequest_Update(t *testing.T) {
	InitTestDb(t)

	newCode := util.RandomToken()
	newState := util.RandomToken()

	req, ok := GetGrantRequest(grantReqTokenAlice)
	if !ok {
		t.Error("Grant request does not exist")
	}

	req.Code = sql.NullString{String: newCode, Valid: true}
	req.State = newState

	err := req.Update()
	if err != nil {
		t.Error(err)
	}

	check, ok := GetGrantRequest(grantReqTokenAlice)
	if !ok {
		t.Error("Grant request does not exist")
	}
	if check.Code.Valid && check.Code.String != newCode {
		t.Error("Code does not match")
	}
	if check.State != newState {
		t.Error("State does not match")
	}
}

func TestGrantRequest_IsApproved(t *testing.T) {
	InitTestDb(t)

	// request with approved client
	request, ok := GetGrantRequest(grantReqTokenAlice)
	if !ok {
		t.Error("Grant request does not exist")
	}
	if !request.IsApproved() {
		t.Error("Grant request should be approved")
	}

	// request without approval
	request, ok = GetGrantRequest(grantReqTokenBob)
	if !ok {
		t.Error("Grant request does not exist")
	}
	if request.IsApproved() {
		t.Error("Grant request should not be approved")
	}

	// request without approval but whitelisted scope
	request, ok = GetGrantRequest(grantReqWBAlice)
	if !ok {
		t.Error("Grant request does not exist")
	}
	if !request.IsApproved() {
		t.Error("Grant request should be approved")
	}
}

func TestGrantRequest_Delete(t *testing.T) {
	InitTestDb(t)

	req, ok := GetGrantRequest(grantReqTokenAlice)
	if !ok {
		t.Error("Grant request does not exist")
	}

	err := req.Delete()
	if err != nil {
		t.Error(err)
	}

	_, ok = GetGrantRequest(uuidClientGin)
	if ok {
		t.Error("Grant request should not exist")
	}
}

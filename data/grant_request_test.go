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
	"github.com/G-Node/gin-auth/util"
	"testing"
)

const (
	grantReqTokenAlice = "U7JIKKYI"
	grantReqTokenBob   = "B4LIMIMB"
)

func TestListGrantRequests(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	requests := ListGrantRequests()
	if len(requests) != 2 {
		t.Error("Exactly two grant requests expected in list")
	}
}

func TestGetGrantRequest(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	req, ok := GetGrantRequest(grantReqTokenAlice)
	if !ok {
		t.Error("Grant request does not exist")
	}
	if req.ScopeRequested[0] != "repo-read" {
		t.Errorf("First requested scope was expected to be 'repo-read'")
	}

	_, ok = GetGrantRequest("doesNotExist")
	if ok {
		t.Error("Grant request should not exist")
	}
}

func TestCreateGrantRequest(t *testing.T) {
	initTestDb(t)

	token := util.RandomToken()
	state := util.RandomToken()
	code := util.RandomToken()
	fresh := GrantRequest{
		Token:          token,
		GrantType:      "code",
		State:          state,
		Code:           code,
		ScopeRequested: SqlStringSlice{"foo-read", "foo-write", "foo-admin"},
		ScopeApproved:  SqlStringSlice{"foo-read"},
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
	if check.Code != code {
		t.Error("Code does not match")
	}
	if check.ScopeRequested[0] != "foo-read" {
		t.Error("First requested scope was expected to be 'foo-read'")
	}
	if check.ScopeApproved[0] != "foo-read" {
		t.Error("First approved scope was expected to be 'foo-read'")
	}
}

func TestUpdateGrantRequest(t *testing.T) {
	initTestDb(t)

	newCode := util.RandomToken()
	newState := util.RandomToken()

	req, ok := GetGrantRequest(grantReqTokenAlice)
	if !ok {
		t.Error("Grant request does not exist")
	}

	req.Code = newCode
	req.State = newState

	err := req.Update()
	if err != nil {
		t.Error(err)
	}

	check, ok := GetGrantRequest(grantReqTokenAlice)
	if !ok {
		t.Error("Grant request does not exist")
	}
	if check.Code != newCode {
		t.Error("Code does not match")
	}
	if check.State != newState {
		t.Error("State does not match")
	}
}

func TestDeleteGrantRequest(t *testing.T) {
	initTestDb(t)

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

func TestGrantRequestGetApproval(t *testing.T) {
	initTestDb(t)

	req, ok := GetGrantRequest(grantReqTokenAlice)
	if !ok {
		t.Error("Grant request does not exist")
	}
	cli, ok := req.GetClientApproval()
	if !ok {
		t.Error("Approval does not exist")
	}
	if cli.ClientUUID != req.ClientUUID {
		t.Errorf("Client UUID should be '%s' but was '%s'", req.ClientUUID, cli.ClientUUID)
	}
	if cli.AccountUUID != uuidAlice {
		t.Errorf("Account UUID should be '%s' but was '%s'", uuidAlice, cli.AccountUUID)
	}

	req, ok = GetGrantRequest(grantReqTokenBob)
	if !ok {
		t.Error("Grant request does not exist")
	}
	cli, ok = req.GetClientApproval()
	if ok {
		t.Error("Approval should not exist")
	}
}

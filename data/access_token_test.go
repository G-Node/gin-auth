// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"github.com/G-Node/gin-auth/util"
	"testing"
	"time"
)

const (
	accessTokenAlice = "3N7MP7M7"
	accessTokenBob   = "LJ3W7ZFK" // is expired
)

func TestListAccessTokens(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	accessTokens := ListAccessTokens()
	if len(accessTokens) != 2 {
		t.Error("Exactly to access tokens expected in slice")
	}
}

func TestGetAccessToken(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	tok, ok := GetAccessToken(accessTokenAlice)
	if !ok {
		t.Error("Access token does not exist")
	}
	if tok.AccountUUID != uuidAlice {
		t.Errorf("AccountUUID was expectd to be '%s'", uuidAlice)
	}

	_, ok = GetAccessToken("doesNotExist")
	if ok {
		t.Error("Access token should not exist")
	}
}

func TestClearOldAccessTokens(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	deleted := ClearOldAccessTokens()
	if deleted != 1 {
		t.Error("Exactly one access token is supposed to be deleted")
	}

	_, ok := GetAccessToken(accessTokenBob)
	if ok {
		t.Error("Bobs access token should not exist")
	}
}

func TestCreateAccessToken(t *testing.T) {
	InitTestDb(t)

	token := util.RandomToken()
	fresh := AccessToken{
		Token:       token,
		Scope:       util.NewStringSet("foo-read", "foo-write"),
		Expires:     time.Now().Add(time.Hour * 12),
		ClientUUID:  uuidClientGin,
		AccountUUID: uuidAlice}

	err := fresh.Create()
	if err != nil {
		t.Error(err)
	}

	check, ok := GetAccessToken(token)
	if !ok {
		t.Error("Token does not exist")
	}
	if check.AccountUUID != uuidAlice {
		t.Errorf("AccountUUID is supposed to be '%s'", uuidAlice)
	}
	if !check.Scope.Contains("foo-read") {
		t.Error("Scope should contain 'foo-read'")
	}
	if !check.Scope.Contains("foo-write") {
		t.Error("Scope should contain 'foo-write'")
	}
}

func TestAccessTokenUpdateExpirationTime(t *testing.T) {
	InitTestDb(t)

	tok, ok := GetAccessToken(accessTokenBob)
	if !ok {
		t.Error("Access token does not exist")
	}
	if time.Since(tok.Expires) < 0 {
		t.Error("Token should be expired")
	}

	tok.UpdateExpirationTime()
	if time.Since(tok.Expires) > 0 {
		t.Error("Access token should not be expired")
	}

	check, ok := GetAccessToken(accessTokenBob)
	if !ok {
		t.Error("Access token does not exist")
	}
	if time.Since(check.Expires) > 0 {
		t.Error("Token should not be expired")
	}
}

func TestAccessTokenDelete(t *testing.T) {
	InitTestDb(t)

	tok, ok := GetAccessToken(accessTokenAlice)
	if !ok {
		t.Error("Access token does not exist")
	}

	err := tok.Delete()
	if err != nil {
		t.Error(err)
	}

	_, ok = GetAccessToken(accessTokenAlice)
	if ok {
		t.Error("Access token should not exist")
	}
}

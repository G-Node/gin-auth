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
)

const (
	refreshTokenAlice = "YYPTDSVZ"
)

func TestListRefreshTokens(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	refreshTokens := ListRefreshTokens()
	if len(refreshTokens) != 2 {
		t.Error("Exactly to refresh tokens expected in slice")
	}
}

func TestGetRefreshToken(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	tok, ok := GetRefreshToken(refreshTokenAlice)
	if !ok {
		t.Error("Refresh token does not exist")
	}
	if tok.AccountUUID != uuidAlice {
		t.Errorf("AccountUUID was expectd to be '%s'", uuidAlice)
	}

	_, ok = GetRefreshToken("doesNotExist")
	if ok {
		t.Error("Refresh token should not exist")
	}
}

func TestCreateRefreshToken(t *testing.T) {
	initTestDb(t)

	token := util.RandomToken()
	new := RefreshToken{
		Token:           token,
		Scope:           SqlStringSlice{"foo-read", "foo-write"},
		OAuthClientUUID: uuidClientGin,
		AccountUUID:     uuidAlice}

	err := new.Create()
	if err != nil {
		t.Error(err)
	}

	check, ok := GetRefreshToken(token)
	if !ok {
		t.Error("Token does not exist")
	}
	if check.AccountUUID != uuidAlice {
		t.Errorf("AccountUUID is supposed to be '%s'", uuidAlice)
	}
	if check.Scope[1] != "foo-write" {
		t.Error("Second scope is supposed to be 'foo-write'")
	}
}

func TestRefreshTokenDelete(t *testing.T) {
	initTestDb(t)

	tok, ok := GetRefreshToken(refreshTokenAlice)
	if !ok {
		t.Error("Refresh token does not exist")
	}

	err := tok.Delete()
	if err != nil {
		t.Error(err)
	}

	_, ok = GetRefreshToken(refreshTokenAlice)
	if ok {
		t.Error("Refresh token should not exist")
	}
}

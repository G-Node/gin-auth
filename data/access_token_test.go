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
	defer failOnPanic(t)
	initTestDb(t)

	accessTokens := ListAccessTokens()
	if len(accessTokens) != 2 {
		t.Error("Exactly to access tokens expected in slice")
	}
}

func TestGetAccessToken(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

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
	defer failOnPanic(t)
	initTestDb(t)

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
	initTestDb(t)

	token := util.RandomToken()
	new := AccessToken{
		Token:           token,
		Scope:           SqlStringSlice{"foo-read", "foo-write"},
		Expires:         time.Now().Add(time.Hour * 12),
		OAuthClientUUID: uuidClientGin,
		AccountUUID:     uuidAlice}

	err := new.Create()
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
	if check.Scope[1] != "foo-write" {
		t.Error("Second scope is supposed to be 'foo-write'")
	}
}

func TestAccessTokenUpdateExpirationTime(t *testing.T) {
	initTestDb(t)

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
	initTestDb(t)

	tok, ok := GetAccessToken(accessTokenAlice)
	if !ok {
		t.Error("AccessToken does not exist")
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

package data

import (
	"database/sql"
	"github.com/G-Node/gin-auth/util"
	"testing"
)

const (
	grantReqTokenAlice = "U7JIKKYI"
)

func TestListGrantRequests(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	initTestDb(t)

	requests := ListGrantRequests()
	if len(requests) != 2 {
		t.Error("Exactly two grant requests expected in list")
	}
}

func TestGetGrantRequest(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	initTestDb(t)

	req, err := GetGrantRequest(grantReqTokenAlice)
	if err != nil {
		t.Error(err)
	}
	if req.ScopeRequested[0] != "repo-read" {
		t.Errorf("First requested scope was expected to be 'repo-read'")
	}

	_, err = GetGrantRequest("doesNotExist")
	if err != nil {
		if err != sql.ErrNoRows {
			t.Error("Error must be sql.ErrNoRows")
		}
	} else {
		t.Error("Error expected")
	}
}

func TestCreateGrantRequest(t *testing.T) {
	initTestDb(t)

	token := util.RandomToken()
	state := util.RandomToken()
	code := util.RandomToken()
	new := GrantRequest{
		Token:           token,
		GrantType:       "code",
		State:           state,
		Code:            code,
		ScopeRequested:  SqlStringSlice{"foo-read", "foo-write", "foo-admin"},
		ScopeApproved:   SqlStringSlice{"foo-read"},
		OAuthClientUUID: uuidClientGin,
		AccountUUID:     uuidAlice}

	err := new.Create()
	if err != nil {
		t.Error(err)
	}

	check, err := GetGrantRequest(token)
	if err != nil {
		t.Error(err)
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

	req, err := GetGrantRequest(grantReqTokenAlice)
	if err != nil {
		t.Error(err)
	}

	req.Code = newCode
	req.State = newState

	err = req.Update()
	if err != nil {
		t.Error(err)
	}

	check, err := GetGrantRequest(grantReqTokenAlice)
	if err != nil {
		t.Error(err)
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

	req, err := GetGrantRequest(grantReqTokenAlice)
	if err != nil {
		t.Error(err)
	}

	err = req.Delete()
	if err != nil {
		t.Error(err)
	}

	_, err = GetGrantRequest(uuidClientGin)
	if err != nil {
		if err != sql.ErrNoRows {
			t.Error("Error must be sql.ErrNoRows")
		}
	} else {
		t.Error("Error expected")
	}
}

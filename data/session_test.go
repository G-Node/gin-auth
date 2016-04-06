package data

import (
	"github.com/G-Node/gin-auth/util"
	"testing"
	"time"
)

const (
	sessionTokenAlice = "DNM5RS3C"
	sessionTokenBob   = "2MFZZUKI"
)

func TestListSessions(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	sessions := ListSessions()
	if len(sessions) != 2 {
		t.Error("Exactly to sessions expected in slice")
	}
}

func TestGetSession(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	sess, ok := GetSession(sessionTokenAlice)
	if !ok {
		t.Error("Session does not exist")
	}
	if sess.AccountUUID != uuidAlice {
		t.Errorf("AccountUUID was expectd to be '%s'", uuidAlice)
	}

	_, ok = GetSession("doesNotExist")
	if ok {
		t.Error("Session should not exist")
	}
}

func TestClearOldSessions(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	deleted := ClearOldSessions()
	if deleted != 1 {
		t.Error("Exactly one session is supposed to be deleted")
	}

	_, ok := GetSession(sessionTokenBob)
	if ok {
		t.Error("Bobs session should not exist")
	}
}

func TestCreateSession(t *testing.T) {
	initTestDb(t)

	token := util.RandomToken()
	new := Session{
		Token:       token,
		Expires:     time.Now().Add(time.Hour * 12),
		AccountUUID: uuidAlice}

	err := new.Create()
	if err != nil {
		t.Error(err)
	}

	check, ok := GetSession(token)
	if !ok {
		t.Error("Token does not exist")
	}
	if check.AccountUUID != uuidAlice {
		t.Errorf("AccountUUID is supposed to be '%s'", uuidAlice)
	}
}

func TestSessionUpdateExpirationTime(t *testing.T) {
	initTestDb(t)

	sess, ok := GetSession(sessionTokenBob)
	if !ok {
		t.Error("Session does not exist")
	}
	if time.Since(sess.Expires) < 0 {
		t.Error("Sesssion should be expired")
	}

	sess.UpdateExpirationTime()
	if time.Since(sess.Expires) > 0 {
		t.Error("Sesssion should not be expired")
	}

	check, ok := GetSession(sessionTokenBob)
	if !ok {
		t.Error("Session does not exist")
	}
	if time.Since(check.Expires) > 0 {
		t.Error("Sesssion should not be expired")
	}
}

func TestSessionDelete(t *testing.T) {
	initTestDb(t)

	sess, ok := GetSession(sessionTokenAlice)
	if !ok {
		t.Error("Session does not exist")
	}

	err := sess.Delete()
	if err != nil {
		t.Error(err)
	}

	_, ok = GetSession(sessionTokenAlice)
	if ok {
		t.Error("Session should not exist")
	}
}

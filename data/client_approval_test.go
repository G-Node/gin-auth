package data

import (
	"github.com/pborman/uuid"
	"testing"
)

const (
	approvalUuidAlice = "31da7869-4593-4682-b9f2-5f47987aa5fc"
)

func TestListClientApprovals(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	approval := ListClientApprovals()
	if len(approval) != 1 {
		t.Error("Exactly to approval expected in slice")
	}
}

func TestGetClientApproval(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	app, ok := GetClientApproval(approvalUuidAlice)
	if !ok {
		t.Error("Client approval does not exist")
	}
	if app.AccountUUID != uuidAlice {
		t.Errorf("AccountUUID was expectd to be '%s'", uuidAlice)
	}

	_, ok = GetClientApproval("doesNotExist")
	if ok {
		t.Error("Client approval should not exist")
	}
}

func TestClientApprovalCreate(t *testing.T) {
	initTestDb(t)

	uuid := uuid.NewRandom().String()
	new := ClientApproval{
		UUID:        uuid,
		Scope:       SqlStringSlice{"foo-read", "foo-write"},
		ClientUUID:  uuidClientGin,
		AccountUUID: uuidBob}

	err := new.Create()
	if err != nil {
		t.Error(err)
	}

	check, ok := GetClientApproval(uuid)
	if !ok {
		t.Error("Approval does not exist")
	}
	if check.AccountUUID != uuidBob {
		t.Errorf("AccountUUID is supposed to be '%s'", uuidBob)
	}
	if check.Scope[1] != "foo-write" {
		t.Error("Second scope is supposed to be 'foo-write'")
	}
}

func TestClientApprovalUpdate(t *testing.T) {
	initTestDb(t)

	newScope := SqlStringSlice{"bar-read", "bar-write"}

	app, ok := GetClientApproval(approvalUuidAlice)
	if !ok {
		t.Error("Approval does not exist")
	}

	app.Scope = newScope

	err := app.Update()
	if err != nil {
		t.Error(err)
	}

	check, ok := GetClientApproval(approvalUuidAlice)
	if !ok {
		t.Error("Approval does not exist")
	}
	if check.Scope[0] != "bar-read" {
		t.Error("First scope expected to be 'bar-read'")
	}
	if check.Scope[1] != "bar-write" {
		t.Error("Second scope expected to be 'bar-write'")
	}
}

func TestClientApprovalDelete(t *testing.T) {
	initTestDb(t)

	app, ok := GetClientApproval(approvalUuidAlice)
	if !ok {
		t.Error("Approval does not exist")
	}

	err := app.Delete()
	if err != nil {
		t.Error(err)
	}

	_, ok = GetClientApproval(approvalUuidAlice)
	if ok {
		t.Error("Approval should not exist")
	}
}

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
	"github.com/pborman/uuid"
	"testing"
)

const (
	approvalUuidAlice = "31da7869-4593-4682-b9f2-5f47987aa5fc"
)

func TestListClientApprovals(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	approval := ListClientApprovals()
	if len(approval) != 2 {
		t.Error("Exactly to approval expected in slice")
	}
}

func TestGetClientApproval(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

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
	InitTestDb(t)

	uuid := uuid.NewRandom().String()
	fresh := ClientApproval{
		UUID:        uuid,
		Scope:       util.NewStringSet("foo-read", "foo-write"),
		ClientUUID:  uuidClientGin,
		AccountUUID: uuidBob}

	err := fresh.Create()
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
	if !check.Scope.Contains("foo-write") {
		t.Error("Scope should contain 'foo-write'")
	}
	if !check.Scope.Contains("foo-read") {
		t.Error("Scope should contain 'foo-read'")
	}
}

func TestClientApprovalUpdate(t *testing.T) {
	InitTestDb(t)

	newScope := util.NewStringSet("bar-read", "bar-write")

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
	if !check.Scope.Contains("bar-read") {
		t.Error("Scope should contain 'bar-read'")
	}
	if !check.Scope.Contains("bar-write") {
		t.Error("Scope should contain 'bar-write'")
	}
}

func TestClientApprovalDelete(t *testing.T) {
	InitTestDb(t)

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

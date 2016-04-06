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
)

const (
	keyPrintAlice = "SHA256:68a7N8YngrRrQF51SqLOONxILfaPa2A6ooW02Uiz+wM"
)

func TestListSSHKeys(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	keys := ListSSHKeys()
	if len(keys) != 2 {
		t.Error("Two SSH keys expected in list")
	}
}

func TestGetSSHKey(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	key, err := GetSSHKey(keyPrintAlice)
	if err != nil {
		t.Error(err)
	}
	if key.AccountUUID != uuidAlice {
		t.Errorf("AccountUUID was expected to be '%s'", uuidAlice)
	}

	_, err = GetSSHKey("doesNotExist")
	if err != nil {
		if err != sql.ErrNoRows {
			t.Error("Error must be sql.ErrNoRows")
		}
	} else {
		t.Error("Error expected")
	}
}

func TestCreateSSHKey(t *testing.T) {
	initTestDb(t)

	fingerprint := "SHA256:A3tkBXFQWkjU6rzhkofY55G7tPR/Lmna4B+WEGVFXOQ"
	new := &SSHKey{
		Fingerprint: fingerprint,
		Key:         "fake key",
		Description: "Alice 2nd key",
		AccountUUID: uuidAlice}

	err := new.Create()
	if err != nil {
		t.Error(err)
	}

	check, err := GetSSHKey(fingerprint)
	if err != nil {
		t.Error(err)
	}
	if check.AccountUUID != uuidAlice {
		t.Errorf("Login was expected to be $s", uuidAlice)
	}
}

func TestDeleteSSHKey(t *testing.T) {
	initTestDb(t)

	key, err := GetSSHKey(keyPrintAlice)
	if err != nil {
		t.Error(err)
	}

	err = key.Delete()
	if err != nil {
		t.Error(err)
	}

	_, err = GetSSHKey(keyPrintAlice)
	if err != nil {
		if err != sql.ErrNoRows {
			t.Error("Error must be sql.ErrNoRows")
		}
	} else {
		t.Error("Error expected")
	}
}

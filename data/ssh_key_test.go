// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"testing"

	"github.com/G-Node/gin-auth/util"
)

const (
	keyPrintAlice = "A3tkBXFQWkjU6rzhkofY55G7tPR/Lmna4B+WEGVFXOQ"
)

func TestListSSHKeys(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	keys := ListSSHKeys()
	if len(keys) != 3 {
		t.Error("Three SSH keys expected in list")
	}
}

func TestGetSSHKey(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	key, ok := GetSSHKey(keyPrintAlice)
	if !ok {
		t.Error("SSH key does not exist")
	}
	if key.AccountUUID != uuidAlice {
		t.Errorf("AccountUUID was expected to be '%s'", uuidAlice)
	}

	_, ok = GetSSHKey("doesNotExist")
	if ok {
		t.Error("SSH key should not exist")
	}
}

func TestCreateSSHKey(t *testing.T) {
	InitTestDb(t)

	// Test normal ssh key creation
	fingerprint := "SHA256:A3tkBXFQWkjU6rzhkofY55G7tPR/Lmna4B+WEGVFXOQ"
	fresh := &SSHKey{
		Fingerprint: fingerprint,
		Key:         "fake key",
		Description: "Alice 2nd key",
		AccountUUID: uuidAlice}

	err := fresh.Create()
	if err != nil {
		t.Errorf("Error creating ssh key: %s\n", err.Error())
	}

	check, ok := GetSSHKey(fingerprint)
	if !ok {
		t.Error("SSH key does not exist")
	}
	if check.AccountUUID != uuidAlice {
		t.Errorf("Login was expected to be '%s'", uuidAlice)
	}
	if check.IsTemporary {
		t.Error("Temporary key flag was expected to be false but was true")
	}

	// Test normal ssh key creation with temporary false flag
	fingerprint = "SHA256:A3tkBXFQWkjU6rzhkofY55G7tPR/Lmna4B+WEGVFXOc"
	fresh = &SSHKey{
		Fingerprint: fingerprint,
		Key:         "fake 3rd key",
		Description: "Alice 3rd key",
		AccountUUID: uuidAlice,
		IsTemporary: false}

	err = fresh.Create()
	if err != nil {
		t.Errorf("Error creating ssh key: %s\n", err.Error())
	}

	check, ok = GetSSHKey(fingerprint)
	if !ok {
		t.Error("SSH key does not exist")
	}
	if check.IsTemporary {
		t.Error("Temporary key flag was expected to be false but was true")
	}

	// Test temporary ssh key creation
	fingerprint = "SHA256:A3tkBXFQWkjU6rzhkofY55G7tPR/Lmna4B+WEGVFXOb"
	fresh = &SSHKey{
		Fingerprint: fingerprint,
		Key:         "fake temporary key",
		Description: "Temporary key",
		AccountUUID: uuidAlice,
		IsTemporary: true}

	err = fresh.Create()
	if err != nil {
		t.Errorf("Error creating temporary ssh key: %s\n", err.Error())
	}
	check, ok = GetSSHKey(fingerprint)
	if !ok {
		t.Error("Temporary ssh key does not exist")
	}
	if !check.IsTemporary {
		t.Error("Temporary key flag was expected to be true but was false")
	}
}

func TestDeleteSSHKey(t *testing.T) {
	InitTestDb(t)

	key, ok := GetSSHKey(keyPrintAlice)
	if !ok {
		t.Error("SSH key does not exist")
	}

	err := key.Delete()
	if err != nil {
		t.Error(err)
	}

	_, ok = GetSSHKey(keyPrintAlice)
	if ok {
		t.Error("SSH key should not exist")
	}
}

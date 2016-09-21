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

	fingerprint := "SHA256:A3tkBXFQWkjU6rzhkofY55G7tPR/Lmna4B+WEGVFXOQ"
	fresh := &SSHKey{
		Fingerprint: fingerprint,
		Key:         "fake key",
		Description: "Alice 2nd key",
		AccountUUID: uuidAlice}

	err := fresh.Create()
	if err != nil {
		t.Error(err)
	}

	check, ok := GetSSHKey(fingerprint)
	if !ok {
		t.Error("SSH key does not exist")
	}
	if check.AccountUUID != uuidAlice {
		t.Errorf("Login was expected to be '%s'", uuidAlice)
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

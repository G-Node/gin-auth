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
	if len(keys) != 6 {
		t.Errorf("Expected six SSH keys but got: %d.\n", len(keys))
	}
}

func TestGetSSHKey(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	const keyBobPermanent = "XDKYPWTM9ffhH+MvRs/zrNVP7eoYLf5YG8/1BJrZCJw"
	const keyBobTmpInvalid = "LTPF+bl45+47oT1X+Yxy0oNH4P6xufQhNxGMjRvxP2A"
	const keyBobTmpValid = "dgU2JX3eCYur5xbKhFQ+jEACSurCwtRaG+Qn6SYq7lE"

	// Test non existing key
	_, ok := GetSSHKey("doesNotExist")
	if ok {
		t.Error("SSH key should not exist.")
	}

	// Test permanent SSH key with createdat within TmpSshKeyLifeTime
	key, ok := GetSSHKey(keyPrintAlice)
	if !ok {
		t.Errorf("Permanent SSH key with fingerprint '%s' was not returned.\n", keyPrintAlice)
	}
	if key.AccountUUID != uuidAlice {
		t.Errorf("AccountUUID was expected to be '%s', but was '%s'.\n", uuidAlice, key.AccountUUID)
	}

	// Test permanent SSH key with createdat after TmpSshKeyLifeTime
	key, ok = GetSSHKey(keyBobPermanent)
	if !ok {
		t.Errorf("Permanent SSH key with fingerprint '%s' was not returned.\n", keyBobPermanent)
	}
	if key.AccountUUID != uuidBob {
		t.Errorf("AccountUUID was expected to be '%s', but was '%s'.\n", uuidBob, key.AccountUUID)
	}

	// Test temporary SSH key with createdat after TmpSshKeyLifeTime
	_, ok = GetSSHKey(keyBobTmpInvalid)
	if ok {
		t.Errorf("Did not expect invalid temporary SSH key with fingerprint '%s' to be returned.\n", keyBobTmpInvalid)
	}

	// Test temporary SSH key with createdat before TmpSshKeyLifeTime
	key, ok = GetSSHKey(keyBobTmpValid)
	if !ok {
		t.Errorf("Valid temporary SSH key with fingerprint '%s' was not returned.\n", keyBobTmpValid)
	}
	if key.AccountUUID != uuidBob {
		t.Errorf("AccountUUID was expected to be '%s', but was '%s'.\n", uuidBob, key.AccountUUID)
	}
}

func TestCreateSSHKey(t *testing.T) {
	InitTestDb(t)

	// Test normal ssh key creation
	fingerprint := "A3tkBXFQWkjU6rzhkofY55G7tPR/Lmna4B+WEGVFXOq"
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
	if check.Temporary {
		t.Error("Temporary key flag was expected to be false but was true")
	}

	// Test normal ssh key creation with temporary false flag
	fingerprint = "A3tkBXFQWkjU6rzhkofY55G7tPR/Lmna4B+WEGVFXOc"
	fresh = &SSHKey{
		Fingerprint: fingerprint,
		Key:         "fake 3rd key",
		Description: "Alice 3rd key",
		AccountUUID: uuidAlice,
		Temporary:   false}

	err = fresh.Create()
	if err != nil {
		t.Errorf("Error creating ssh key: %s\n", err.Error())
	}

	check, ok = GetSSHKey(fingerprint)
	if !ok {
		t.Error("SSH key does not exist")
	}
	if check.Temporary {
		t.Error("Temporary key flag was expected to be false but was true")
	}

	// Test temporary ssh key creation
	fingerprint = "A3tkBXFQWkjU6rzhkofY55G7tPR/Lmna4B+WEGVFXOb"
	fresh = &SSHKey{
		Fingerprint: fingerprint,
		Key:         "fake temporary key",
		Description: "Temporary key",
		AccountUUID: uuidAlice,
		Temporary:   true}

	err = fresh.Create()
	if err != nil {
		t.Errorf("Error creating temporary ssh key: %s\n", err.Error())
	}
	check, ok = GetSSHKey(fingerprint)
	if !ok {
		t.Error("Temporary ssh key does not exist")
	}
	if !check.Temporary {
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

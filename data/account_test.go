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
	uuidAlice = "bf431618-f696-4dca-a95d-882618ce4ef9"
)

func TestListAccounts(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	initTestDb(t)

	accounts := ListAccounts()
	if len(accounts) != 2 {
		t.Error("Two accounts expected in list")
	}
}

func TestGetAccount(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	initTestDb(t)

	alice, err := GetAccount(uuidAlice)
	if err != nil {
		t.Error(err)
	}
	if alice.Login != "alice" {
		t.Error("Login was expected to be 'alice'")
	}

	_, err = GetAccount("doesNotExist")
	if err != nil {
		if err != sql.ErrNoRows {
			t.Error("Error must be sql.ErrNoRows")
		}
	} else {
		t.Error("Error expected")
	}

}

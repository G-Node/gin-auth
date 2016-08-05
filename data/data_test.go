// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"testing"

	"github.com/G-Node/gin-auth/conf"
)

func TestEmailDispatch(t *testing.T) {
	InitTestDb(t)

	emails, err := GetQueuedEmails()
	if err != nil {
		t.Errorf("Error fetching queued e-mails: '%s'\n", err.Error())
	}
	if len(emails) < 3 {
		t.Error("Expected queued e-mails, but result did not match")
	}

	// There should be three database entries, one for each smtp modes.
	// Print and skip should result in the entries being deleted, the
	// empty mode should result in a bad username error and
	// should therefore not be deleted.
	username := conf.GetSmtpCredentials().Username
	conf.GetSmtpCredentials().Username = "iDoNotExist"
	EmailDispatch()
	conf.GetSmtpCredentials().Username = username

	emails, err = GetQueuedEmails()
	if err != nil {
		t.Errorf("Error fetching queued e-mails: '%s'\n", err.Error())
	}
	if len(emails) != 1 {
		t.Error("Number of db entries do not match expected result")
	}
}
